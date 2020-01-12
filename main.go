package main

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

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/siggy/gographs/cache"
)

type revInfo struct {
	Version string    `json:"Version"`
	Time    time.Time `json:"Time"`
}

// URL -> SVG
// TODO: per-revision caching
var repoCache *cache.Cache

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stderr)

	var err error
	repoCache, err = cache.New()
	if err != nil {
		log.Error(err)
		return
	}

	r := mux.NewRouter()
	r.PathPrefix("/repo").Queries("cluster", "{cluster:true|false}").HandlerFunc(repoHandler)
	r.PathPrefix("/repo").HandlerFunc(repoHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	host := "localhost:8888"
	log.Infof("serving on %s", host)
	err = http.ListenAndServe(host, r)
	if err != nil {
		log.Errorf("Failed to ListenAndServe on %s: %s", host, err)
		return
	}
}

func repoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	svg, err := repoCache.GetURLToSVG(r.URL.String())
	if err != nil {
		route := mux.CurrentRoute(r)
		cluster := mux.Vars(r)["cluster"] == "true"

		tpl, err := route.GetPathTemplate()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		repo := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, tpl+"/"), ".svg")
		svg, err = genSVG(repo, cluster)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		repoCache.SetURLToSVG(r.URL.String(), svg)
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(svg))
}

func genSVG(repo string, cluster bool) (string, error) {
	rev, err := getRev(repo)
	if err != nil {
		log.Errorf("failed to get revision: %s", err)
		return "", err
	}

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

	codeDir := fmt.Sprintf("%s/%s@%s", tmpDir, repo, rev)
	dot, err := runGoda(codeDir, cluster)
	if err != nil {
		log.Errorf("goda failed: %s", err)
	}

	svg, err := dotToSVG(dot)
	if err != nil {
		log.Errorf("error converting dot to svg: %s", err)
		return "", err
	}

	return svg, nil
}

func getRev(repo string) (string, error) {
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
