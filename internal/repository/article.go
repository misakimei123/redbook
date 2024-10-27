package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ecodeclub/ekit/slice"
	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/repository/cache"
	"github.com/misakimei123/redbook/internal/repository/dao"
	"gorm.io/gorm"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
	Sync(ctx context.Context, article domain.Article) (int64, error)
	SyncV1(ctx context.Context, article domain.Article) (int64, error)
	SyncV2(ctx context.Context, article domain.Article) (int64, error)
	SyncStatus(ctx context.Context, id int64, uid int64, articleStatus domain.ArticleStatus) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, uid int64, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
	GetPubs(ctx context.Context, start int64, end int64, offset int, size int) ([]domain.Article, error)
}

type CachedArticleRepository struct {
	dao       dao.ArticleDao
	userRepo  UserRepository
	authorDao dao.ArticleAuthorDao
	readerDao dao.ArticleReaderDao
	db        *gorm.DB
	cache     cache.ArticleCache
}

func (c *CachedArticleRepository) GetPubs(ctx context.Context, start int64, end int64, offset int, size int) ([]domain.Article, error) {
	arts, err := c.dao.GetPubs(ctx, start, end, offset, size)
	if err != nil {
		return nil, err
	}
	res := slice.Map[dao.PublishedArticle, domain.Article](arts, func(idx int, src dao.PublishedArticle) domain.Article {
		return c.toDomain(dao.Article(src))
	})
	return res, nil
}

func (c *CachedArticleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	art, err := c.cache.GetPub(ctx, id)
	if err == nil {
		return art, err
	}
	pubArt, err := c.dao.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	art = c.toDomain(dao.Article(pubArt))
	user, err := c.userRepo.FindByID(ctx, art.Author.Id)
	if err != nil {
		return domain.Article{}, err
	}
	art.Author = domain.Author{
		Id:   art.Author.Id,
		Name: user.Nick,
	}
	go func() {
		err := c.cache.SetPub(ctx, art)
		if err != nil {
			//	TODO: record log
		}
	}()
	return art, err
}

func (c *CachedArticleRepository) GetById(ctx context.Context, uid int64, id int64) (domain.Article, error) {
	art, err := c.cache.Get(ctx, uid, id)
	if err == nil {
		return art, nil
	} else {
		//	TODO: log this error
	}
	article, err := c.dao.GetById(ctx, uid, id)
	if err != nil {
		return domain.Article{}, err
	}
	art = c.toDomain(article)
	err = c.cache.Set(ctx, art)
	if err != nil {
		//	TODO: log this error
	}
	return art, err
}

func NewCachedArticleRepository(articleDao dao.ArticleDao, articleCache cache.ArticleCache, userRepository UserRepository) ArticleRepository {
	return &CachedArticleRepository{dao: articleDao, cache: articleCache, userRepo: userRepository}
}

func NewCachedArticleRepositoryV1(authorDao dao.ArticleAuthorDao, readerDao dao.ArticleReaderDao) ArticleRepository {
	return &CachedArticleRepository{
		authorDao: authorDao,
		readerDao: readerDao,
	}
}

// Sync dao层分发
func (c *CachedArticleRepository) Sync(ctx context.Context, article domain.Article) (int64, error) {
	id, err := c.dao.Sync(ctx, c.toEntity(article))
	if err != nil {
		return 0, err
	}
	err = c.cache.DelFirstPage(ctx, article.Author.Id)
	if err != nil {
		//	TODO: log the error
	}
	go func() {
		user, err := c.userRepo.FindByID(ctx, article.Author.Id)
		if err != nil {
			return
		}
		article.Author = domain.Author{
			Id:   article.Author.Id,
			Name: user.Nick,
		}
		err = c.cache.SetPub(ctx, article)
		if err != nil {
			//	 TODO: record log
		}
	}()
	return id, nil
}

