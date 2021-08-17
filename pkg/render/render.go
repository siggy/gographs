package render

// This package takes GoLang repos as input and outputs SVG and DOT files:
//
// 1. repo => dot
//    curl --data '{"repo":"github.com/siggy/gographs","cluster":true}' -X POST [graph-addr]/graph
// 2. dot => svg
//    echo "..." | \
//      dot -Tsvg \
//      -Gfontname=Roboto,Arial,sans-serif \
//      -Nfontname=Roboto,Arial,sans-serif \
//      -Efontname=Roboto,Arial,sans-serif \
//      -o graph2.svg
//
// Nested control-flow accommodates caching:
//
// ToSVG(repo) {
//   ToDOT(repo) {} => DOT
//   dotToSVG(DOT) {} => SVG
// } => SVG

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/siggy/gographs/pkg/cache"
	"github.com/siggy/gographs/pkg/graph"
	log "github.com/sirupsen/logrus"
)

// ToSVG takes a GoLang repo as input and returns an SVG dependency graph
func ToSVG(graph *graph.Client, cache *cache.Cache, repo string, cluster bool) (string, error) {
	svg, err := cache.GetSVG(repo, cluster)
	if err == nil {
		return svg, nil
	}

	dot, err := ToDOT(graph, cache, repo, cluster)
	if err != nil {
		log.Errorf("error generating dot: %s", err)
		return "", err
	}

	svg, err = dotToSVG(dot)
	if err != nil {
		log.Errorf("error converting dot to svg: %s", err)
		return "", err
	}

	go cache.SetSVG(repo, cluster, svg)

	return svg, nil
}

// ToDOT takes a GoLang repo as input and returns a DOT dependency graph
func ToDOT(graph *graph.Client, cache *cache.Cache, repo string, cluster bool) (string, error) {
	dot, err := cache.GetDOT(repo, cluster)
	if err == nil {
		return dot, nil
	}

	dot, err = graph.Get(repo, cluster)
	if err != nil {
		return "", err
	}

	go cache.SetDOT(repo, cluster, dot)

	return dot, nil
}

func dotToSVG(dot string) (string, error) {
	command := exec.Command(
		"dot",
		"-Tsvg",
		"-Gfontname=Roboto,Arial,sans-serif",
		"-Nfontname=Roboto,Arial,sans-serif",
		"-Efontname=Roboto,Arial,sans-serif",
	)
	command.Stdin = strings.NewReader(dot)
	var stderr bytes.Buffer
	command.Stderr = &stderr

	log.Debugf("running dot: %s", command)
	svg, err := command.Output()
	if err != nil {
		log.Errorf("dot cmd failed: %s", err)
		return "", err
	}

	return string(svg), nil
}
