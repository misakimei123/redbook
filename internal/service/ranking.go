package service

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/ecodeclub/ekit/slice"
	intrv1 "github.com/misakimei123/redbook/api/proto/gen/intr/v1"
	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/repository"
	"github.com/misakimei123/redbook/pkg/container/heap"
	"github.com/misakimei123/redbook/pkg/logger"
)

type RankingService interface {
	GetTopN(ctx context.Context) ([]domain.Article, error)
	Rank(ctx context.Context) error
}

type ArticleRankingService struct {
	repo      repository.RankingRepository
	artSvc    ArticleService
	intraSvc  intrv1.InteractiveServiceClient
	before    time.Duration
	n         int
	l         logger.LoggerV1
	batchSize int
	bizStr    string
	scoreFunc func(likeCnt int64, utime time.Time) float64
}

func (a *ArticleRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return a.repo.GetTopN(ctx)
}

func (a *ArticleRankingService) Rank(ctx context.Context) error {
	articles, err := a.doRanking(ctx)
	if err != nil {
		return err
	}
	err = a.repo.SetTopN(ctx, articles)
	if err != nil {
		a.l.Error("set Top n fail", logger.Error(err))
	}
	return nil
}

func NewArticleRankingService(repo repository.RankingRepository,
	artSvc ArticleService,
	intraSvc intrv1.InteractiveServiceClient,
	l logger.LoggerV1) RankingService {
	return &ArticleRankingService{
		repo:      repo,
		artSvc:    artSvc,
		intraSvc:  intraSvc,
		before:    time.Hour * 24 * 7,
		n:         10,
		l:         l,
		batchSize: 50,
		bizStr:    "article",
		scoreFunc: func(likeCnt int64, utime time.Time) float64 {
			// 时间
			duration := time.Since(utime).Seconds()
			return float64(likeCnt-1) / math.Pow(duration+2, 1.5)
		},
	}
}

func (a *ArticleRankingService) doRanking(ctx context.Context) ([]domain.Article, error) {
	a.l.Info("start ranking")
	now := time.Now()
	start := now.Add(-1 * a.before)
	minHeap := heap.NewLocalMinHeap[domain.Article](a.n)
	offset := 0
	for {
		arts, err := a.artSvc.ListPub(ctx, start, now, offset, a.batchSize)
		if err != nil {
			return nil, err
		}

		ids := slice.Map(arts, func(idx int, art domain.Article) int64 {
			return art.Id
		})
		resp, err := a.intraSvc.GetByIds(ctx, &intrv1.GetByIdsRequest{
			BizStr: a.bizStr,
			Ids:    ids,
		})

		if err != nil {
			return nil, err
		}
		intrasMap := resp.Interacs

		for _, art := range arts {
			intra := intrasMap[art.Id]
			if intra == nil {
				continue
			}
			artScore := a.scoreFunc(intra.LikeCnt, art.Utime)
			er := minHeap.Push(art, artScore)
			if errors.Is(er, heap.ErrHeapFull) {
				articleMin, scoreMin, _ := minHeap.Pop()
				var pushErr error
				if scoreMin < artScore {
					pushErr = minHeap.Push(art, artScore)
				} else {
					pushErr = minHeap.Push(articleMin, scoreMin)
				}
				if pushErr != nil {
					continue
				}
			}
		}

		offset = offset + len(arts)

		if len(arts) < a.batchSize {
			break
		}
	}
	length := minHeap.Len()
	res := make([]domain.Article, length)
	for i := length - 1; i >= 0; i-- {
		article, _, err := minHeap.Pop()
		if err != nil {
			return res, nil
		}
		res[i] = article
	}
	return res, nil
}
