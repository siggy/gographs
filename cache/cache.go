package cache

import (
	"fmt"

	"github.com/go-redis/redis/v7"
	log "github.com/sirupsen/logrus"
)

type Cache struct {
	client *redis.Client
}

const (
	// http://localhost:8888/repo/github.com/siggy/gographs.svg?cluster=true
	// =>
	// [dot or svg file]
	urlHash = "url"

	// repo+cluster
	// =>
	// [svg file]
	svgHash = "svg"

	// repo+cluster
	// =>
	// [dot file]
	dotHash = "dot"

	// github.com/prometheus/prometheus
	// =>
	// v1.8.2-0.20200110142541-64194f7d45cb
	repoVersionHash = "repo-version"

	// repo+version
	// =>
	// /tmp/foo
	repoDirHash = "repo-dir"

	// repo
	// =>
	// [numeric popularity score]
	repoScores = "repo-scores"
)

// URL -> SVG
// TODO: per-revision caching

func New() (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &Cache{
		client: client,
	}, nil
}

func (c *Cache) SetURL(k, v string) {
	if err := c.client.HSet(urlHash, k, v).Err(); err != nil {
		log.Errorf("SetURL failed: %s", err)
	}
}

func (c *Cache) GetURL(k string) (string, error) {
	return c.client.HGet(urlHash, k).Result()
}

func (c *Cache) SetSVG(repo string, cluster bool, svg string) {
	if err := c.client.HSet(svgHash, repoKey(repo, cluster), svg).Err(); err != nil {
		log.Errorf("SetSVG failed: %s", err)
	}
}

func (c *Cache) GetSVG(repo string, cluster bool) (string, error) {
	return c.client.HGet(svgHash, repoKey(repo, cluster)).Result()
}

func (c *Cache) SetDOT(repo string, cluster bool, dot string) {
	if err := c.client.HSet(dotHash, repoKey(repo, cluster), dot).Err(); err != nil {
		log.Errorf("SetDOT failed: %s", err)
	}
}

func (c *Cache) GetDOT(repo string, cluster bool) (string, error) {
	return c.client.HGet(dotHash, repoKey(repo, cluster)).Result()
}

func (c *Cache) SetRepoVersion(repo string, version string) {
	if err := c.client.HSet(repoVersionHash, repo, version).Err(); err != nil {
		log.Errorf("SetRepoVersion failed: %s", err)
	}
}

func (c *Cache) GetRepoVersion(repo string) (string, error) {
	return c.client.HGet(repoVersionHash, repo).Result()
}

func (c *Cache) SetRepoDir(repo string, version string, repoDir string) {
	if err := c.client.HSet(repoDirHash, repoDirKey(repo, version), repoDir).Err(); err != nil {
		log.Errorf("SetRepoDir failed: %s", err)
	}
}

func (c *Cache) GetRepoDir(repo string, version string) (string, error) {
	return c.client.HGet(repoDirHash, repoDirKey(repo, version)).Result()
}

func (c *Cache) RepoSetIncr(repo string) {
	c.client.ZIncrBy(repoScores, 1, repo)
}

func repoKey(repo string, cluster bool) string {
	return fmt.Sprintf("%s+%t", repo, cluster)
}

func repoDirKey(repo string, version string) string {
	return fmt.Sprintf("%s+%s", repo, version)
}
