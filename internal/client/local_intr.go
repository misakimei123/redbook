package client

import (
	"context"

	intrv1 "github.com/misakimei123/redbook/api/proto/gen/intr/v1"
	"github.com/misakimei123/redbook/interactive/domain"
	"github.com/misakimei123/redbook/interactive/service"
	"google.golang.org/grpc"
)

type LocalInteractiveServiceAdapter struct {
	svc service.InteractiveService
}

func NewLocalInteractiveServiceAdapter(svc service.InteractiveService) intrv1.InteractiveServiceClient {
	return &LocalInteractiveServiceAdapter{svc: svc}
}

func (l *LocalInteractiveServiceAdapter) IncrReadCnt(ctx context.Context, in *intrv1.IncrReadCntRequest, opts ...grpc.CallOption) (*intrv1.IncrReadCntResponse, error) {
	err := l.svc.IncrReadCnt(ctx, in.GetBizStr(), in.GetBizId())
	return &intrv1.IncrReadCntResponse{}, err
}

func (l *LocalInteractiveServiceAdapter) Like(ctx context.Context, in *intrv1.LikeRequest, opts ...grpc.CallOption) (*intrv1.LikeResponse, error) {
	err := l.svc.Like(ctx, in.GetLike(), in.GetBizStr(), in.GetBizId(), in.GetUid())
	return &intrv1.LikeResponse{}, err
}

func (l *LocalInteractiveServiceAdapter) Collect(ctx context.Context, in *intrv1.CollectRequest, opts ...grpc.CallOption) (*intrv1.CollectResponse, error) {
	err := l.svc.Collect(ctx, in.GetBizStr(), in.GetBizId(), in.GetCid(), in.GetUid())
	return &intrv1.CollectResponse{}, err
}

func (l *LocalInteractiveServiceAdapter) Get(ctx context.Context, in *intrv1.GetRequest, opts ...grpc.CallOption) (*intrv1.GetResponse, error) {
	interactive, err := l.svc.Get(ctx, in.GetBizStr(), in.GetBizId(), in.GetUid())
	if err != nil {
		return nil, err
	}
	return &intrv1.GetResponse{Interactive: l.toDTO(interactive)}, nil
}

func (l *LocalInteractiveServiceAdapter) GetByIds(ctx context.Context, in *intrv1.GetByIdsRequest, opts ...grpc.CallOption) (*intrv1.GetByIdsResponse, error) {
	resp, err := l.svc.GetByIds(ctx, in.GetBizStr(), in.GetIds())
	if err != nil {
		return nil, err
	}
	intrs := make(map[int64]*intrv1.Interactive)
	for k, v := range resp {
		intrs[k] = l.toDTO(v)
	}
	return &intrv1.GetByIdsResponse{Interacs: intrs}, nil
}

func (l *LocalInteractiveServiceAdapter) toDTO(interactive domain.Interactive) *intrv1.Interactive {
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
