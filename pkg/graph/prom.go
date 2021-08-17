package graph

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// TODO: move to prom package?

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "gographs",
		Subsystem: "graphclient",
		Name:      "requests_total",
		Help:      "Count of HTTP requests.",
	}, []string{"repo", "cluster"})

	httpErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "gographs",
		Subsystem: "graphclient",
		Name:      "errors_total",
		Help:      "Count of HTTP errors.",
	}, []string{"repo", "cluster", "error"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "gographs",
		Subsystem: "graphclient",
		Name:      "duration_seconds",
		Help:      "Duration of HTTP requests.",
		Buckets:   prometheus.ExponentialBuckets(0.001, 1.3, 50),
	}, []string{"repo", "cluster"})
)
