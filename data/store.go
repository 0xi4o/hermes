package data

import (
	"errors"
	"sync"
	"time"
)

var Store *RedisStore

func InitStore() {
	cache := NewCache()
	Store = &RedisStore{
		Cache: cache,
	}
}

type RedisStore struct {
	Cache Cache
}

type CacheItem struct {
	Value     any
	ExpiresAt int64
}

func NewCacheItem(value any, durationMs int64) CacheItem {
	var expiresAt int64 = -1
	if durationMs > 0 {
		expiresAt = time.Now().UnixMilli() + durationMs
	}
	return CacheItem{
		Value:     value,
		ExpiresAt: expiresAt,
	}
}

type Cache struct {
	Items map[string]CacheItem
	mu    *sync.Mutex
}

func NewCache() Cache {
	items := make(map[string]CacheItem)
	var mu *sync.Mutex

	return Cache{
		Items: items,
		mu:    mu,
	}
}

// TODO: return error
func (cache *Cache) Put(key string, value any, durationMs int64) {
	cacheItem := NewCacheItem(value, durationMs)
	Store.Cache.Items[key] = cacheItem
}

// TODO: return error
func (cache *Cache) Get(key string) (CacheItem, error) {
	item, ok := Store.Cache.Items[key]
	if !ok {
		return CacheItem{}, errors.New("key not found")
	}
	return item, nil
}
