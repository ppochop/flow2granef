package xidcache

/*
Cache for tracking *only* actively timed out flows.

The point is to have a a way to track flow records that can be modified
 because they have been actively timed out recently.

This "freshness" of the active timeout is validated before returning the xid of the record.
*/

import (
	"log/slog"
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
	isPlaceholder    bool
}

type CacheHitResult uint8

const (
	Miss CacheHitResult = iota
	HitPlaceholder
	Hit
)

func New(timeout time.Duration) *IdCache {
	cache := cache.New(timeout, 2*timeout)
	return &IdCache{
		cache:   cache,
		timeout: timeout,
	}
}

func (c *IdCache) Add(commId string, placeholder bool, xid string, lastTs time.Time) error {
	entry := IdCacheEntry{
		xid:              xid,
		timeoutThreshold: lastTs.Add(c.timeout),
		isPlaceholder:    placeholder,
	}
	return c.cache.Add(commId, &entry, cache.DefaultExpiration)
}

func (c *IdCache) Set(commId string, placeholder bool, xid string, lastTs time.Time) {
	entry := IdCacheEntry{
		xid:              xid,
		timeoutThreshold: lastTs.Add(c.timeout),
		isPlaceholder:    placeholder,
	}
	c.cache.Set(commId, &entry, cache.DefaultExpiration)
}

func (c *IdCache) Get(commId string, firstTs time.Time) (string, CacheHitResult) {
	res, found := c.cache.Get(commId)
	if !found {
		return "", Miss
	}
	entry := res.(*IdCacheEntry)
	if firstTs.Round(time.Second).After(entry.timeoutThreshold) {
		return "", Miss
	}
	if entry.isPlaceholder {
		return entry.xid, HitPlaceholder
	}
	return entry.xid, Hit
}

func (c *IdCache) AddOrGet(commId string, placeholder bool, xid string, firstTs time.Time, lastTs time.Time) (string, CacheHitResult) {
	err := c.Add(commId, placeholder, xid, lastTs)

	// Cache miss, use your provided xid
	if err == nil {
		return xid, Miss
	}

	// Cache hit, get the xid from cache
	xidFromCache, hit := c.Get(commId, firstTs)

	if hit == Miss {
		// wtf
		slog.Error("Failed to both add and get xid from cache", "xid", xid, "comm_id", commId, "fromCache", xidFromCache)
		return xid, Miss
	}
	// Should "update" the cache with new timeout
	// NOT thread-safe, should probably be reworked
	// but the risk is acceptable for now
	c.Set(commId, placeholder, xidFromCache, lastTs)
	return xidFromCache, hit
}
