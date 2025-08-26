package cache

import (
	"container/list"
	"log/slog"
	"sync"
	"wb-examples-l0/internal/models"
	"wb-examples-l0/internal/storage/postgres"
)

type cacheItem struct {
	key   string
	value *models.Order
}

type LRUCache struct {
	capacity int
	storage  *postgres.Storage
	list     *list.List
	cache    map[string]*list.Element
	mu       sync.Mutex
	logger   *slog.Logger
}

func NewLRUCache(capacity int, storage *postgres.Storage, logger *slog.Logger) *LRUCache {
	cache := &LRUCache{
		capacity: capacity,
		storage:  storage,
		list:     list.New(),
		cache:    make(map[string]*list.Element),
		mu:       sync.Mutex{},
		logger:   logger,
	}

	go cache.preloadCache()

	return cache
}

func (c *LRUCache) Get(key string) (*models.Order, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[key]; exists {
		c.list.MoveToFront(elem)
		item := elem.Value.(*cacheItem)
		return item.value, true
	}
	return nil, false
}

func (c *LRUCache) Put(key string, val *models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.cache[key]; exists {
		item := elem.Value.(*cacheItem)
		item.value = val
		c.list.MoveToFront(elem)
		c.logger.Debug("Updated existing key in cache", "key", key)
		return
	}

	if c.list.Len() >= c.capacity {
		c.removeOldest()
	}

	item := &cacheItem{key: key, value: val}
	elem := c.list.PushFront(item)
	c.cache[key] = elem
	c.logger.Debug("Added new key to cache", "key", key, "cache_size", c.list.Len())
}

func (c *LRUCache) removeOldest() {
	elem := c.list.Back()
	if elem == nil {
		return
	}

	c.list.Remove(elem)

	item := elem.Value.(*cacheItem)
	delete(c.cache, item.key)
	c.logger.Debug("Removed oldest key from cache", "key", item.key)
}

func (c *LRUCache) preloadCache() {
	uids, err := c.storage.GetAllLimitOrderUIDs(c.capacity)
	if err != nil {
		c.logger.Error("error load cache", "error", err)
		return
	}

	for _, uid := range uids {
		if order, err := c.storage.GetOrderByUID(uid); err == nil {
			c.Put(uid, order)
		} else {
			c.logger.Error("error loading order for cache", "uid", uid, "error", err)
		}
	}
	c.logger.Info("Cache preloaded", "items_loaded", len(uids))
}
