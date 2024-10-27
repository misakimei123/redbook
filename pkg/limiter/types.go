package limiter

import "context"

type Limiter interface {
	// Limit return true means limit
	Limit(ctx context.Context, key string) (bool, error)
}
