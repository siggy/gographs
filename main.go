package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/siggy/gographs/pkg/cache"
	"github.com/siggy/gographs/pkg/graph"
	"github.com/siggy/gographs/pkg/web"
	log "github.com/sirupsen/logrus"
)

const (
	targetAll   = "all"
	targetWeb   = "web"
	targetGraph = "graph"
)

func main() {
	target := flag.String("target", targetAll, fmt.Sprintf("program target, must be one of: %s, %s, %s", targetAll, targetWeb, targetGraph))
	webAddr := flag.String("addr", "localhost:8888", "web address to listen on")
	graphAddr := flag.String("graph-addr", graph.DefaultGraphAddr, "graph address to listen on")
	logLevel := flag.String("log-level", log.DebugLevel.String(), "log level, must be one of: panic, fatal, error, warn, info, debug, trace")
	metricsAddr := flag.String("metrics-addr", "localhost:8080", "address to listen on for metrics requests")
	redisAddr := flag.String("redis-addr", "localhost:6379", "address to connect to redis")
	flag.Parse()

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("invalid log-level: %s", *logLevel)
	}
	log.SetLevel(level)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Infof("metrics server listening on %s", *metricsAddr)
		err = http.ListenAndServe(*metricsAddr, nil)
		if err != nil {
			log.Fatalf("failed to listen on metrics address [%s]: %s", *metricsAddr, err)
		}
	}()

	if *target == targetAll || *target == targetGraph {
		go func() {
			err := graph.Start(*graphAddr)
			if err != nil {
				log.Fatalf("failed to start graph server [%s]: %s", *webAddr, err)
			}
		}()
	}

	if *target == targetAll || *target == targetWeb {
		go func() {
			c, err := cache.New(*redisAddr)
			if err != nil {
				log.Fatalf("failed to initialize cache: %s", err)
			}

			graph := graph.NewClient(*graphAddr)
			err = web.Start(c, *webAddr, graph)
			if err != nil {
				log.Fatalf("failed to start web server [%s]: %s", *webAddr, err)
			}
		}()
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
