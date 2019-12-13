package cache

import "sync"

type Cache struct {
	data map[string]string
	sync.RWMutex
}

func New() *Cache {
	return &Cache{
		data: map[string]string{},
	}
}

func (c *Cache) Set(k, v string) {
	c.Lock()
	defer c.Unlock()

	c.data[k] = v
}

func (c *Cache) Get(k string) (string, bool) {
	c.RLock()
	defer c.RUnlock()

	v, ok := c.data[k]
	return v, ok
}
