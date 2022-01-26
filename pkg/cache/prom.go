package cache

import (
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func registerGauges(client *redis.Client) {
	registerHashGauge(client, dotHash)
	registerHashGauge(client, svgHash)
	registerSetGauge(client, repoScores)
}

func registerHashGauge(client *redis.Client, key string) {
	registerGauge(
		func() float64 {
			size, _ := client.HLen(client.Context(), key).Result()
			return float64(size)
		},
		key,
	)
}

func registerSetGauge(client *redis.Client, key string) {
	registerGauge(
		func() float64 {
			size, _ := client.ZCount(client.Context(), key, "-inf", "+inf").Result()
			return float64(size)
		},
		key,
	)
}

func registerGauge(function func() float64, key string) {
	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace:   "gographs",
			Subsystem:   "cache",
			Name:        "size",
			Help:        "Size of the cache",
			ConstLabels: prometheus.Labels{"key": key},
		},
		function,
	)
}
