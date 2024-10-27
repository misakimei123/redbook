package ratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ecodeclub/ekit/queue"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CounterLimiter struct {
	cnt       atomic.Int32
	threshold int32
}

func (l *CounterLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		cnt := l.cnt.Add(1)
		defer func() {
			l.cnt.Add(-1)
		}()
		if cnt <= l.threshold {
			resp, err = handler(ctx, req)
			return
		}
		return nil, status.Error(codes.ResourceExhausted, "限流")
	}
}

type FixedWindowLimiter struct {
	window          time.Duration
	lastWindowStart time.Time
	cnt             int
	threshold       int
	lock            sync.Mutex
}

func (l *FixedWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		l.lock.Lock()
		now := time.Now()
		if now.After(l.lastWindowStart.Add(l.window)) {
			l.cnt = 0
			l.lastWindowStart = now
		}
		cnt := l.cnt + 1
		l.lock.Unlock()
		if cnt <= l.threshold {
			resp, err = handler(ctx, req)
			return
		}
		return nil, status.Error(codes.ResourceExhausted, "限流")
	}
}

type SlidingWindowLimiter struct {
	window    time.Duration
	queue     queue.ConcurrentPriorityQueue[time.Time]
	lock      sync.Mutex
	cnt       int
	threshold int
}

func (s *SlidingWindowLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		s.lock.Lock()
		now := time.Now()

		for {
			start, err := s.queue.Peek()
			if err == nil && now.After(start.Add(s.window)) {
				s.cnt--
				_, _ = s.queue.Dequeue()
			} else {
				break
			}
		}

		s.queue.Enqueue(now)
		s.cnt++
		cnt := s.queue.Len()
		s.lock.Unlock()
		if cnt < s.threshold {
			resp, err = handler(ctx, req)
			return
		}
		return nil, status.Error(codes.ResourceExhausted, "限流")
	}
}

type TokenBucketLimiter struct {
	intervel  time.Duration
	buckets   chan struct{}
	closeCh   chan struct{}
	closeOnce sync.Once
	ticker    *time.Ticker
}

func (t *TokenBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	t.ticker = time.NewTicker(t.intervel)
	go func() {
		for {
			select {
			case <-t.ticker.C:
				select {
				case t.buckets <- struct{}{}:
				default:
					// bucket full
				}
			case <-t.closeCh:
				return
			}
		}
	}()

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		select {
		case <-t.buckets:
			return handler(ctx, req)
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return nil, status.Error(codes.ResourceExhausted, "限流")
		}
	}
}

func (t *TokenBucketLimiter) Close() error {
	t.closeOnce.Do(func() {
		close(t.closeCh)
		// 停止ticker，以释放资源
		if t.ticker != nil {
			t.ticker.Stop()
		}
	})
	return nil
}

type LeakyBucketLimiter struct {
	interval  time.Duration
	capacity  int // 漏桶的容量
	water     int // 当前桶中的水量
	closeCh   chan struct{}
	closeOnce sync.Once
	ticker    *time.Ticker
	mu        sync.Mutex
}

func (t *LeakyBucketLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	t.ticker = time.NewTicker(t.interval)
	go t.leak() // 启动一个goroutine，定期漏水

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		t.mu.Lock()
		defer t.mu.Unlock()

		if t.water < t.capacity {
			t.water++ // 放入一个水滴（请求）
			return handler(ctx, req)
		}
		// 如果桶已满，则拒绝请求
		return nil, status.Error(codes.ResourceExhausted, "限流")
	}
}

func (t *LeakyBucketLimiter) leak() {
	for {
		select {
		case <-t.ticker.C:
			t.mu.Lock()
			if t.water > 0 {
				t.water-- // 桶中水量减少（漏水）
			}
			t.mu.Unlock()
		case <-t.closeCh:
			t.ticker.Stop() // 停止ticker
			return
		}
	}
}

func (t *LeakyBucketLimiter) Close() error {
	t.closeOnce.Do(func() {
		close(t.closeCh)
		if t.ticker != nil {
			t.ticker.Stop() // 确保在关闭时停止ticker
		}
	})
	return nil
}
