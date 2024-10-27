package domain

import (
	"time"

	"github.com/robfig/cron/v3"
)

type Job struct {
	Id         int64
	CancelFunc func()
	Expression string
	Executor   string
	// job type
	Name string
}

func (j *Job) NextTime() time.Time {
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom |
		cron.Month | cron.Dow | cron.Descriptor)
	s, _ := parser.Parse(j.Expression)
	return s.Next(time.Now())
}
