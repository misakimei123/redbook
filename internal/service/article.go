package service

import (
	"context"
	"errors"
	"time"

	eventArticle "github.com/misakimei123/redbook/internal/events/article"
	"github.com/misakimei123/redbook/internal/repository"
	"github.com/misakimei123/redbook/pkg/logger"

	"github.com/misakimei123/redbook/internal/domain"
)

//go:generate mockgen -source=./article.go -package=svcmocks -destination=./mocks/article.mock.go ArticleService
type ArticleService interface {
	Save(ctx context.Context, article domain.Article) (int64, error)
	Publish(ctx context.Context, article domain.Article) (int64, error)
	PublishV1(ctx context.Context, article domain.Article) (int64, error)
	Withdraw(ctx context.Context, id int64, uid int64) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, uid int64, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64, uid int64) (domain.Article, error)
	ListPub(ctx context.Context, start time.Time, end time.Time, offset int, batchSize int) ([]domain.Article, error)
}

type articleService struct {
	repo       repository.ArticleRepository
	authorRepo repository.ArticleAuthorRepository
	readerRepo repository.ArticleReaderRepository
	producer   eventArticle.Producer
	l          logger.LoggerV1
}

func NewArticleService(repo repository.ArticleRepository, producer eventArticle.Producer, log logger.LoggerV1) ArticleService {
	return &articleService{
		repo:     repo,
		producer: producer,
		l:        log,
	}
}

func NewArticleServiceV1(authorRepo repository.ArticleAuthorRepository, readerRepo repository.ArticleReaderRepository) ArticleService {
	return &articleService{
		authorRepo: authorRepo,
		readerRepo: readerRepo,
	}
}

func (a *articleService) GetPubById(ctx context.Context, id int64, uid int64) (domain.Article, error) {
	article, err := a.repo.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	// TODO: send message
	go func() {
		er := a.producer.ProduceReadEvent(eventArticle.ReadEvent{
			Aid: id,
			Uid: uid,
		})
		if er != nil {
			a.l.Error("produce read event fail",
				logger.Int64("aid", id),
				logger.Int64("uid", uid),
				logger.Error(err))
		}
	}()
	return article, nil
}

func (a *articleService) GetById(ctx context.Context, uid int64, id int64) (domain.Article, error) {
	return a.repo.GetById(ctx, uid, id)
}

func (a *articleService) PublishV1(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusPublished
	var (
		id  = article.Id
		err error
	)

	if id > 0 {
		err = a.authorRepo.Update(ctx, article)
	} else {
		id, err = a.authorRepo.Create(ctx, article)
	}

	if err != nil {
		return 0, err
	}

	article.Id = id
	for i := 0; i < 3; i++ {
		err = a.readerRepo.Save(ctx, article)
		if err == nil {
			return id, nil
		}
	}

	return id, errors.New("多次保存线上库失败")
}

func (a *articleService) Publish(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusPublished
	return a.repo.Sync(ctx, article)
}

func (a *articleService) Save(ctx context.Context, article domain.Article) (int64, error) {
	article.Status = domain.ArticleStatusUnpublished
	if article.Id > 0 {
		return article.Id, a.repo.Update(ctx, article)
	}
	return a.repo.Create(ctx, article)
}

func (a *articleService) Withdraw(ctx context.Context, id int64, uid int64) error {
	return a.repo.SyncStatus(ctx, id, uid, domain.ArticleStatusPrivate)
}

func (a *articleService) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	return a.repo.GetByAuthor(ctx, uid, offset, limit)
}

func (a *articleService) ListPub(ctx context.Context, start time.Time, end time.Time, offset int, batchSize int) ([]domain.Article, error) {
	return a.repo.GetPubs(ctx, start.UnixMilli(), end.UnixMilli(), offset, batchSize)
}
