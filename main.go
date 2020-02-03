package main

import (
	"flag"

	log "github.com/sirupsen/logrus"

	"github.com/siggy/gographs/cache"
	"github.com/siggy/gographs/web"
)

func main() {
	addr := flag.String("addr", "localhost:8888", "address to listen on")
	logLevel := flag.String("log-level", log.DebugLevel.String(), "log level, must be one of: panic, fatal, error, warn, info, debug, trace")
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

	err = web.Start(c, *addr)
	if err != nil {
		log.Fatalf("failed to start web server: %s", err)
	}
}
