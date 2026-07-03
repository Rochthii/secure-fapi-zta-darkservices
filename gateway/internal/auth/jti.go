package auth

import (
	"sync"
	"time"
)

type JTICache struct {
	jtis sync.Map // Map containing jti -> expiration time
}

var (
	jtiInstance *JTICache
	jtiOnce     sync.Once
)

// GetJTICache returns the singleton instance of JTICache
func GetJTICache() *JTICache {
	jtiOnce.Do(func() {
		jtiInstance = &JTICache{}
		go jtiInstance.startCleanupTicker()
	})
	return jtiInstance
}

// IsJTIUsedAndSave checks if JTI is already used.
// If not used, it saves it with a TTL and returns false.
// If it was already used, it returns true (replay detected).
func (c *JTICache) IsJTIUsedAndSave(jti string, ttl time.Duration) bool {
	now := time.Now()
	val, loaded := c.jtis.LoadOrStore(jti, now.Add(ttl))
	if loaded {
		expireTime := val.(time.Time)
		if now.Before(expireTime) {
			return true // Replay attack detected!
		}
		c.jtis.Store(jti, now.Add(ttl))
		return false
	}
	return false
}

// startCleanupTicker runs a periodic cleanup job
func (c *JTICache) startCleanupTicker() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		now := time.Now()
		c.jtis.Range(func(key, val interface{}) bool {
			expireTime := val.(time.Time)
			if now.After(expireTime) {
				c.jtis.Delete(key)
			}
			return true
		})
	}
}
