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
	Length    int64
}

func NewCacheItem(value any, durationMs int64) *CacheItem {
	var expiresAt int64 = -1
	if durationMs > 0 {
		expiresAt = time.Now().UnixMilli() + durationMs
	}
	return &CacheItem{
		Value:     value,
		ExpiresAt: expiresAt,
	}
}

type Cache struct {
	Items map[string]*CacheItem
	mu    *sync.Mutex
}

func NewCache() Cache {
	items := make(map[string]*CacheItem)
	var mu *sync.Mutex

	return Cache{
		Items: items,
		mu:    mu,
	}
}

// TODO: return error
func (cache *Cache) Append(key string, value []string) error {
	_, err := Store.Cache.Get(key)
	if err != nil {
		Store.Cache.Put(key, []string{}, -1)
	}
	switch items := Store.Cache.Items[key].Value.(type) {
	case []string:
		Store.Cache.Items[key].Value = append(items, value...)
		Store.Cache.Items[key].Length = int64(len(Store.Cache.Items[key].Value.([]string)))
		return nil
	default:
		return errors.New("value is not a list")
	}
}

// TODO: return error
func (cache *Cache) Put(key string, value any, durationMs int64) {
	cacheItem := NewCacheItem(value, durationMs)
	Store.Cache.Items[key] = cacheItem
}

func (cache *Cache) Get(key string) (*CacheItem, error) {
	item, ok := Store.Cache.Items[key]
	if !ok {
		return &CacheItem{}, errors.New("key not found")
	}
	return item, nil
}
