package grpcx

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/misakimei123/redbook/pkg/logger"
	"github.com/misakimei123/redbook/pkg/netx"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
)

type Server struct {
	*grpc.Server
	EtcdAddr string
	Port     int
	Name     string
	cli      *etcdv3.Client
	kaCancel context.CancelFunc
	L        logger.LoggerV1
}

func (s *Server) Serve() error {
	addr := ":" + strconv.Itoa(s.Port)
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	err = s.Register()
	if err != nil {
		return err
	}
	return s.Server.Serve(listen)
}

func (s *Server) Close() error {
	if s.kaCancel != nil {
		s.kaCancel()
	}
	if s.cli != nil {
		err := s.cli.Close()
		return err
	}
	s.GracefulStop()
	return nil
}

func (s *Server) Register() error {
	cli, err := etcdv3.NewFromURL(s.EtcdAddr)
	s.cli = cli
	if err != nil {
		return err
	}
	em, err := endpoints.NewManager(cli, "service/"+s.Name)
	if err != nil {
		return err
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()
	var ttl int64 = 5
	lease, err := cli.Grant(ctx, ttl)
	if err != nil {
		return err
	}
	addr := netx.GetOutboundIP() + ":" + strconv.Itoa(s.Port)
	key := "service/" + s.Name + "/" + addr
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{Addr: addr}, etcdv3.WithLease(lease.ID))
	if err != nil {
		return err
	}
	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.kaCancel = kaCancel
	ch, err := s.cli.KeepAlive(kaCtx, lease.ID)
	if err != nil {
		return err
	}
	go func() {
		for kaResp := range ch {
			s.L.Debug(kaResp.String())
		}
	}()
	return nil
}
