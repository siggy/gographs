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
	router := mux.NewRouter().Methods(http.MethodGet).Subrouter()

	log := log.WithFields(
		log.Fields{
			"web": addr,
		},
	)

	repoHandler := mkRepoHandler(c, log)
	router.PathPrefix("/repo").HandlerFunc(repoHandler)
	router.HandleFunc("/top-repos", mkTopReposHandler(c))
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	log.Infof("serving on %s", addr)

	return http.ListenAndServe(addr, router)
}

func mkRepoHandler(cache *cache.Cache, log *log.Entry) http.HandlerFunc {

	// /repo/github.com/siggy/gographs.svg?cluster=false&refresh=false
	return func(rw http.ResponseWriter, r *http.Request) {
		vars := r.URL.Query()
		cluster := vars.Get("cluster") == "true"
		refresh := vars.Get("refresh") == "true"

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
			err := cache.Clear(goRepo)
			if err != nil {
				log.Errorf("Failed to clear cache for repo %s: %s", goRepo, err)
			}
		}

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
