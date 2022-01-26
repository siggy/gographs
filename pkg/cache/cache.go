package cache

import (
	"fmt"

	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

// Cache holds Redis client and log state.
type Cache struct {
	client *redis.Client
	log    *log.Entry
}

const (
	// dot[repo+cluster]
	// github.com/siggy/gographs+false
	// =>
	// [dot file]
	dotHash = "dot"

	// svg[repo+cluster]
	// github.com/siggy/gographs+false
	// =>
	// [svg file]
	svgHash = "svg"

	// repo-scores[repo]
	// github.com/siggy/gographs
	// =>
	// [numeric popularity score]
	repoScores = "reposcores"
)

// New initializes a new cache client.
func New(addr string) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	_, err := client.Ping(client.Context()).Result()
	if err != nil {
		return nil, err
	}

	log := log.WithFields(
		log.Fields{
			"cache": addr,
		},
	)

	registerGauges(client)

	log.Infof("Cache initialized")

	return &Cache{
		client: client,
		log:    log,
	}, nil
}

// Clear deletes all cache entries relevant to a GoLang repo.
func (c *Cache) Clear(repo string) error {
	var rerr error
	_, err := c.hdel(dotHash, repoKey(repo, false))
	if err != nil && rerr == nil {
		rerr = err
	}
	_, err = c.hdel(dotHash, repoKey(repo, true))
	if err != nil && rerr == nil {
		rerr = err
	}

	_, err = c.hdel(svgHash, repoKey(repo, false))
	if err != nil && rerr == nil {
		rerr = err
	}
	_, err = c.hdel(svgHash, repoKey(repo, true))
	if err != nil && rerr == nil {
		rerr = err
	}

	return rerr
}

// SetSVG sets an SVG for a repo.
func (c *Cache) SetSVG(repo string, cluster bool, svg string) {
	if err := c.hset(svgHash, repoKey(repo, cluster), svg); err != nil {
		c.log.Errorf("SetSVG failed: %s", err)
	}
}

// GetSVG gets an SVG for a repo.
func (c *Cache) GetSVG(repo string, cluster bool) (string, error) {
	return c.hget(svgHash, repoKey(repo, cluster))
}

// SetDOT sets a DOT for a repo.
func (c *Cache) SetDOT(repo string, cluster bool, dot string) {
	if err := c.hset(dotHash, repoKey(repo, cluster), dot); err != nil {
		c.log.Errorf("SetDOT failed: %s", err)
	}
}

// GetDOT gets a DOT for a repo.
func (c *Cache) GetDOT(repo string, cluster bool) (string, error) {
	return c.hget(dotHash, repoKey(repo, cluster))
}

// RepoScoreIncr increments the popularity score for a repo.
func (c *Cache) RepoScoreIncr(repo string) {
	c.client.ZIncrBy(c.client.Context(), repoScores, 1, repo)
}

// RepoScores returns the top-10 most popular repos
func (c *Cache) RepoScores() ([]string, error) {
	cmd := c.client.ZRevRangeByScore(c.client.Context(), repoScores, &redis.ZRangeBy{
		Min:    "0",
		Max:    "+inf",
		Offset: 0,
		Count:  1000,
	})
	if cmd.Err() != nil {
		return nil, cmd.Err()
	}

	return cmd.Val(), nil
}

func (c *Cache) hget(key, field string) (string, error) {
	c.log.Tracef("hget[%s,%s]", key, field)
	return c.client.HGet(c.client.Context(), key, field).Result()
}

func (c *Cache) hset(key, field string, value interface{}) error {
	c.log.Tracef("hset[%s,%s]", key, field)
	return c.client.HSet(c.client.Context(), key, field, value).Err()
}

func (c *Cache) hdel(key, field string) (int64, error) {
	c.log.Debugf("hdel[%s,%s]", key, field)
	return c.client.HDel(c.client.Context(), key, field).Result()
}

func repoKey(repo string, cluster bool) string {
	return fmt.Sprintf("%s+%t", repo, cluster)
}
