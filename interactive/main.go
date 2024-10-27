package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	initViperV1()
	app := InitApp()
	for _, consumer := range app.consumers {
		err := consumer.Start()
		if err != nil {
			panic(err)
		}
	}
	go func() {
		err := app.web.Run(":8083")
		if err != nil {
			panic(err)
		}
	}()
	zap.L().Info("grpc initialed")
	//server := grpc.NewServer()
	//intrv1.RegisterInteractiveServiceServer(server, app.server)
	//listen, err := net.Listen("tcp", "8090")
	//if err != nil {
	//	panic(err)
	//}
	//err = server.Serve(listen)
	//if err != nil {
	//	panic(err)
	//}
	err := app.server.Serve()
	if err != nil {
		panic(err)
	}
}

func initPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(":8082", nil)
		if err != nil {
			panic(err)
		}
	}()
}

func initViperV1() {
	cfile := pflag.String("config", "config/config.yaml", "config file path")
	pflag.Parse()
	//viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	//viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("read config fail %s", err))
	}
	fmt.Println(viper.Get("test.key"))
}
