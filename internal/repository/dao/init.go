package dao

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{}, &Profile{},
		&Article{},
		&PublishedArticle{},
		// &Job{},
	)
}

func InitCollection(db *mongo.Database) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	col := db.Collection("articles")
	_, err := col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{bson.E{"id", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{bson.E{"author_id", 1}}},
	})
	if err != nil {
		return err
	}
	col = db.Collection("published_articles")
	_, err = col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{bson.E{"id", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{bson.E{"author_id", 1}}},
	})
	return err
}
