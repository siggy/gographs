package repo

// curl https://proxy.golang.org/github.com/siggy/heypic/@v/master.info
// curl -O https://proxy.golang.org/github.com/siggy/heypic/@v/v0.0.0-20180506171301-e384f182b391.zip | unzip
// goda graph github.com/linkerd/linkerd2...:root | dot -Tsvg -o graph2.svg

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/siggy/gographs/cache"
	log "github.com/sirupsen/logrus"
)

type revInfo struct {
	Version string    `json:"Version"`
	Time    time.Time `json:"Time"`
}

func GenSVG(cache *cache.Cache, repo string, cluster bool) (string, error) {
	svg, err := cache.GetSVG(repo, cluster)
	if err == nil {
		return svg, nil
	}

	dot, err := GenDOT(cache, repo, cluster)
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

func GenDOT(cache *cache.Cache, repo string, cluster bool) (string, error) {
	dot, err := cache.GetDOT(repo, cluster)
	if err == nil {
		return dot, nil
	}

	rev, err := getRev(cache, repo)
	if err != nil {
		log.Errorf("failed to get revision: %s", err)
		return "", err
	}

	codeDir, err := cache.GetRepoDir(repo, rev)
	if err != nil || !exists(codeDir) {
		zipBody, err := downloadZip(repo, rev)
		if err != nil {
			log.Errorf("failed to download zip: %s", err)
			return "", err
		}

		tmpDir, err := unzip(zipBody)
		if err != nil {
			log.Errorf("failed unzip: %s", err)
			return "", err
		}

		codeDir = fmt.Sprintf("%s/%s@%s", tmpDir, repo, rev)

		go cache.SetRepoDir(repo, rev, codeDir)
	}

	dot, err = runGoda(codeDir, cluster)
	if err != nil {
		log.Errorf("goda failed: %s", err)
		return "", err
	}

	go cache.SetDOT(repo, cluster, dot)

	return dot, nil
}

func getRev(cache *cache.Cache, repo string) (string, error) {
	version, err := cache.GetRepoVersion(repo)
	if err == nil {
		return version, nil
	}

	url := fmt.Sprintf("https://proxy.golang.org/%s/@v/master.info", repo)
	log.Debugf("requesting: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorf("http NewRequest err: %s", err)
		return "", err
	}

	req.Header.Set("User-Agent", "gographs.io/0.1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Errorf("http get err: %s", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("[%d]: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		log.Errorf("%s %s", url, err)
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("ioutil err: %s", err)
		return "", err
	}

	rev := revInfo{}
	err = json.Unmarshal(body, &rev)
	if err != nil {
		log.Errorf("unmarshal err:%s", err)
		return "", err
	}

	go cache.SetRepoVersion(repo, rev.Version)

	return rev.Version, nil
}

func downloadZip(repo string, rev string) ([]byte, error) {
	repoURL := fmt.Sprintf("https://proxy.golang.org/%s/@v/%s.zip", repo, rev)
	log.Debugf("requesting: %s", repoURL)

	rsp2, err := http.Get(repoURL)
	if err != nil {
		log.Errorf("http get err: %s", err)
		return nil, err
	}
	defer rsp2.Body.Close()

	zipBody, err := ioutil.ReadAll(rsp2.Body)
	if err != nil {
		log.Errorf("ioutil err: %s", err)
		return nil, err
	}

	log.Debugf("downloaded %d bytes", len(zipBody))

	return zipBody, nil
}

func unzip(zipBody []byte) (string, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(zipBody), int64(len(zipBody)))
	if err != nil {
		log.Errorf("zip.NewReader err: %s", err)
		return "", err
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Errorf("TempDir err: %s", err)
		return "", err
	}
	log.Debugf("writing to tempDir: %s", tmpDir)

	// Read all the files from zip archive
	for _, f := range zipReader.File {
		fpath := filepath.Join(tmpDir, f.Name)

		err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm)
		if err != nil {
			log.Errorf("os.MkdirAll err: %s", err)
			return "", err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			log.Errorf("os.OpenFile err: %s", err)
			return "", err
		}

		rc, err := f.Open()
		if err != nil {
			log.Errorf("Open err: %s", err)
			return "", err
		}

		_, err = io.Copy(outFile, rc)
		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()
		if err != nil {
			log.Errorf("io.Copy err: %s", err)
			return "", err
		}
	}

	return tmpDir, nil
}

func runGoda(dir string, cluster bool) (string, error) {
	args := []string{"graph", "-short"}
	if cluster {
		args = append(args, "-cluster")
	}
	args = append(args, fmt.Sprintf("./...:root"))

	command := exec.Command("goda", args...)
	command.Dir = dir

	log.Debugf("running: %s", command)
	dot, err := command.Output()
	if err != nil {
		log.Errorf("goda cmd failed: %s", err)
		return "", nil
	}

	return string(dot), nil
}

func dotToSVG(dot string) (string, error) {
	command := exec.Command("dot", "-Gsize=13,7!", "-Tsvg")
	command.Stdin = strings.NewReader(dot)
	var stderr bytes.Buffer
	command.Stderr = &stderr

	log.Debugf("running: %s", command)
	svg, err := command.Output()
	if err != nil {
		log.Errorf("dot cmd failed: %s", err)
		return "", nil
	}

	return string(svg), nil
}

func exists(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}
