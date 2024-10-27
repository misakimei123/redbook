package repository

import (
	"context"

	"github.com/misakimei123/redbook/internal/domain"
)

type ArticleAuthorRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
}

type CachedArticleAuthorRepository struct {
}

func (c *CachedArticleAuthorRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (c *CachedArticleAuthorRepository) Update(ctx context.Context, article domain.Article) error {
	//TODO implement me
	panic("implement me")
}
