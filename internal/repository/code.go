package repository

import (
	"context"

	"github.com/misakimei123/redbook/internal/repository/cache/code"
)

var ErrCodeVerifyTooMany = code.ErrCodeSendTooMany

type CodeRepository interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type CacheCodeRepository struct {
	codeCache code.CodeCache
}

func NewCacheCodeRepository(codeCache code.CodeCache) CodeRepository {
	return &CacheCodeRepository{
		codeCache: codeCache,
	}
}

func (r *CacheCodeRepository) Set(ctx context.Context, biz, phone, code string) error {
	return r.codeCache.Set(ctx, biz, phone, code)
}

func (r *CacheCodeRepository) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	return r.codeCache.Verify(ctx, biz, phone, code)
}
