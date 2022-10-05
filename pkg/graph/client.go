package graph

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

// Post defines the input POST body to the `/graph` endpoint.
// curl --data '{"repo":"github.com/siggy/gographs","cluster":true}' -X POST [graph-addr]/graph
type Post struct {
	Repo    string `json:"repo"`
	Cluster bool   `json:"cluster"`
}

// Client provides a client to the graph server.
type Client struct {
	url string
	log *log.Entry
}

// DefaultGraphAddr defines the graph server's address when running locally. If
// the caller uses something other than the default it is assumed to be TLS'd.
const DefaultGraphAddr = "localhost:8889"

// NewClient creates a client to the graph server.
func NewClient(addr string) *Client {
	url := fmt.Sprintf("http://%s/graph", addr)
	if addr != DefaultGraphAddr {
		url = fmt.Sprintf("https://%s/graph", addr)
	}

	log := log.WithFields(
		log.Fields{
			"graphclient": url,
		},
	)

	log.Infof("Graph client initialized")

	return &Client{url, log}
}

// Get takes a repo and cluster flag and returns a DOT representation of the
// repo.
func (c *Client) Get(repo string, cluster bool) (string, error) {
	body, err := json.Marshal(
		Post{
			Repo:    repo,
			Cluster: cluster,
		},
	)
	if err != nil {
		return "", err
	}

	c.log.Debugf("POST Request: %s", string(body))

	labels := prometheus.Labels{"repo": repo, "cluster": strconv.FormatBool(cluster)}
	httpRequests.With(labels).Inc()
	httpErrors, err := httpErrors.CurryWith(labels)
	if err != nil {
		return "", err
	}

	timer := prometheus.NewTimer(httpDuration.With(labels))
	defer timer.ObserveDuration()

	resp, err := http.Post(c.url, "text/plain; charset=utf-8", bytes.NewBuffer(body))
	if err != nil {
		httpErrors.WithLabelValues(err.Error()).Inc()
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		httpErrors.WithLabelValues(err.Error()).Inc()
		return "", err
	}

	debugStr := fmt.Sprintf("POST response[%d] (%d bytes): %s ", resp.StatusCode, len(respBody), string(respBody))
	c.log.Debug(debugStr)

	if resp.StatusCode != http.StatusOK {
		err := errors.New(debugStr)
		httpErrors.WithLabelValues(err.Error()).Inc()
		return "", err
	}

	return string(respBody), nil
}
