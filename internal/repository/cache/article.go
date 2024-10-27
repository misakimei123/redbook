package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/redis/go-redis/v9"
)

type ArticleCache interface {
	GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error)
	SetFirstPage(ctx context.Context, uid int64, articles []domain.Article) error
	DelFirstPage(ctx context.Context, uid int64) error
	Get(ctx context.Context, uid int64, id int64) (domain.Article, error)
	Set(ctx context.Context, article domain.Article) error
	GetPub(ctx context.Context, id int64) (domain.Article, error)
	SetPub(ctx context.Context, art domain.Article) error
}

type ArticleRedisCache struct {
	client redis.Cmdable
}

func NewArticleRedisCache(client redis.Cmdable) ArticleCache {
	return &ArticleRedisCache{client: client}
}

func (a *ArticleRedisCache) GetPub(ctx context.Context, id int64) (domain.Article, error) {
	key := a.pubKey(id)
	result, err := a.client.Get(ctx, key).Result()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal([]byte(result), &art)
	return art, err
}

func (a *ArticleRedisCache) SetPub(ctx context.Context, art domain.Article) error {
	key := a.pubKey(art.Id)
	val, err := json.Marshal(art)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, key, val, time.Second*10).Err()
}

func (a *ArticleRedisCache) pubKey(id int64) string {
	return fmt.Sprintf("article:pub:%d:%d", id)
}

func (a *ArticleRedisCache) Get(ctx context.Context, uid int64, id int64) (domain.Article, error) {
	key := a.articleKey(uid, id)
	result, err := a.client.Get(ctx, key).Result()
	if err != nil {
		return domain.Article{}, err
	}
	var art domain.Article
	err = json.Unmarshal([]byte(result), &art)
	return art, err
}

func (a *ArticleRedisCache) Set(ctx context.Context, article domain.Article) error {
	key := a.articleKey(article.Author.Id, article.Id)
	val, err := json.Marshal(article)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, key, val, time.Second*10).Err()
}

func (a *ArticleRedisCache) articleKey(uid int64, id int64) string {
	return fmt.Sprintf("article:detail:%d:%d", uid, id)
}

func (a *ArticleRedisCache) GetFirstPage(ctx context.Context, uid int64) ([]domain.Article, error) {
	key := a.firstPageKey(uid)
	result, err := a.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var res []domain.Article
	err = json.Unmarshal([]byte(result), &res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (a *ArticleRedisCache) SetFirstPage(ctx context.Context, uid int64, articles []domain.Article) error {
	for _, article := range articles {
		article.Content = article.Abstract()
	}
	key := a.firstPageKey(uid)
	val, err := json.Marshal(articles)
	if err != nil {
		return err
	}
	return a.client.Set(ctx, key, val, time.Minute*10).Err()

}

func (a *ArticleRedisCache) DelFirstPage(ctx context.Context, uid int64) error {
	key := a.firstPageKey(uid)
	return a.client.Del(ctx, key).Err()
}

func (a *ArticleRedisCache) firstPageKey(uid int64) string {
	return fmt.Sprintf("article:first_page:%d", uid)
}
