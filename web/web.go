package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/siggy/gographs/cache"
	"github.com/siggy/gographs/repo"
)

// Start initializes the web server and starts listening
func Start(c *cache.Cache, addr string) error {
	router := mux.NewRouter()
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
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
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
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("svg or dot suffix required"))
			return
		}

		goRepo := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, tpl+"/"), suffix)

		if refresh {
			log.Debugf("Clearing cache for %s", goRepo)
			err := cache.Clear(goRepo)
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
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
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
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		scores, err := cache.RepoScores()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		j, err := json.Marshal(scores)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(j))
	}
}
