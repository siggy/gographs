package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/siggy/gographs/cache"
	"github.com/siggy/gographs/web"
)

func main() {
	log.SetLevel(log.DebugLevel)

	c, err := cache.New()
	if err != nil {
		log.Error(err)
		return
	}

	err = web.Start(c)
	if err != nil {
		log.Error(err)
		return
	}
}
