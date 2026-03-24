package ttl

import "time"

type TTLInterface interface {
	IsExpired(key string) bool
	IsExpiredAndClean(key string) bool
	StartTTLCleaner(interval time.Duration)
	Expire(key string, seconds int) (bool, error)
	CleanExpiredKeys()
	TTL(key string) (int, error)
	Persist(key string) (bool, error)
}
