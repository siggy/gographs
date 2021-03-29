package main

import (
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/siggy/gographs/pkg/cache"
	"github.com/siggy/gographs/pkg/web"
	log "github.com/sirupsen/logrus"
)

func main() {
	addr := flag.String("addr", "localhost:8888", "address to listen on")
	logLevel := flag.String("log-level", log.DebugLevel.String(), "log level, must be one of: panic, fatal, error, warn, info, debug, trace")
	metricsAddr := flag.String("metrics-addr", "localhost:8080", "address to listen on for metrics requests")
	redisAddr := flag.String("redis-addr", "localhost:6379", "address to connect to redis")
	flag.Parse()

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("invalid log-level: %s", *logLevel)
	}
	log.SetLevel(level)

	c, err := cache.New(*redisAddr)
	if err != nil {
		log.Fatalf("failed to initialize cache: %s", err)
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Infof("Listening for metrics on %s", *metricsAddr)
		err = http.ListenAndServe(*metricsAddr, nil)
		if err != nil {
			log.Fatalf("failed to listen on metrics address [%s]: %s", *metricsAddr, err)
		}
	}()

	err = web.Start(c, *addr)
	if err != nil {
		log.Fatalf("failed to start web server [%s]: %s", *addr, err)
	}
}
