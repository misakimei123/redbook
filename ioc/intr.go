package ioc

import (
	"github.com/fsnotify/fsnotify"
	intrv1 "github.com/misakimei123/redbook/api/proto/gen/intr/v1"
	"github.com/misakimei123/redbook/interactive/service"
	"github.com/misakimei123/redbook/internal/client"
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitIntrClient(svc service.InteractiveService) intrv1.InteractiveServiceClient {
	type Config struct {
		Addr      string `yaml:"addr"`
		Secure    bool   `yaml:"secure"`
		Threshold int32  `yaml:"threshold"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
	if err != nil {
		panic(err)
	}
	var opts []grpc.DialOption
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	cc, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		return nil
	}
	remote := intrv1.NewInteractiveServiceClient(cc)
	local := client.NewLocalInteractiveServiceAdapter(svc)
	intrClient := client.NewInteractiveClient(remote, local)
	intrClient.UpdateThreshold(cfg.Threshold)
	viper.OnConfigChange(func(in fsnotify.Event) {
		var cfg Config
		err := viper.UnmarshalKey("grpc.client.intr", &cfg)
		if err != nil {
			panic(err)
		}
		intrClient.UpdateThreshold(cfg.Threshold)
	})
	return intrClient
}

func InitIntrClientV1(cli *etcdv3.Client) intrv1.InteractiveServiceClient {
	type Config struct {
		Addr   string `yaml:"addr"`
		Secure bool   `yaml:"secure"`
	}
	var cfg Config
	err := viper.UnmarshalKey("grpc.client.intr", &cfg)
	if err != nil {
		panic(err)
	}
	var opts []grpc.DialOption
	if !cfg.Secure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	etcdResolver, err := resolver.NewBuilder(cli)
	opts = append(opts, grpc.WithResolvers(etcdResolver))
	cc, err := grpc.Dial(cfg.Addr, opts...)
	if err != nil {
		return nil
	}
	remote := intrv1.NewInteractiveServiceClient(cc)
	return remote
}
