package prom

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "gographs",
		Subsystem: "server",
		Name:      "requests_total",
		Help:      "Count of HTTP requests.",
	}, []string{"server", "method", "path"})

	httpErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "gographs",
		Subsystem: "server",
		Name:      "errors_total",
		Help:      "Count of HTTP errors.",
	}, []string{"server", "method", "path", "status", "message", "error"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "gographs",
		Subsystem: "server",
		Name:      "duration_seconds",
		Help:      "Duration of HTTP requests.",
		Buckets:   prometheus.ExponentialBuckets(0.001, 1.3, 50),
	}, []string{"server", "method", "path"})
)

// Middleware returns a function that implements mux.MiddlewareFunc.
func Middleware(server string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			route := mux.CurrentRoute(r)
			path, _ := route.GetPathTemplate()
			httpRequests.WithLabelValues(server, r.Method, path).Inc()
			timer := prometheus.NewTimer(httpDuration.WithLabelValues(server, r.Method, path))
			next.ServeHTTP(w, r)
			timer.ObserveDuration()
		})
	}
}

// CountError increments http error counters.
func CountError(server string, r *http.Request, status int, message string, err error) {
	route := mux.CurrentRoute(r)
	path, _ := route.GetPathTemplate()
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	httpErrors.WithLabelValues(server, r.Method, path, http.StatusText(status), message, errStr).Inc()
}
