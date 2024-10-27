package job

import "github.com/google/uuid"

type Job interface {
	Name() string
	Run() error
}

var NodeId = uuid.New().String()

type RunNode struct {
	Id string
	// 0~100 越大越空闲
	Performance float64
}
