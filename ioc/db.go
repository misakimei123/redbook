package ioc

import (
	"fmt"

	"github.com/misakimei123/redbook/internal/repository/dao"
	"github.com/misakimei123/redbook/pkg/gormx"
	"github.com/misakimei123/redbook/pkg/logger"
	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/plugin/opentelemetry/tracing"
	"gorm.io/plugin/prometheus"
)

func InitDB(l logger.LoggerV1) *gorm.DB {
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var c Config
	err := viper.UnmarshalKey("db", &c)
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
		DBName:          "redbook",
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
		Namespace: "misakimei123",
		Subsystem: "redbook",
		Name:      "gorm_db",
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

	err = db.Use(tracing.NewPlugin(tracing.WithoutMetrics(), tracing.WithDBName("redbook_dev")))
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
