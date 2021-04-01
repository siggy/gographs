package repo

//
// Based on https://github.com/gojp/goreportcard, specifically:
// https://github.com/gojp/goreportcard/blob/6ecdf3c5c38cf0855cec02ab2a02ecb78b6e456f/download/download.go
//

// This package takes GoLang repos as input and outputs SVG and DOT files:
//
// 1. repo => dir
//    git clone --depth 1 https://github.com/siggy/gographs /repos/https://github.com/siggy/gographs"
// 2. dir => dot
//    goda graph -short -cluster github.com/siggy/gographs...:root
// 3. dot => svg
//    echo "..." | dot -Tsvg -o graph2.svg
//
// Nested control-flow accommodates caching at all levels:
//
// ToSVG(repo) {
//   ToDOT(repo) {
//     toDir(repo) {} => Dir
//     dirToDot(Dir) {} => DOT
//   } => DOT
//   dotToSVG(DOT) {} => SVG
// } => SVG

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/siggy/gographs/pkg/cache"
	log "github.com/sirupsen/logrus"
	"golang.org/x/tools/go/vcs"
)

// ToSVG takes a GoLang repo as input and returns an SVG dependency graph
func ToSVG(cache *cache.Cache, repo string, cluster bool) (string, error) {
	svg, err := cache.GetSVG(repo, cluster)
	if err == nil {
		return svg, nil
	}

	dot, err := ToDOT(cache, repo, cluster)
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
func ToDOT(cache *cache.Cache, repo string, cluster bool) (string, error) {
	dot, err := cache.GetDOT(repo, cluster)
	if err == nil {
		return dot, nil
	}

	codeDir, err := toDir(cache, repo)
	if err != nil {
		log.Errorf("failed to get dir: %s", err)
		return "", err
	}

	dot, err = dirToDot(codeDir, cluster)
	if err != nil {
		// goda command failed, delete repo dir
		go cache.DelRepoDir(repo)

		log.Errorf("goda failed: %s", err)
		return "", err
	}

	go cache.SetDOT(repo, cluster, dot)

	return dot, nil
}

func Exists(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}

func toDir(cache *cache.Cache, repo string) (string, error) {
	codeDir, err := cache.GetRepoDir(repo)
	if err == nil && Exists(codeDir) {
		// repo already present
		return codeDir, nil
	}

	codeDir, err = ioutil.TempDir("", "")
	if err != nil {
		log.Errorf("TempDir err: %s", err)
		return "", err
	}
	log.Debugf("writing to tempDir: %s", codeDir)
	err = os.MkdirAll(codeDir, os.ModePerm)
	if err != nil {
		log.Errorf("MkdirAll err: %s", err)
		return "", err
	}

	vcs.ShowCmd = true
	root, err := vcs.RepoRootForImportPath(trimScheme(repo), true)
	if err != nil {
		log.Errorf("RepoRootForImportPath err: %s", err)
		return "", err
	}

	root.VCS.CreateCmd = "clone --depth 1 --no-tags {repo} {dir}"
	err = root.VCS.Create(codeDir, root.Repo)
	if err != nil {
		log.Errorf("cmd.Create err: %s", err)
		return "", err
	}

	go cache.SetRepoDir(repo, codeDir)

	return codeDir, nil
}

func dirToDot(dir string, cluster bool) (string, error) {
	args := []string{"graph", "-short"}
	if cluster {
		args = append(args, "-cluster")
	}
	args = append(args, "./...:root")

	cmd := exec.Command("goda", args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Debugf("running goda: %s", cmd)
	err := cmd.Run()
	if err != nil {
		log.Errorf("goda cmd failed: %s", err)
		return "", err
	}

	serr := stderr.String()
	if strings.Contains(serr, "matched no packages") {
		err := fmt.Errorf("goda cmd returned stderr: %s", serr)
		log.Error(err)
		return "", err
	}

	return stdout.String(), nil
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

func trimScheme(repo string) string {
	schemeSep := "://"
	schemeSepIdx := strings.Index(repo, schemeSep)
	if schemeSepIdx > -1 {
		return repo[schemeSepIdx+len(schemeSep):]
	}

	return repo
}
