package graph

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// TODO: move to prom package?

const (
	gographsNamespace    = "gographs"
	graphclientSubsystem = "graphclient"
	repoLabel            = "repo"
	clusterLabel         = "cluster"
)

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: gographsNamespace,
		Subsystem: graphclientSubsystem,
		Name:      "requests_total",
		Help:      "Count of HTTP requests.",
	}, []string{repoLabel, clusterLabel})

	httpErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: gographsNamespace,
		Subsystem: graphclientSubsystem,
		Name:      "errors_total",
		Help:      "Count of HTTP errors.",
	}, []string{repoLabel, clusterLabel, "error"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: gographsNamespace,
		Subsystem: graphclientSubsystem,
		Name:      "duration_seconds",
		Help:      "Duration of HTTP requests.",
		Buckets:   prometheus.ExponentialBuckets(0.001, 1.3, 50),
	}, []string{repoLabel, clusterLabel})
)
