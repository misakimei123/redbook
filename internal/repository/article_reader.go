package repository

import (
	"context"

	"github.com/misakimei123/redbook/internal/domain"
)

type ArticleReaderRepository interface {
	Save(ctx context.Context, article domain.Article) error
	Update(ctx context.Context, article domain.Article) error
}

type CachedArticleReaderRepository struct {
}

func (c *CachedArticleReaderRepository) Save(ctx context.Context, article domain.Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CachedArticleReaderRepository) Update(ctx context.Context, article domain.Article) error {
	//TODO implement me
	panic("implement me")
}
