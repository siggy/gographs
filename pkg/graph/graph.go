package graph

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/siggy/gographs/pkg/prom"
	log "github.com/sirupsen/logrus"
)

const graphServer = "graph"

// Start initializes the graph server and starts listening.
func Start(addr string) error {
	router := mux.NewRouter()
	router.Use(prom.Middleware(graphServer))

	log := log.WithFields(
		log.Fields{
			graphServer: addr,
		},
	)

	// apis
	graphHandler := mkGraphHandler(log)
	router.HandleFunc("/graph", graphHandler).Methods(http.MethodPost)

	log.Infof("%s server listening on %s", graphServer, addr)

	return http.ListenAndServe(addr, router)
}

func mkGraphHandler(log *log.Entry) http.HandlerFunc {
	// curl --data '{"repo":"github.com/siggy/gographs","cluster":true}' -X POST /graph
	return func(rw http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var p Post
		err := decoder.Decode(&p)
		if err != nil {
			message := fmt.Sprintf("Failed to decode POST body %s", p.Repo)
			writeError(rw, r, http.StatusInternalServerError, message, err)
			return
		}

		log.Debugf("Processing %s", p.Repo)

		dot, err := repoToDot(p.Repo, p.Cluster)
		if err != nil {
			message := fmt.Sprintf("Failed to render dot: %s", p.Repo)
			writeError(rw, r, http.StatusInternalServerError, message, err)
			return
		}

		rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(dot))
	}
}

// writeError handles all errors returned by the web server. It writes an error
// header, an optional error message, counts the error in metrics, and logs it.
// TODO: factor out with web.go
func writeError(rw http.ResponseWriter, r *http.Request, status int, message string, err error) {
	rw.WriteHeader(status)
	if message != "" {
		rw.Write([]byte(message))
	}

	route := mux.CurrentRoute(r)
	path, _ := route.GetPathTemplate()

	log.Errorf("Failed request for [%s]: [%d] Message: [%s] Error: [%s]", path, status, message, err)
	prom.CountError(graphServer, r, status, message, err)
}
