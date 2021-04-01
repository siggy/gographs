package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/siggy/gographs/pkg/cache"
	"github.com/siggy/gographs/pkg/repo"
	log "github.com/sirupsen/logrus"
)

// Start initializes the web server and starts listening.
func Start(c *cache.Cache, addr string) error {
	router := mux.NewRouter()
	router.Use(prometheusMiddleware)

	getRouter := router.Methods(http.MethodGet).Subrouter()
	postRouter := router.Methods(http.MethodPost).Subrouter()

	log := log.WithFields(
		log.Fields{
			"web": addr,
		},
	)

	// web views
	getRouter.PathPrefix("/repo").HandlerFunc(repoHandler)
	getRouter.HandleFunc("/svg", repoHandler)
	getRouter.HandleFunc("/", repoHandler)

	// apis
	graphHandler := mkGraphHandler(c, log)
	getRouter.PathPrefix("/graph").HandlerFunc(graphHandler)
	postRouter.PathPrefix("/graph").HandlerFunc(graphHandler)
	getRouter.HandleFunc("/top-repos", mkTopReposHandler(c))

	// assets
	getRouter.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	log.Infof("Web server listening on %s", addr)

	return http.ListenAndServe(addr, router)
}

func repoHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/index.html")
}

func mkGraphHandler(cache *cache.Cache, log *log.Entry) http.HandlerFunc {
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
			// get the repo's current location and delete
			codeDir, err := cache.GetRepoDir(goRepo)
			if err == nil && repo.Exists(codeDir) {
				go func(codeDir string) {
					err := os.RemoveAll(codeDir)
					if err != nil {
						log.Errorf("Failed to remove %s: %s", codeDir, err)
					}
				}(codeDir)
			}

			log.Debugf("Clearing cache for %s", goRepo)
			err = cache.Clear(goRepo)
			if err != nil {
				log.Errorf("Failed to clear cache for repo %s: %s", goRepo, err)
			}
		}

		log.Debugf("Processing %s", goRepo)

		out := ""
		if suffix == ".svg" {
			out, err = repo.ToSVG(cache, goRepo, cluster)
		} else if suffix == ".dot" {
			out, err = repo.ToDOT(cache, goRepo, cluster)
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
	countError(r, status, message, err)
}
