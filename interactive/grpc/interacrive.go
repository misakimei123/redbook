package grpc

import (
	"context"

	intrv1 "github.com/misakimei123/redbook/api/proto/gen/intr/v1"
	"github.com/misakimei123/redbook/interactive/domain"
	"github.com/misakimei123/redbook/interactive/service"

	"google.golang.org/grpc"
)

type InteractiveServiceServer struct {
	intrv1.UnimplementedInteractiveServiceServer
	svc service.InteractiveService
}

func NewInteractiveServiceServer(svc service.InteractiveService) *InteractiveServiceServer {
	return &InteractiveServiceServer{svc: svc}
}

func (i *InteractiveServiceServer) Register(server *grpc.Server) {
	intrv1.RegisterInteractiveServiceServer(server, i)
}

func (i *InteractiveServiceServer) IncrReadCnt(ctx context.Context, request *intrv1.IncrReadCntRequest) (*intrv1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, request.GetBizStr(), request.GetBizId())
	return &intrv1.IncrReadCntResponse{}, err
}

func (i *InteractiveServiceServer) Like(ctx context.Context, request *intrv1.LikeRequest) (*intrv1.LikeResponse, error) {
	err := i.svc.Like(ctx, request.GetLike(), request.GetBizStr(), request.GetBizId(), request.GetUid())
	return &intrv1.LikeResponse{}, err
}

func (i *InteractiveServiceServer) Collect(ctx context.Context, request *intrv1.CollectRequest) (*intrv1.CollectResponse, error) {
	err := i.svc.Collect(ctx, request.GetBizStr(), request.GetBizId(), request.GetCid(), request.GetUid())
	return &intrv1.CollectResponse{}, err
}

func (i *InteractiveServiceServer) Get(ctx context.Context, request *intrv1.GetRequest) (*intrv1.GetResponse, error) {
	interactive, err := i.svc.Get(ctx, request.GetBizStr(), request.GetBizId(), request.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv1.GetResponse{Interactive: i.toDTO(interactive)}, nil
}

func (i *InteractiveServiceServer) GetByIds(ctx context.Context, request *intrv1.GetByIdsRequest) (*intrv1.GetByIdsResponse, error) {
	res, err := i.svc.GetByIds(ctx, request.GetBizStr(), request.GetIds())
	if err != nil {
		return nil, err
	}
	intrs := make(map[int64]*intrv1.Interactive)
	for k, v := range res {
		intrs[k] = i.toDTO(v)
	}

	return &intrv1.GetByIdsResponse{Interacs: intrs}, nil
}

func (i *InteractiveServiceServer) toDTO(interactive domain.Interactive) *intrv1.Interactive {
	return &intrv1.Interactive{
		BizStr:     interactive.Biz,
		BizId:      interactive.BizId,
		ReadCnt:    interactive.ReadCnt,
		LikeCnt:    interactive.LikeCnt,
		CollectCnt: interactive.CollectCnt,
		Liked:      interactive.Liked,
		Collected:  interactive.Collected,
	}
}
