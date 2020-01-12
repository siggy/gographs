package main

// curl https://proxy.golang.org/github.com/siggy/heypic/@v/master.info
// curl -O https://proxy.golang.org/github.com/siggy/heypic/@v/v0.0.0-20180506171301-e384f182b391.zip | unzip
// goda graph github.com/linkerd/linkerd2...:root | dot -Tsvg -o graph2.svg

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/siggy/gographs/cache"
	"github.com/siggy/gographs/web"
)

type revInfo struct {
	Version string    `json:"Version"`
	Time    time.Time `json:"Time"`
}

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stderr)

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
