package integration

import (
	"context"
	"testing"
	"time"

	"github.com/misakimei123/redbook/internal/domain"
	"github.com/misakimei123/redbook/internal/integration/startup"
	"github.com/misakimei123/redbook/internal/job"
	"github.com/misakimei123/redbook/internal/repository/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

func TestScheduler(t *testing.T) {
	suite.Run(t, &SchedulerTestSuite{})
}

type SchedulerTestSuite struct {
	suite.Suite
	scheduler *job.Scheduler
	db        *gorm.DB
}

func (s *SchedulerTestSuite) SetupSuite() {
	s.db = startup.InitDB()
	s.scheduler = startup.InitJobScheduler()
}

func (s *SchedulerTestSuite) TearDownSuite() {
	err := s.db.Exec("TRUNCATE TABLE `jobs`").Error
	assert.NoError(s.T(), err)
}

func (s *SchedulerTestSuite) TestSchedule() {
	t := s.T()
	testCases := []struct {
		name     string
		before   func(t *testing.T)
		after    func(t *testing.T)
		interval time.Duration
		wantErr  error
		wantJob  *testJob
	}{
		{
			name: "test job",
			before: func(t *testing.T) {
				now := time.Now().UnixMilli()
				j := dao.Job{
					Id:         1,
					Expression: "*/5 * * * * ?",
					Executor:   "local",
					Name:       "test_job",
					NextTime:   now,
					Ctime:      now,
					Utime:      now,
				}
				err := s.db.Create(&j).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				var j dao.Job
				err := s.db.Where("id=?", 1).First(&j).Error
				assert.NoError(t, err)
				assert.True(t, j.NextTime > time.Now().UnixMilli())
				assert.True(t, j.Ctime > 0)
				assert.True(t, time.Now().UnixMilli() < j.Utime+time.Second.Milliseconds())
				j.NextTime = 0
				j.Utime = 0
				j.Ctime = 0
				assert.Equal(t, dao.Job{
					Id:         1,
					Expression: "*/5 * * * * ?",
					Executor:   "local",
					Name:       "test_job",
					NextTime:   0,
					Ctime:      0,
					Utime:      0,
					Version:    1,
				}, j)
			},
			interval: 1 * time.Second,
			wantErr:  context.DeadlineExceeded,
			wantJob:  &testJob{cnt: 1},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			exec := job.NewLocalFuncExecutor()
			j := &testJob{}
			exec.RegisterLocalFunc("test_job", j.Do)
			s.scheduler.RegisterExecutor(exec)
			ctx, cancelFunc := context.WithTimeout(context.Background(), tc.interval)
			defer cancelFunc()
			err := s.scheduler.Schedule(ctx)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantJob, j)
		})
	}
}

type testJob struct {
	cnt int
}

func (t *testJob) Do(ctx context.Context, j domain.Job) error {
	t.cnt++
	return nil
}
