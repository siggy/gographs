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
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func repoToDot(repo string, cluster bool) (string, error) {
	codeDir, err := toDir(repo)
	if err != nil {
		log.Errorf("failed to get dir: %s", err)
		return "", err
	}

	return dirToDot(codeDir, cluster)
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
		log.Errorf("goda cmd failed [%s]: %s", err, stderr.String())
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
