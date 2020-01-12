package cache

import "github.com/go-redis/redis/v7"

type Cache struct {
	client *redis.Client
}

const (
	// http://localhost:8888/repo/github.com/siggy/gographs.svg?cluster=true
	// =>
	// [dot or svg file]
	urlHash = "url"
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
