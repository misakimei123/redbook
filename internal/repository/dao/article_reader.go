package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ArticleReaderDao interface {
	Upsert(ctx context.Context, article Article) error
	UpsertV2(ctx context.Context, article PublishedArticle) error
}

type ArticleGormReaderDao struct {
	db *gorm.DB
}

// UpsertV2 同库不同表
func (c *ArticleGormReaderDao) UpsertV2(ctx context.Context, article PublishedArticle) error {
	now := time.Now().UnixMilli()
	article.Ctime = now
	article.Utime = now
	return c.db.Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   article.Title,
			"content": article.Content,
			"status":  article.Status,
			"utime":   now,
		}),
	}).Create(&article).Error
}

// Upsert 不同库，与author使用不同db即可
func (c *ArticleGormReaderDao) Upsert(ctx context.Context, article Article) error {
	now := time.Now().UnixMilli()
	article.Ctime = now
	article.Utime = now
	return c.db.Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   article.Title,
			"content": article.Content,
			"status":  article.Status,
			"utime":   now,
		}),
	}).Create(&article).Error
}

func NewArticleGormReaderDao(db *gorm.DB) ArticleReaderDao {
	return &ArticleGormReaderDao{db: db}
}
