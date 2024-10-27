package lock

import "time"

type Lock interface {
	AcquireLock(key string, ttl time.Duration) (bool, error)
	ReleaseLock(key string) error
	AutoRefresh(key string, ttl time.Duration, interval time.Duration) error
}
