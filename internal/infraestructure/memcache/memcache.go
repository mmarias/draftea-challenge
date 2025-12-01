package memcache

import (
	"errors"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

type Cache interface {
	SetNX(key string) error
	Delete(key string)
}

type cache struct {
	ttl time.Duration
	cli *gocache.Cache
}

var ErrExistCacheKey = errors.New("cache key already exists")

func NewCache(ttl time.Duration) *cache {
	return &cache{
		ttl: ttl,
		cli: gocache.New(ttl, ttl*2),
	}
}

func (c *cache) SetNX(key string) error {
	if _, found := c.cli.Get(key); found {
		return ErrExistCacheKey
	}

	c.cli.Set(key, 1, c.ttl)
	return nil
}

func (c *cache) Delete(key string) {
	c.cli.Delete(key)
}
