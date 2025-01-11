package xidcache

/*
Cache for tracking *only* actively timed out flows.

The point is to have a a way to track flow records that can be modified
 because they have been actively timed out recently.

This "freshness" of the active timeout is validated before returning the xid of the record.
*/

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type IdCache struct {
	cache   *cache.Cache
	timeout time.Duration
}

type IdCacheEntry struct {
	xid              string
	timeoutThreshold time.Time
}

func New(timeout time.Duration) *IdCache {
	cache := cache.New(timeout, 2*timeout)
	return &IdCache{
		cache: cache,
	}
}

func (c *IdCache) Add(commId string, xid string, lastTs time.Time) error {
	entry := IdCacheEntry{
		xid:              xid,
		timeoutThreshold: lastTs.Add(c.timeout),
	}
	return c.cache.Add(commId, &entry, cache.DefaultExpiration)
}

func (c *IdCache) Set(commId string, xid string, lastTs time.Time) {
	entry := IdCacheEntry{
		xid:              xid,
		timeoutThreshold: lastTs.Add(c.timeout),
	}
	c.cache.Set(commId, &entry, cache.DefaultExpiration)
}

func (c *IdCache) Get(commId string) (string, bool) {
	res, found := c.cache.Get(commId)
	if !found {
		return "", false
	}
	now := time.Now().Round(time.Second)
	entry := res.(*IdCacheEntry)
	if now.After(entry.timeoutThreshold) {
		return "", false
	}
	return entry.xid, true
}

func (c *IdCache) AddOrGet(commId string, xid string, lastTs time.Time) (string, bool) {
	err := c.Add(commId, xid, lastTs)

	// Cache miss, use your provided xid
	if err == nil {
		return xid, false
	}

	// Cache hit, get the xid from cache
	xidFromCache, _ := c.Get(commId)

	// Should "update" the cache with new timeout
	// NOT thread-safe, should probably be reworked
	// but the risk is acceptable for now
	c.Set(commId, xidFromCache, lastTs)
	return xidFromCache, true
}
