package dao

import (
	"context"
	"github.com/bwmarrin/snowflake"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type MongoDBArticleDAO struct {
	db      *mongo.Database
	col     *mongo.Collection
	liveCol *mongo.Collection
	node    *snowflake.Node
}

func NewMongoDBArticleDAO(db *mongo.Database, node *snowflake.Node) ArticleDao {
	return &MongoDBArticleDAO{
		db:      db,
		col:     db.Collection("articles"),
		liveCol: db.Collection("published_articles"),
		node:    node,
	}
}

func (m *MongoDBArticleDAO) Insert(ctx context.Context, article Article) (int64, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	id := m.node.Generate().Int64()
	article.Id = id
	now := time.Now().UnixMilli()
	article.Ctime = now
	article.Utime = now
	_, err := m.col.InsertOne(ctx, article)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *MongoDBArticleDAO) UpdateById(ctx context.Context, article Article) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	filter := bson.D{bson.E{"id", article.Id}, bson.E{"author_id", article.AuthorId}}
	set := bson.D{bson.E{"$set", bson.M{
		"title":   article.Title,
		"content": article.Content,
		"status":  article.Status,
		"utime":   time.Now().UnixMilli(),
	}}}
	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrArticleAuthorNotMatch
	}
	return nil
}

func (m *MongoDBArticleDAO) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		err error
	)

	if art.Id > 0 {
		err = m.UpdateById(ctx, art)
	} else {
		art.Id, err = m.Insert(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	// insert or update

	now := time.Now().UnixMilli()
	art.Utime = now
	filter := bson.D{bson.E{"id", art.Id}, bson.E{"author_id", art.AuthorId}}
	set := bson.D{bson.E{"$set", art},
		bson.E{"$setOnInsert", bson.D{bson.E{"ctime", now - 10}}},
	}
	_, err = m.liveCol.UpdateOne(ctx, filter, set, options.Update().SetUpsert(true))
	if err != nil {
		return 0, err
	}
	return art.Id, nil
}

func (m *MongoDBArticleDAO) SyncV1(ctx context.Context, art Article) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) SyncStatus(ctx context.Context, id int64, uid int64, status uint8) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	filter := bson.D{bson.E{"id", id}, bson.E{"author_id", uid}}
	set := bson.D{bson.E{"$set", bson.M{
		"status": status,
		"utime":  time.Now().UnixMilli(),
	}}}
	res, err := m.col.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	if res.ModifiedCount == 0 {
		return ErrArticleAuthorNotMatch
	}
	res, err = m.liveCol.UpdateOne(ctx, filter, set)
	if err != nil {
		return err
	}
	return nil
}

func (m *MongoDBArticleDAO) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) GetById(ctx context.Context, uid int64, id int64) (Article, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) GetPubById(ctx context.Context, id int64) (PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MongoDBArticleDAO) GetPubs(ctx context.Context, start int64, end int64, offset int, limit int) ([]PublishedArticle, error) {
	//TODO implement me
	panic("implement me")
}
