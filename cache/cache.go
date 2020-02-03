package cache

import (
	"fmt"

	"github.com/go-redis/redis/v7"
	log "github.com/sirupsen/logrus"
)

type Cache struct {
	client *redis.Client
	log    *log.Entry
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
	addr := "localhost:6379"

	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}

	return &Cache{
		client: client,
		log: log.WithFields(
			log.Fields{
				"cache": addr,
			},
		),
	}, nil
}

func (c *Cache) SetURL(k, v string) {
	if err := c.client.HSet(urlHash, k, v).Err(); err != nil {
		c.log.Errorf("SetURL failed: %s", err)
	}
}

func (c *Cache) GetURL(k string) (string, error) {
	return c.hget(urlHash, k)
}

func (c *Cache) SetSVG(repo string, cluster bool, svg string) {
	if err := c.client.HSet(svgHash, repoKey(repo, cluster), svg).Err(); err != nil {
		c.log.Errorf("SetSVG failed: %s", err)
	}
}

func (c *Cache) GetSVG(repo string, cluster bool) (string, error) {
	return c.hget(svgHash, repoKey(repo, cluster))
}

func (c *Cache) SetDOT(repo string, cluster bool, dot string) {
	if err := c.client.HSet(dotHash, repoKey(repo, cluster), dot).Err(); err != nil {
		c.log.Errorf("SetDOT failed: %s", err)
	}
}

func (c *Cache) GetDOT(repo string, cluster bool) (string, error) {
	return c.hget(dotHash, repoKey(repo, cluster))
}

func (c *Cache) SetRepoVersion(repo string, version string) {
	if err := c.client.HSet(repoVersionHash, repo, version).Err(); err != nil {
		c.log.Errorf("SetRepoVersion failed: %s", err)
	}
}

func (c *Cache) GetRepoVersion(repo string) (string, error) {
	return c.hget(repoVersionHash, repo)
}

func (c *Cache) SetRepoDir(repo string, version string, repoDir string) {
	if err := c.client.HSet(repoDirHash, repoDirKey(repo, version), repoDir).Err(); err != nil {
		c.log.Errorf("SetRepoDir failed: %s", err)
	}
}

func (c *Cache) GetRepoDir(repo string, version string) (string, error) {
	return c.hget(repoDirHash, repoDirKey(repo, version))
}

func (c *Cache) DelRepoDir(repo string, version string) (int64, error) {
	return c.hdel(repoDirHash, repoDirKey(repo, version))
}

func (c *Cache) RepoScoreIncr(repo string) {
	c.client.ZIncrBy(repoScores, 1, repo)
}

func (c *Cache) RepoScores() ([]string, error) {
	cmd := c.client.ZRevRangeByScore(repoScores, &redis.ZRangeBy{
		Min:    "0",
		Max:    "+inf",
		Offset: 0,
		Count:  10,
	})
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}

	return cmd.Val(), nil
}

func (c *Cache) hget(key, field string) (string, error) {
	c.log.Debugf("hget[%s,%s]", key, field)
	return c.client.HGet(key, field).Result()
}

func (c *Cache) hdel(key, field string) (int64, error) {
	c.log.Debugf("hdel[%s,%s]", key, field)
	return c.client.HDel(key, field).Result()
}

func repoKey(repo string, cluster bool) string {
	return fmt.Sprintf("%s+%t", repo, cluster)
}

func repoDirKey(repo string, version string) string {
	return fmt.Sprintf("%s+%s", repo, version)
}
