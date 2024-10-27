package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/misakimei123/redbook/internal/web/middleware"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"go.uber.org/zap"
)

func main() {
	initViperWatch()
	//initLogger()
	// shutDownFunc := ioc.InitOTEL()
	// defer func() {
	// 	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*3)
	// 	defer cancelFunc()
	// 	shutDownFunc(ctx)
	// }()
	initPrometheus()
	app := InitWebServer()
	// for _, consumer := range app.consumers {
	// 	err := consumer.Start()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }
	// app.cron.Start()
	// defer func() {
	// 	<-app.cron.Stop().Done()
	// }()
	// zap.L().Info("webserver initialed")

	err := app.server.Run(":8081")
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

func useSession(server *gin.Engine) {
	builder := middleware.LoginMiddlewareBuilder{}
	//store := cookie.NewStore([]byte("secret"))
	//authenticated key , encrypt key could not beyond 32 length
	store := memstore.NewStore([]byte(`5ITErZoacSrqncBT7QgFDEsAvFnA31eS`),
		[]byte(`ncRzFOdfkzfmf8GXCvjOOa3dMpaaENtH`))
	//store, err := redis.NewStore(16, "tcp", "127.0.0.1:6379", "",
	//	[]byte(`5ITErZoacSrqncBT7QgFDEsAvFnA31eS`),
	//	[]byte(`ncRzFOdfkzfmf8GXCvjOOa3dMpaaENtH`))
	//if err != nil {
	//	panic(err)
	//}
	server.Use(sessions.Sessions("ssid", store), builder.CheckLogin())
}

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return
	}
	zap.ReplaceGlobals(logger)
}

//func initViperV1() {
//	cfile := pflag.String("config", "config/config.yaml", "config file path")
//	pflag.Parse()
//	//viper.SetConfigName("dev")
//	viper.SetConfigType("yaml")
//	viper.SetConfigFile(*cfile)
//	//viper.AddConfigPath("config")
//	err := viper.ReadInConfig()
//	if err != nil {
//		panic(fmt.Errorf("read config fail %s", err))
//	}
//	fmt.Println(viper.Get("test.key"))
//}

func initViperWatch() {
	cfile := pflag.String("config", "config/dev.yaml", "config file path")
	pflag.Parse()
	//viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	viper.SetConfigFile(*cfile)
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Printf("on change test.key: %s", viper.GetString("test.key"))
	})
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("read config fail %s", err))
	}
	fmt.Println(viper.Get("test.key"))
}

func initViperV2() {
	cfg := `
test:
  key: value1

redis:
  addr: 192.168.252.128:31379

db:
  dsn: root:ahyang@tcp(192.168.252.128:30306)/webook_dev
`
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewReader([]byte(cfg)))
	if err != nil {
		panic(fmt.Errorf("read config fail %s", err))
	}
}

func initViperRemote() {
	err := viper.AddRemoteProvider("etcd3", "http://192.168.252.128:12379", "/webook")
	if err != nil {
		panic(fmt.Errorf("read config fail %s", err))
	}
	viper.SetConfigType("yaml")
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Printf("remote config change")
	})
	err = viper.ReadRemoteConfig()
	if err != nil {
		panic(fmt.Errorf("read config fail %s", err))
	}
	go func() {
		//err := viper.WatchRemoteConfig()
		//if err != nil {
		//	panic(err)
		//}
		for {
			err := viper.WatchRemoteConfig()
			if err != nil {
				panic(err)
			}
			fmt.Println("watch ", viper.GetString("test.key"))
			time.Sleep(time.Second * 3)
		}
	}()
}
