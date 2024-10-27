package code

import (
	"context"
	"errors"
)

type CodeCache interface {
	Set(ctx context.Context, biz, phone, code string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

var (
	ErrCodeSendTooMany = errors.New("send code too frequent")
	ErrCodeVerifyFail  = errors.New("verify too frequent")
)
