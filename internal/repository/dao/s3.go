package dao

import (
	"bytes"
	"context"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ecodeclub/ekit"
	"github.com/misakimei123/redbook/internal/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ArticleS3dao struct {
	ArticleGormDao
	client *s3.S3
}

func NewArticleS3DAO(db *gorm.DB, client *s3.S3) ArticleDao {
	return &ArticleS3dao{
		ArticleGormDao: ArticleGormDao{db: db},
		client:         client,
	}
}

func (a *ArticleS3dao) Sync(ctx context.Context, art Article) (int64, error) {
	var id = art.Id
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dao := NewArticleGormDao(tx)
		var (
			err error
		)

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
		publishedArticle := &PublishedArticleV2{
			Id:       art.Id,
			Title:    art.Title,
			AuthorId: art.AuthorId,
			Status:   art.Status,
			Ctime:    now,
			Utime:    now,
		}
		err = tx.Clauses(clause.OnConflict{
			DoUpdates: clause.Assignments(map[string]interface{}{
				"title":  publishedArticle.Title,
				"status": publishedArticle.Status,
				"utime":  now,
			}),
		}).Create(&publishedArticle).Error
		return err
	})
	if err != nil {
		return 0, err
	}
	a.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      ekit.ToPtr[string]("webook"),
		Key:         ekit.ToPtr[string](strconv.FormatInt(art.Id, 10)),
		Body:        bytes.NewReader([]byte(art.Content)),
		ContentType: ekit.ToPtr[string]("text/plain;charset=utf-8"),
	})
	return id, err
}

func (a *ArticleS3dao) SyncStatus(ctx context.Context, id int64, uid int64, status uint8) error {
	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

		res = tx.WithContext(ctx).Model(&PublishedArticleV2{}).Where("id=? and author_id=?", id, uid).Updates(map[string]any{
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
	if err != nil {
		return err
	}
	if status == domain.ArticleStatusPrivate {
		_, err := a.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
			Bucket: ekit.ToPtr[string]("webook"),
			Key:    ekit.ToPtr[string](strconv.FormatInt(id, 10)),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

type PublishedArticleV2 struct {
	Id       int64  `gorm:"primaryKey, autoIncrement" bson:"id,omitempty"`
	Title    string `gorm:"type=varchar(4096)" bson:"title,omitempty"`
	AuthorId int64  `gorm:"index" bson:"author_id,omitempty"`
	Status   uint8  `bson:"status,omitempty"`
	Ctime    int64  `bson:"ctime,omitempty"`
	Utime    int64  `bson:"utime,omitempty"`
}
