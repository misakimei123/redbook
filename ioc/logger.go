package ioc

import (
	logger "github.com/misakimei123/redbook/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func InitialLogger() logger.LoggerV1 {
	cfg := zap.NewDevelopmentConfig()
	err := viper.UnmarshalKey("log", &cfg)
	if err != nil {
		panic(err)
	}
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}
