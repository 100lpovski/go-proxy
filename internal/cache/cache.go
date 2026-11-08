package cache

import (
	"sync"
	"time"
)

type entry struct {
	data      []byte
	headers   http.Header
	status    int
	expiresAt time.Time
}

type Cache struct {
	mu      sync.RWMutex
	items   map[string]*entry
	ttl     time.Duration
	maxSize int
}

func New(ttl time.Duration, maxSize int) *Cache {
	c := &Cache{
		items:   make(map[string]*entry),
		ttl:     ttl,
		maxSize: maxSize,
	}
	go c.janitor()
	return c
}

func (c *Cache) Get(key string) (*entry, bool) {
	c.mu.RLock()
	e, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e, true
}

func (c *Cache) Set(key string, status int, headers http.Header, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// простейшая защита от роста: сносим всё, если переполнились.
	// для продакшена — LRU.
	if len(c.items) >= c.maxSize {
		c.items = make(map[string]*entry)
	}

	c.items[key] = &entry{
		data:      data,
		headers:   headers,
		status:    status,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *Cache) janitor() {
	t := time.NewTicker(c.ttl)
	defer t.Stop()
	for range t.C {
		now := time.Now()
		c.mu.Lock()
		for k, e := range c.items {
			if now.After(e.expiresAt) {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}
