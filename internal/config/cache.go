package config

import (
	"fmt"
	"time"
)

const CacheTTL = 1 * time.Hour

type Cache struct {
	Items     []CacheItem `json:"items"`
	FetchedAt time.Time   `json:"fetched_at"`
	Bindings  []string    `json:"bindings"`
}

type CacheItem struct {
	ID       string `json:"id"`
	Key      string `json:"key"`
	Summary  string `json:"summary"`
	Status   string `json:"status"`
	Type     string `json:"type"`
	Assignee string `json:"assignee"`
}

func (c *Cache) IsFresh() bool {
	return time.Since(c.FetchedAt) < CacheTTL
}

func (c *Cache) MatchesBindings(bindings []string) bool {
	if len(c.Bindings) != len(bindings) {
		return false
	}
	for i, b := range c.Bindings {
		if b != bindings[i] {
			return false
		}
	}
	return true
}

func (c *Cache) GetByKey(key string) (*CacheItem, error) {
	for i := range c.Items {
		if c.Items[i].Key == key {
			return &c.Items[i], nil
		}
	}

	return nil, fmt.Errorf("item not found in cache.")
}

func (c *Cache) indexByKey() map[string]int {
	idx := make(map[string]int, len(c.Items))
	for i, item := range c.Items {
		idx[item.Key] = i
	}
	return idx
}

func (c *Cache) GetByKeyBatch(keys []string) ([]*CacheItem, error) {
	idx := c.indexByKey()

	var cachedTasks []*CacheItem
	for _, t := range keys {
		i, ok := idx[t]
		if !ok {
			return nil, fmt.Errorf("item not found in cache.")
		}
		cachedTasks = append(cachedTasks, &c.Items[i])
	}

	return cachedTasks, nil
}
