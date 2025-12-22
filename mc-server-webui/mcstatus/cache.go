package mcstatus

import (
	"sync"
	"time"
)

const (
	// CacheExpiration defines how long a cached server status is considered fresh.
	CacheExpiration = 60 * time.Second
	// ErrorCacheExpiration defines how long an error status is considered fresh to prevent hammering unreachable servers.
	ErrorCacheExpiration = 10 * time.Second
)

// cacheEntry holds the server status and the time it was cached.
type cacheEntry struct {
	Status    *ServerStatus
	Timestamp time.Time
}

// ServerStatusCache provides an in-memory cache for Minecraft server statuses.
type ServerStatusCache struct {
	mu    sync.RWMutex
	cache map[string]*cacheEntry
}

// NewServerStatusCache creates and returns a new ServerStatusCache.
func NewServerStatusCache() *ServerStatusCache {
	return &ServerStatusCache{
		cache: make(map[string]*cacheEntry),
	}
}

// Get retrieves a server status from the cache if it's fresh.
func (c *ServerStatusCache) Get(serverName string) (*ServerStatus, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, found := c.cache[serverName]
	if !found {
		return nil, false
	}

	expiration := CacheExpiration
	if !entry.Status.Online { // If server is offline/error, use shorter error cache expiration
		expiration = ErrorCacheExpiration
	}

	if time.Since(entry.Timestamp) < expiration {
		return entry.Status, true // Cache hit and fresh
	}

	return nil, false // Cache hit but stale
}

// Set stores a server status in the cache.
func (c *ServerStatusCache) Set(serverName string, status *ServerStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[serverName] = &cacheEntry{
		Status:    status,
		Timestamp: time.Now(),
	}
}

// Global cache instance
var GlobalServerStatusCache = NewServerStatusCache()
