package cache

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/valkey-io/valkey-go"
)

func registerGauges(client valkey.Client) {
	registerHashGauge(client, dotHash)
	registerHashGauge(client, svgHash)
	registerSetGauge(client, repoScores)
}

func registerHashGauge(client valkey.Client, key string) {
	registerGauge(
		func() float64 {
			size, _ := client.Do(
				context.Background(),
				client.B().Hlen().Key(key).Build(),
			).AsInt64()
			return float64(size)
		},
		key,
	)
}

func registerSetGauge(client valkey.Client, key string) {
	registerGauge(
		func() float64 {
			size, _ := client.Do(
				context.Background(),
				client.B().Zcount().Key(key).Min("-inf").Max("+inf").Build(),
			).AsInt64()
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
