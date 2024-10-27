package client

import (
	"context"
	"math/rand"

	"github.com/ecodeclub/ekit/syncx/atomicx"
	intrv1 "github.com/misakimei123/redbook/api/proto/gen/intr/v1"
	"google.golang.org/grpc"
)

func NewInteractiveClient(remote intrv1.InteractiveServiceClient, local intrv1.InteractiveServiceClient) *InteractiveClient {
	return &InteractiveClient{remote: remote, local: local, threshold: atomicx.NewValue[int32]()}
}

type InteractiveClient struct {
	remote    intrv1.InteractiveServiceClient
	local     intrv1.InteractiveServiceClient
	threshold *atomicx.Value[int32]
}

func (i *InteractiveClient) IncrReadCnt(ctx context.Context, in *intrv1.IncrReadCntRequest, opts ...grpc.CallOption) (*intrv1.IncrReadCntResponse, error) {
	return i.selectClient().IncrReadCnt(ctx, in, opts...)
}

func (i *InteractiveClient) Like(ctx context.Context, in *intrv1.LikeRequest, opts ...grpc.CallOption) (*intrv1.LikeResponse, error) {
	return i.selectClient().Like(ctx, in, opts...)
}

func (i *InteractiveClient) Collect(ctx context.Context, in *intrv1.CollectRequest, opts ...grpc.CallOption) (*intrv1.CollectResponse, error) {
	return i.selectClient().Collect(ctx, in, opts...)
}

func (i *InteractiveClient) Get(ctx context.Context, in *intrv1.GetRequest, opts ...grpc.CallOption) (*intrv1.GetResponse, error) {
	return i.selectClient().Get(ctx, in, opts...)
}

func (i *InteractiveClient) GetByIds(ctx context.Context, in *intrv1.GetByIdsRequest, opts ...grpc.CallOption) (*intrv1.GetByIdsResponse, error) {
	return i.selectClient().GetByIds(ctx, in, opts...)
}

func (i *InteractiveClient) selectClient() intrv1.InteractiveServiceClient {
	num := rand.Int31n(100)
	// threshold default 0
	if num < i.threshold.Load() {
		return i.remote
	}
	return i.local
}

func (i *InteractiveClient) UpdateThreshold(threshold int32) {
	i.threshold.Store(threshold)
}
