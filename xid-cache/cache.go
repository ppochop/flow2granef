// Package xidcache provides a thread-safe mapping of the Community ID to relevant information about the related flow.
// It allows flow2granef to work with flows from different sources, handle active timeouts and detect duplicity in monitored traffic.
package xidcache

import (
	"sync"
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
	mutex            sync.Mutex
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
	entry := c.getEntry(commId)
	if entry == nil {
		return "", Miss
	}
	entry.mutex.Lock()
	defer entry.mutex.Unlock()
	return entry.evaluateHit(firstTs)
}

func (c *IdCache) AddOrGet(commId string, placeholder bool, xid string, firstTs time.Time, lastTs time.Time) (string, CacheHitResult) {
	err := c.Add(commId, placeholder, xid, lastTs)

	// Cache miss, use your provided xid
	if err == nil {
		return xid, Miss
	}

	// Cache hit, get the xid from cache
	entry := c.getEntry(commId)

	if entry == nil {
		// wtf
		c.Set(commId, placeholder, xid, lastTs)
		return xid, Miss
	}

	newTimeout := lastTs.Add(c.timeout)
	entry.mutex.Lock()
	defer entry.mutex.Unlock()
	xidFromCache, cacheHit := entry.evaluateHit(firstTs)
	switch cacheHit {
	case Miss:
		entry.isPlaceholder = placeholder
		entry.xid = xid
		entry.timeoutThreshold = newTimeout
		return xid, Miss
	default:
		if newTimeout.After(entry.timeoutThreshold) {
			// more recent record
			entry.timeoutThreshold = newTimeout
			entry.isPlaceholder = placeholder
		}
		return xidFromCache, cacheHit
	}
}

func (c *IdCache) getEntry(commId string) *IdCacheEntry {
	res, found := c.cache.Get(commId)
	if !found {
		return nil
	}
	return res.(*IdCacheEntry)
}

func (entry *IdCacheEntry) evaluateHit(firstTs time.Time) (string, CacheHitResult) {
	if firstTs.Round(time.Second).After(entry.timeoutThreshold) {
		return "", Miss
	}
	if entry.isPlaceholder {
		return entry.xid, HitPlaceholder
	}
	return entry.xid, Hit
}
