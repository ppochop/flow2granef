package xidcache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type DuplCache struct {
	cache *cache.Cache
}

type DuplCacheEntry struct {
	firstTs time.Time
	lastTs  time.Time
	source  string
}

func NewDuplCache(ttl time.Duration) *DuplCache {
	cache := cache.New(ttl, 2*ttl)
	return &DuplCache{
		cache: cache,
	}
}

func (c *DuplCache) DuplHandle(commId string, firstTs time.Time, lastTs time.Time, source string) (string, bool) {
	entry := DuplCacheEntry{
		firstTs: firstTs,
		lastTs:  lastTs,
		source:  source,
	}
	res, found := c.cache.Get(commId)
	if found {
		resEntry := res.(*DuplCacheEntry)
		if resEntry.lastTs.After(firstTs) && resEntry.firstTs.Before(lastTs) { //overlap
			return resEntry.source, true
		}
	} else {
		c.cache.Set(commId, &entry, cache.DefaultExpiration)
	}
	return "", false
}
