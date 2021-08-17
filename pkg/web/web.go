package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/siggy/gographs/pkg/cache"
	"github.com/siggy/gographs/pkg/graph"
	"github.com/siggy/gographs/pkg/prom"
	"github.com/siggy/gographs/pkg/render"
	log "github.com/sirupsen/logrus"
)

const webServer = "web"

// Start initializes the web server and starts listening.
func Start(c *cache.Cache, addr string, graph *graph.Client) error {
	router := mux.NewRouter()
	router.Use(prom.Middleware(webServer))

	getRouter := router.Methods(http.MethodGet).Subrouter()
	postRouter := router.Methods(http.MethodPost).Subrouter()

	log := log.WithFields(
		log.Fields{
			webServer: addr,
		},
	)

	// web views
	getRouter.PathPrefix("/repo").HandlerFunc(repoHandler)
	getRouter.HandleFunc("/svg", repoHandler)
	getRouter.HandleFunc("/", repoHandler)

	// apis
	graphHandler := mkGraphHandler(graph, c, log)
	getRouter.PathPrefix("/graph").HandlerFunc(graphHandler)
	postRouter.PathPrefix("/graph").HandlerFunc(graphHandler)
	getRouter.HandleFunc("/top-repos", mkTopReposHandler(c))

	// assets
	getRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	log.Infof("%s server listening on %s", webServer, addr)

	return http.ListenAndServe(addr, router)
}

func repoHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/index.html")
}

func mkGraphHandler(graph *graph.Client, cache *cache.Cache, log *log.Entry) http.HandlerFunc {
	// GET  /graph/github.com/siggy/gographs.svg
	// POST /graph/github.com/siggy/gographs.svg (for refresh)
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		cluster := vars.Get("cluster") == "true"

		refresh := r.Method == http.MethodPost

		tpl, err := mux.CurrentRoute(r).GetPathTemplate()
		if err != nil {
			writeError(rw, r, http.StatusInternalServerError, err.Error(), err)
			return
		}

		suffix := ""
		contentType := ""
		if strings.HasSuffix(r.URL.Path, ".svg") {
			suffix = ".svg"
			contentType = "image/svg+xml; charset=utf-8"
		} else if strings.HasSuffix(r.URL.Path, ".dot") {
			suffix = ".dot"
			contentType = "text/plain; charset=utf-8"
		} else {
			writeError(rw, r, http.StatusBadRequest, "svg or dot suffix required", nil)
			return
		}

		goRepo := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, tpl+"/"), suffix)

		if refresh {
			log.Debugf("Clearing cache for %s", goRepo)
			err = cache.Clear(goRepo)
			if err != nil {
				log.Errorf("Failed to clear cache for repo %s: %s", goRepo, err)
			}
		}

		log.Debugf("Processing %s", goRepo)

		out := ""
		if suffix == ".svg" {
			out, err = render.ToSVG(graph, cache, goRepo, cluster)
		} else if suffix == ".dot" {
			out, err = render.ToDOT(graph, cache, goRepo, cluster)
		}
		if err != nil {
			message := fmt.Sprintf("Failed to render %s to %s", goRepo, suffix)
			writeError(rw, r, http.StatusInternalServerError, message, err)
			return
		}

		go cache.RepoScoreIncr(goRepo)

		rw.Header().Set("Content-Type", contentType)
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(out))
	}
}

// TODO: poll for this every interval, hold result in local mem
func mkTopReposHandler(cache *cache.Cache) http.HandlerFunc {
	// /top-repos
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(rw, r, http.StatusMethodNotAllowed, "", nil)
			return
		}

		scores, err := cache.RepoScores()
		if err != nil {
			writeError(rw, r, http.StatusInternalServerError, "", err)
			return
		}

		j, err := json.Marshal(scores)
		if err != nil {
			writeError(rw, r, http.StatusInternalServerError, "", err)
			return
		}

		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write(j)
	}
}

// writeError handles all errors returned by the web server. It writes an error
// header, an optional error message, counts the error in metrics, and logs it.
func writeError(rw http.ResponseWriter, r *http.Request, status int, message string, err error) {
	rw.WriteHeader(status)
	if message != "" {
		rw.Write([]byte(message))
	}

	route := mux.CurrentRoute(r)
	path, _ := route.GetPathTemplate()

	log.Errorf("Failed request for [%s]: [%d] Message: [%s] Error: [%s]", path, status, message, err)
	prom.CountError(webServer, r, status, message, err)
}
