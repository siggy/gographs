package cache

import (
	"fmt"

	"github.com/go-redis/redis/v7"
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

func (c *Cache) SetURL(k, v string) error {
	return c.client.HSet(urlHash, k, v).Err()
}

func (c *Cache) GetURL(k string) (string, error) {
	return c.client.HGet(urlHash, k).Result()
}

func (c *Cache) SetSVG(repo string, cluster bool, svg string) error {
	return c.client.HSet(svgHash, repoKey(repo, cluster), svg).Err()
}

func (c *Cache) GetSVG(repo string, cluster bool) (string, error) {
	return c.client.HGet(svgHash, repoKey(repo, cluster)).Result()
}

func (c *Cache) SetDOT(repo string, cluster bool, dot string) error {
	return c.client.HSet(dotHash, repoKey(repo, cluster), dot).Err()
}

func (c *Cache) GetDOT(repo string, cluster bool) (string, error) {
	return c.client.HGet(dotHash, repoKey(repo, cluster)).Result()
}

func (c *Cache) SetRepoVersion(repo string, version string) error {
	return c.client.HSet(repoVersionHash, repo, version).Err()
}

func (c *Cache) GetRepoVersion(repo string) (string, error) {
	return c.client.HGet(repoVersionHash, repo).Result()
}

func (c *Cache) SetRepoDir(repo string, version string, repoDir string) error {
	return c.client.HSet(repoDirHash, repoDirKey(repo, version), repoDir).Err()
}

func (c *Cache) GetRepoDir(repo string, version string) (string, error) {
	return c.client.HGet(repoDirHash, repoDirKey(repo, version)).Result()
}

func repoKey(repo string, cluster bool) string {
	return fmt.Sprintf("%s+%t", repo, cluster)
}

func repoDirKey(repo string, version string) string {
	return fmt.Sprintf("%s+%s", repo, version)
}