// SyncV2 repo层分发 事务方式，db事务方式只能在同一个库实现
func (c *CachedArticleRepository) SyncV2(ctx context.Context, article domain.Article) (int64, error) {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	defer tx.Rollback()
	authorDao := dao.NewArticleGORMAuthorDao(tx)
	readerDao := dao.NewArticleGormReaderDao(tx)
	art := c.toEntity(article)
	var (
		err error
		id  int64
	)
	if art.Id > 0 {
		err = authorDao.UpdateById(ctx, art)
		id = art.Id
	} else {
		id, err = authorDao.Create(ctx, art)
		art.Id = id
	}

	if err != nil {
		return 0, err
	}
	err = readerDao.UpsertV2(ctx, dao.PublishedArticle(art))
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, nil
}

// SyncV1 repo层分发 非事务 两个库
func (c *CachedArticleRepository) SyncV1(ctx context.Context, article domain.Article) (int64, error) {
	art := c.toEntity(article)
	var (
		err error
		id  int64
	)
	if art.Id > 0 {
		err = c.authorDao.UpdateById(ctx, art)
		id = art.Id
	} else {
		id, err = c.authorDao.Create(ctx, art)
		art.Id = id
	}

	if err != nil {
		return 0, err
	}
	err = c.readerDao.Upsert(ctx, art)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (c *CachedArticleRepository) Update(ctx context.Context, article domain.Article) error {
	err := c.dao.UpdateById(ctx, c.toEntity(article))
	if err != nil {
		return err
	}
	err = c.cache.DelFirstPage(ctx, article.Author.Id)
	if err != nil {
		//	TODO: log the error
	}
	return nil
}

func (c *CachedArticleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	id, err := c.dao.Insert(ctx, c.toEntity(article))
	if err != nil {
		return 0, err
	}
	err = c.cache.DelFirstPage(ctx, article.Author.Id)
	if err != nil {
		//	TODO: log the error
	}
	return id, nil
}

func (c *CachedArticleRepository) toEntity(article domain.Article) dao.Article {
	return dao.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
		Status:   article.Status.ToInt(),
	}
}

func (c *CachedArticleRepository) toDomain(article dao.Article) domain.Article {
	return domain.Article{
		Id:      article.Id,
		Title:   article.Title,
		Content: article.Content,
		Author: domain.Author{
			Id: article.AuthorId,
		},
		Ctime:  time.UnixMilli(article.Ctime),
		Utime:  time.UnixMilli(article.Utime),
		Status: domain.ArticleStatus(article.Status),
	}
}

func (c *CachedArticleRepository) SyncStatus(ctx context.Context, id int64, uid int64, articleStatus domain.ArticleStatus) error {
	err := c.dao.SyncStatus(ctx, id, uid, articleStatus.ToInt())
	if err != nil {
		return err
	}
	err = c.cache.DelFirstPage(ctx, uid)
	if err != nil {
		//	TODO: log the error
	}
	return nil
}

func (c *CachedArticleRepository) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	if offset == 0 && limit == 100 {
		articles, err := c.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return articles, nil
		} else {
			//TODO: record log

		}
	}
	articles, err := c.dao.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	res := slice.Map[dao.Article, domain.Article](articles, func(idx int, src dao.Article) domain.Article {
		return c.toDomain(src)
	})
	go func() {
		if offset == 0 && limit == 100 {
			err = c.cache.SetFirstPage(ctx, uid, res)
			if err != nil {
				//TODO: record log
			}
		}
	}()
	go func() {
		err = c.preCache(ctx, res)
		if err != nil {
			//TODO: record log
		}
	}()

	return res, nil
}

func (c *CachedArticleRepository) preCache(ctx context.Context, articles []domain.Article) error {
	if len(articles) == 0 {
		return fmt.Errorf("no articles need cached")
	}
	const size = 1024 * 1024
	if len(articles[0].Content) > size {
		return fmt.Errorf("content size too long, will not be cached")
	}
	return c.cache.Set(ctx, articles[0])
}
