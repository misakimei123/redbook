package dao

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Article struct {
	Id       int64  `gorm:"primaryKey, autoIncrement" bson:"id,omitempty"`
	Title    string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	Content  string `gorm:"type=BLOB" bson:"content,omitempty"`
	AuthorId int64  `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	Utime    int64  `bson:"utime,omitempty"`
}

var ErrArticleAuthorNotMatch = errors.New("id or author id is not correct")

type ArticleDao interface {
	Insert(ctx context.Context, article Article) (int64, error)
	UpdateById(ctx context.Context, article Article) error
	Sync(ctx context.Context, art Article) (int64, error)
	SyncV1(ctx context.Context, art Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, uid int64, status uint8) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error)
	GetById(ctx context.Context, uid int64, id int64) (Article, error)
	GetPubById(ctx context.Context, id int64) (PublishedArticle, error)
	GetPubs(ctx context.Context, start int64, end int64, offset int, limit int) ([]PublishedArticle, error)
}

type PublishedArticle Article

type ArticleGormDao struct {
	db *gorm.DB
}

func NewArticleGormDao(db *gorm.DB) ArticleDao {
	return &ArticleGormDao{
		db: db,
	}
}

func (a *ArticleGormDao) Insert(ctx context.Context, article Article) (int64, error) {
	now := time.Now().UnixMilli()
	article.Ctime = now
	article.Utime = now
	err := a.db.WithContext(ctx).Create(&article).Error
	return article.Id, err
}

func (a *ArticleGormDao) UpdateById(ctx context.Context, article Article) error {
	res := a.db.WithContext(ctx).Model(&Article{}).Where("id=? and author_id=?", article.Id, article.AuthorId).Updates(map[string]any{
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

func (a *ArticleGormDao) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dao := NewArticleGormDao(tx)
		var err error
		if art.Id > 0 {
			err = dao.UpdateById(ctx, art)
		} else {
			id, err = dao.Insert(ctx, art)
			art.Id = id
		}

		if err != nil {
			return err
		}

		now := time.Now().UnixMilli()
		publishedArticle := PublishedArticle(art)
		publishedArticle.Ctime = now
		publishedArticle.Utime = now
		err = tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":   publishedArticle.Title,
				"content": publishedArticle.Content,
				"status":  publishedArticle.Status,
				"utime":   now,
			}),
		}).Create(&publishedArticle).Error
		return err
	})

	if err != nil {
		return 0, err
	}
	return id, err
}

func (a *ArticleGormDao) SyncV1(ctx context.Context, art Article) (int64, error) {
	tx := a.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer tx.Rollback()

	dao := NewArticleGormDao(tx)
	var (
		err error
		id  = art.Id
	)

	if art.Id > 0 {
		err = dao.UpdateById(ctx, art)
	} else {
		id, err = dao.Insert(ctx, art)
		art.Id = id
	}

	if err != nil {
		return 0, err
	}
	now := time.Now().UnixMilli()
	publishedArticle := PublishedArticle(art)
	publishedArticle.Ctime = now
	publishedArticle.Utime = now
	err = tx.Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(map[string]interface{}{
			"title":   publishedArticle.Title,
			"content": publishedArticle.Content,
			"status":  publishedArticle.Status,
			"utime":   now,
		}),
	}).Create(&publishedArticle).Error
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, nil

}

func (a *ArticleGormDao) SyncStatus(ctx context.Context, id int64, uid int64, status uint8) error {
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer tx.Rollback()
		now := time.Now().UnixMilli()

		res := tx.WithContext(ctx).Model(&Article{}).Where("id=? and author_id=?", id, uid).Updates(map[string]any{
			"status": status,
			"utime":  now,
		})

		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrArticleAuthorNotMatch
		}
		res = tx.WithContext(ctx).Model(&PublishedArticle{}).Where("id=? and author_id=?", id, uid).Updates(map[string]any{
			"status": status,
			"utime":  now,
		})

		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrArticleAuthorNotMatch
		}
		return nil
	})
}

func (a *ArticleGormDao) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	var arts []Article
	err := a.db.WithContext(ctx).Model(&Article{}).
		Where("author_id=?", uid).
		Offset(offset).Limit(limit).Order("utime DESC").
		Find(&arts).Error
	return arts, err
}

func (a *ArticleGormDao) GetById(ctx context.Context, uid int64, id int64) (Article, error) {
	var art Article
	res := a.db.WithContext(ctx).Model(&Article{}).Where("id = ? and author_id = ?", id, uid).First(&art)
	if res.Error != nil {
		return art, res.Error
	}
	if res.RowsAffected == 0 {
		return art, ErrArticleAuthorNotMatch
	}
	return art, nil
}

func (a *ArticleGormDao) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	var art PublishedArticle
	err := a.db.WithContext(ctx).Model(&PublishedArticle{}).Where("id = ?", id).First(&art).Error
	return art, err
}

func (a *ArticleGormDao) GetPubs(ctx context.Context, start int64, end int64, offset int, limit int) ([]PublishedArticle, error) {
	var arts []PublishedArticle
	err := a.db.WithContext(ctx).Model(&PublishedArticle{}).Where("utime >= ? and utime <= ?", start, end).Offset(offset).Limit(limit).Find(&arts).Error
	return arts, err
}
