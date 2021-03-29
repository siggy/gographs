package web

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "gographs",
		Subsystem: "web",
		Name:      "requests_total",
		Help:      "Count of HTTP requests.",
	}, []string{"method", "path"})

	httpErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "gographs",
		Subsystem: "web",
		Name:      "errors_total",
		Help:      "Count of HTTP errors.",
	}, []string{"method", "path", "status", "message", "error"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "gographs",
		Subsystem: "web",
		Name:      "duration_seconds",
		Help:      "Duration of HTTP requests.",
		Buckets:   prometheus.ExponentialBuckets(0.001, 1.3, 50),
	}, []string{"method", "path"})
)

// prometheusMiddleware implements mux.MiddlewareFunc.
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		httpRequests.WithLabelValues(r.Method, path).Inc()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(r.Method, path))
		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}

func countError(r *http.Request, status int, message string, err error) {
	route := mux.CurrentRoute(r)
	path, _ := route.GetPathTemplate()
	errStr := ""
	if err != nil {
		errStr = err.Error()
	}
	httpErrors.WithLabelValues(r.Method, path, http.StatusText(status), message, errStr).Inc()
}
