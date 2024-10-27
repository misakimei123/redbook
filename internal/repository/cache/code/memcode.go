package code

import (
	"context"
	"fmt"
	lru "github.com/hashicorp/golang-lru/v2/expirable"
	"sync"
	"time"
)

type MemCodeCache struct {
	cache      *lru.LRU[string, Items]
	lock       sync.Mutex
	expiration time.Duration
}

type Items struct {
	code   string
	cnt    int
	expire time.Time
}

func NewMemCodeCache(cache *lru.LRU[string, Items], expiration time.Duration) CodeCache {
	return &MemCodeCache{
		cache:      cache,
		expiration: expiration,
	}
}

func (m *MemCodeCache) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}

func (m *MemCodeCache) Set(ctx context.Context, biz, phone, code string) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	key := m.key(biz, phone)
	items, ok := m.cache.Get(key)
	if !ok || items.expire.Sub(time.Now()) < 9*time.Minute {
		m.cache.Add(key, Items{
			code:   code,
			cnt:    3,
			expire: time.Now().Add(m.expiration),
		})
		return nil
	}
	return ErrCodeSendTooMany
}

func (m *MemCodeCache) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	key := m.key(biz, phone)
	items, ok := m.cache.Get(key)
	if !ok || items.cnt <= 0 {
		return false, ErrCodeVerifyFail
	}

	if items.code == code {
		items.cnt = 0
		return true, nil
	} else {
		items.cnt--
		return false, ErrCodeVerifyFail
	}
}
