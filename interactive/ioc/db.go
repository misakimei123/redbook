package ioc

import (
	"fmt"

	"github.com/misakimei123/redbook/interactive/repository/dao"
	"github.com/misakimei123/redbook/pkg/gormx"
	"github.com/misakimei123/redbook/pkg/gormx/connpool"
	"github.com/misakimei123/redbook/pkg/logger"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
)

type SrcDB *gorm.DB
type DstDB *gorm.DB

func InitSrcDB(l logger.LoggerV1) SrcDB {
	return initDB(l, "src")
}

func InitDstDB(l logger.LoggerV1) DstDB {
	return initDB(l, "dst")
}

func InitDoubleWritePool(src SrcDB, dst DstDB, l logger.LoggerV1) *connpool.DoubleWritePool {
	return connpool.NewDoubleWritePool(src, dst, l)
}

func InitBizDB(p *connpool.DoubleWritePool) *gorm.DB {
	doubleWrite, err := gorm.Open(mysql.New(mysql.Config{
		Conn: p,
	}))
	if err != nil {
		panic(err)
	}
	return doubleWrite
}

func initDB(l logger.LoggerV1, key string) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var c Config
	err := viper.UnmarshalKey("db."+key, &c)
	if err != nil {
		panic(fmt.Errorf("read config db fail %s", err))
	}
	db, err := gorm.Open(mysql.Open(
		c.DSN), &gorm.Config{
		Logger: glogger.New(gormLoggerFunc(l.Debug), glogger.Config{
			SlowThreshold: 0,
			LogLevel:      glogger.Info,
		}),
	})
	if err != nil {
		panic("connect db fail")
	}
	err = db.Use(prometheus.New(prometheus.Config{
		DBName:          "webook" + key,
		RefreshInterval: 15,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{
				VariableNames: []string{"thread_running"},
			},
		},
	}))
	if err != nil {
		panic(err)
	}

	cb := gormx.NewCallbacks(prometheus2.SummaryOpts{
		Namespace: "ahyang",
		Subsystem: "webook",
		Name:      "gorm_db" + key,
		Help:      "统计GORM的数据库查询",
		ConstLabels: map[string]string{
			"instance_id": "myInstance",
		},
		Objectives: map[float64]float64{
			0.5:   0.1,
			0.75:  0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	})

	err = db.Use(cb)
	if err != nil {
		panic(err)
	}

	err = db.Use(tracing.NewPlugin(tracing.WithoutMetrics(), tracing.WithDBName("webook"+key)))
	if err != nil {
		panic(err)
	}

	err = dao.InitTables(db)
	if err != nil {
		panic("init table fail")
	}

	return db
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(s string, i ...interface{}) {
	g(s, logger.Field{
		Key: "args",
		Val: i,
	})
}
