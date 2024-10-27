package balancer

import (
	"sync"
	"time"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const Name = "custom_weighted_round_robin"

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &PickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type PickerBuilder struct {
}

func (p *PickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	conns := make([]*weightedConn, 0, len(info.ReadySCs))
	for subConn, subConnInfo := range info.ReadySCs {
		md, _ := subConnInfo.Address.Metadata.(map[string]any)
		val, _ := md["weight"]
		weight := val.(float64)
		conns = append(conns, &weightedConn{
			SubConn:       subConn,
			weight:        int(weight),
			currentWeight: int(weight),
			available:     true,
			survivalDur:   time.Minute,
		})
	}
	return &Picker{conns: conns}
}

type Picker struct {
	conns []*weightedConn
	lock  sync.Mutex
}

func (p *Picker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if len(p.conns) == 0 {
		return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
	}

	var total int
	var maxCC *weightedConn
	for _, conn := range p.conns {
		total += conn.weight
		conn.currentWeight = conn.currentWeight + conn.weight
		if !conn.available {
			if time.Since(conn.unavailableTime) > conn.survivalDur {
				conn.available = true
			} else {
				continue
			}
		}
		if maxCC == nil || maxCC.currentWeight < conn.currentWeight {
			maxCC = conn
		}
	}

	maxCC.currentWeight = maxCC.currentWeight - total

	return balancer.PickResult{
		SubConn: maxCC.SubConn,
		Done: func(info balancer.DoneInfo) {
			if info.Err == status.Error(codes.ResourceExhausted, "限流") {
				maxCC.unavailableTime = time.Now()
				maxCC.available = true
			}
		},
	}, nil
}

type weightedConn struct {
	balancer.SubConn
	weight          int
	currentWeight   int
	available       bool
	unavailableTime time.Time
	survivalDur     time.Duration
}
