package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type ArticleAuthorDao interface {
	Create(ctx context.Context, article Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
}

type ArticleGORMAuthorDao struct {
	db *gorm.DB
}

func (c *ArticleGORMAuthorDao) Create(ctx context.Context, article Article) (int64, error) {
	err := c.db.WithContext(ctx).Create(article).Error
	if err != nil {
		return 0, err
	}
	return article.Id, nil
}

func (c *ArticleGORMAuthorDao) UpdateById(ctx context.Context, article Article) error {
	res := c.db.WithContext(ctx).
		Model(&Article{}).
		Where("id=? and author_id=?", article.Id, article.AuthorId).
		Updates(map[string]any{
			"title":   article.Title,
			"content": article.Content,
			"status":  article.Status,
			"utime":   time.Now().UnixMilli(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrArticleAuthorNotMatch
	}
	return nil
}

func NewArticleGORMAuthorDao(db *gorm.DB) ArticleAuthorDao {
	return &ArticleGORMAuthorDao{db: db}
}
