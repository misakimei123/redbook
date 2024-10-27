package startup

import (
	"context"
	"fmt"
	"time"

	"github.com/misakimei123/redbook/internal/repository/dao"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongoDB() *mongo.Database {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	clientOptions := options.Client().ApplyURI("mongodb://192.168.252.128:27017/") //.SetMonitor(&monitor)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic(fmt.Errorf("connect mongodb fail %s", err))
	}
	db := client.Database("webook_dev")
	err = dao.InitCollection(db)
	if err != nil {
		panic(fmt.Errorf("init collection fail %s", err))
	}
	return db
}
