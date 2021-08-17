package graph

//
// Based on https://github.com/gojp/goreportcard, specifically:
// https://github.com/gojp/goreportcard/blob/6ecdf3c5c38cf0855cec02ab2a02ecb78b6e456f/download/download.go
//

// This file takes GoLang repos as input and outputs DOT files:
//
// 1. repo => dir
//    git clone --depth 1 https://github.com/siggy/gographs /repos/https://github.com/siggy/gographs"
// 2. dir => dot
//    goda graph -short -cluster github.com/siggy/gographs...

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/tools/go/vcs"
)

func repoToDot(repo string, cluster bool) (string, error) {
	codeDir, err := toDir(repo)
	if err != nil {
		log.Errorf("failed to get dir: %s", err)
		return "", err
	}

	return dirToDot(codeDir, cluster)
}

func toDir(repo string) (string, error) {
	codeDir, err := ioutil.TempDir("", "")
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

	return codeDir, nil
}

func dirToDot(dir string, cluster bool) (string, error) {
	args := []string{"graph", "-short"}
	if cluster {
		args = append(args, "-cluster")
	}
	args = append(args, "./...")

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

func trimScheme(repo string) string {
	schemeSep := "://"
	schemeSepIdx := strings.Index(repo, schemeSep)
	if schemeSepIdx > -1 {
		return repo[schemeSepIdx+len(schemeSep):]
	}

	return repo
}
