package cache

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/valkey-io/valkey-go"
)

// Cache holds Valkey client and log state.
type Cache struct {
	client valkey.Client
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
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{addr},
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	err = client.Do(ctx, client.B().Ping().Build()).Error()
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
	c.client.Do(
		context.Background(),
		c.client.B().Zincrby().Key(repoScores).Increment(1).Member(repo).Build(),
	)
}

// RepoScores returns the top-10 most popular repos
func (c *Cache) RepoScores() ([]string, error) {
	return c.client.Do(
		context.Background(),
		c.client.B().Zrevrangebyscore().Key(repoScores).
			Max("+inf").Min("0").Limit(0, 1000).Build(),
	).AsStrSlice()
}

func (c *Cache) hget(key, field string) (string, error) {
	c.log.Tracef("hget[%s,%s]", key, field)
	return c.client.Do(
		context.Background(), c.client.B().Hget().Key(key).Field(field).Build(),
	).ToString()
}

func (c *Cache) hset(key, field string, value string) error {
	c.log.Tracef("hset[%s,%s]", key, field)
	return c.client.Do(
		context.Background(),
		c.client.B().Hset().Key(key).FieldValue().FieldValue(field, value).Build(),
	).Error()
}

func (c *Cache) hdel(key, field string) (int64, error) {
	c.log.Debugf("hdel[%s,%s]", key, field)
	return c.client.Do(
		context.Background(),
		c.client.B().Hdel().Key(key).Field(field).Build(),
	).AsInt64()
}

func repoKey(repo string, cluster bool) string {
	return fmt.Sprintf("%s+%t", repo, cluster)
}
