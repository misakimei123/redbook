package ioc

import (
	"context"
	"fmt"
	"time"

	"github.com/misakimei123/redbook/internal/repository/dao"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongoDB() *mongo.Database {
	type config struct {
		DSN string `yaml:"DSN"`
		db  string `yaml:"db"`
	}
	var c config
	err := viper.UnmarshalKey("mongodb", &c)
	if err != nil {
		panic(fmt.Errorf("read config mongodb fail %s", err))
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	clientOptions := options.Client().ApplyURI(c.DSN) //.SetMonitor(&monitor)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(fmt.Errorf("connect mongodb fail %s", err))
	}
	db := client.Database(c.db)
	err = dao.InitCollection(db)
	if err != nil {
		panic(fmt.Errorf("init collection fail %s", err))
	}
	return db
}
