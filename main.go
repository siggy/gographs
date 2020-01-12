package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/siggy/gographs/cache"
	"github.com/siggy/gographs/web"
)

func main() {
	log.SetLevel(log.DebugLevel)

	repoCache, err := cache.New()
	if err != nil {
		log.Error(err)
		return
	}

	w := web.New(repoCache)
	err = w.Listen()
	if err != nil {
		log.Error(err)
		return
	}
}
