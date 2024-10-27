package balance

import (
	"context"
)

type Ele[T any] struct {
	Key string
	Val T
}

type LoadBalance[T any] interface {
	Register(ctx context.Context, ele Ele[T], load float64) error
	UpdateLoad(ctx context.Context, load float64) error
	CurNodeIsSuitable(ctx context.Context) (bool, error)
}
