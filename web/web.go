package web

// curl https://proxy.golang.org/github.com/siggy/heypic/@v/master.info
// curl -O https://proxy.golang.org/github.com/siggy/heypic/@v/v0.0.0-20180506171301-e384f182b391.zip | unzip
// goda graph github.com/linkerd/linkerd2...:root | dot -Tsvg -o graph2.svg

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/siggy/gographs/cache"
	"github.com/siggy/gographs/repo"
)

func Start(c *cache.Cache) error {
	router := mux.NewRouter()

	repoHandler := mkRepoHandler(c)
	router.PathPrefix("/repo").Queries("cluster", "{cluster:true|false}").HandlerFunc(repoHandler)
	router.PathPrefix("/repo").HandlerFunc(repoHandler)

	topReposHandler := mkTopReposHandler(c)
	router.HandleFunc("/top-repos", topReposHandler)

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	host := "localhost:8888"
	log.Infof("serving on %s", host)

	return http.ListenAndServe(host, router)
}

func mkRepoHandler(cache *cache.Cache) http.HandlerFunc {

	// /repo/github.com/siggy/gographs.svg?cluster=true
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		suffix := ""
		contentType := ""
		if strings.HasSuffix(r.URL.Path, ".svg") {
			suffix = ".svg"
			contentType = "image/svg+xml;charset=utf-8"
		} else if strings.HasSuffix(r.URL.Path, ".dot") {
			suffix = ".dot"
			contentType = "text/plain;charset=utf-8"
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}

		route := mux.CurrentRoute(r)
		tpl, err := route.GetPathTemplate()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		goRepo := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, tpl+"/"), suffix)

		out, err := cache.GetURL(r.URL.String())
		if err != nil {

			// cache miss
			cluster := mux.Vars(r)["cluster"] == "true"
			if suffix == ".svg" {
				out, err = repo.GenSVG(cache, goRepo, cluster)
			} else if suffix == ".dot" {
				out, err = repo.GenDOT(cache, goRepo, cluster)
			}
			if err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				return
			}

			go cache.SetURL(r.URL.String(), out)
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

		rw.Header().Set("Content-Type", "application/json;charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(j))
	}
}
