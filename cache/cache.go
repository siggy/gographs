package cache

import "github.com/go-redis/redis/v7"

type Cache struct {
	client *redis.Client
}

const (
	urlToSVGHash = "url-to-svg"
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

func (c *Cache) SetURLToSVG(k, v string) error {
	return c.client.HSet(urlToSVGHash, k, v).Err()
}

func (c *Cache) GetURLToSVG(k string) (string, error) {
	return c.client.HGet(urlToSVGHash, k).Result()
}
