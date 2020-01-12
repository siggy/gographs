package web

// curl https://proxy.golang.org/github.com/siggy/heypic/@v/master.info
// curl -O https://proxy.golang.org/github.com/siggy/heypic/@v/v0.0.0-20180506171301-e384f182b391.zip | unzip
// goda graph github.com/linkerd/linkerd2...:root | dot -Tsvg -o graph2.svg

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/siggy/gographs/cache"
	"github.com/siggy/gographs/repo"
)

type Web struct {
	router *mux.Router
	cache  *cache.Cache
}

func New(c *cache.Cache) *Web {

	w := &Web{
		router: mux.NewRouter(),
		cache:  c,
	}

	w.router.PathPrefix("/repo").Queries("cluster", "{cluster:true|false}").HandlerFunc(w.repoHandler)
	w.router.PathPrefix("/repo").HandlerFunc(w.repoHandler)
	w.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	return w
}

func (w *Web) Listen() error {
	host := "localhost:8888"
	log.Infof("serving on %s", host)
	err := http.ListenAndServe(host, w.router)
	if err != nil {
		log.Errorf("Failed to ListenAndServe on %s: %s", host, err)
		return err
	}

	return nil
}

func (w *Web) repoHandler(rw http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	svg, err := w.cache.GetURLToSVG(r.URL.String())
	if err != nil {
		route := mux.CurrentRoute(r)
		cluster := mux.Vars(r)["cluster"] == "true"

		tpl, err := route.GetPathTemplate()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		goRepo := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, tpl+"/"), ".svg")
		svg, err = repo.GenSVG(goRepo, cluster)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.cache.SetURLToSVG(r.URL.String(), svg)
	}

	rw.Header().Set("Content-Type", "image/svg+xml")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte(svg))
}
