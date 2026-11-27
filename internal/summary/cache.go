package summary

import "sync"

// cache remembers recent subject → sentence pairs, dropping the oldest when full.
type cache struct {
	mu    sync.Mutex
	max   int
	order []string
	items map[string]string
}

func newCache(max int) *cache {
	return &cache{max: max, items: make(map[string]string)}
}

func (c *cache) get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.items[key]
	return v, ok
}

func (c *cache) put(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.items[key]; !ok {
		c.order = append(c.order, key)
	}
	c.items[key] = value
	if over := len(c.order) - c.max; over > 0 {
		for _, old := range c.order[:over] {
			delete(c.items, old)
		}
		c.order = append([]string(nil), c.order[over:]...)
	}
}
