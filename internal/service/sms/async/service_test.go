package async

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	smsmocks "github.com/misakimei123/redbook/internal/service/sms/mocks"
	"github.com/misakimei123/redbook/internal/service/sms/repo"
	smsrepomocks "github.com/misakimei123/redbook/internal/service/sms/repo/mock"
	limitmocks "github.com/misakimei123/redbook/pkg/limiter/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAsyncSMS_Send(t *testing.T) {
	testCases := []struct {
		name        string
		mock        func(controller *gomock.Controller) (*smsmocks.MockSMSService, *limitmocks.MockLimiter, *smsrepomocks.MockSMSRepo)
		crashed     atomic.Value
		cnt         int32
		wantCnt     int32
		wantCrashed bool
		wantError   error
	}{
		{
			name: "同步发送成功",
			mock: func(controller *gomock.Controller) (*smsmocks.MockSMSService, *limitmocks.MockLimiter, *smsrepomocks.MockSMSRepo) {
				smsService := smsmocks.NewMockSMSService(controller)
				mockLimiter := limitmocks.NewMockLimiter(controller)
				smsRepo := smsrepomocks.NewMockSMSRepo(controller)
				smsRepo.EXPECT().Get(gomock.Any()).AnyTimes().Return(repo.SMSPara{}, repo.ErrNoSMS)
				mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				smsService.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return smsService, mockLimiter, smsRepo
			},
			cnt:         0,
			wantCnt:     0,
			wantCrashed: false,
			wantError:   nil,
		},
		{
			name: "同步发送超时，短信缓存还未发送",
			mock: func(controller *gomock.Controller) (*smsmocks.MockSMSService, *limitmocks.MockLimiter, *smsrepomocks.MockSMSRepo) {
				smsService := smsmocks.NewMockSMSService(controller)
				mockLimiter := limitmocks.NewMockLimiter(controller)
				smsRepo := smsrepomocks.NewMockSMSRepo(controller)
				smsRepo.EXPECT().Get(gomock.Any()).AnyTimes().Return(repo.SMSPara{}, repo.ErrNoSMS)
				mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				smsService.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(a, b, c, d any) {
					time.Sleep(1600 * time.Millisecond)
				}).Return(nil)
				return smsService, mockLimiter, smsRepo
			},
			cnt:         0,
			wantCnt:     1,
			wantCrashed: false,
			wantError:   nil,
		},
		{
			name: "限速缓存",
			mock: func(controller *gomock.Controller) (*smsmocks.MockSMSService, *limitmocks.MockLimiter, *smsrepomocks.MockSMSRepo) {
				smsService := smsmocks.NewMockSMSService(controller)
				mockLimiter := limitmocks.NewMockLimiter(controller)
				smsRepo := smsrepomocks.NewMockSMSRepo(controller)
				smsRepo.EXPECT().Get(gomock.Any()).AnyTimes().Return(repo.SMSPara{}, repo.ErrNoSMS)
				mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil)
				smsRepo.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil)
				return smsService, mockLimiter, smsRepo
			},
			cnt:         0,
			wantCnt:     0,
			wantCrashed: true,
			wantError:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			controller := gomock.NewController(t)
			smsSvc, l, smsRepo := tc.mock(controller)
			asyncSMS := NewAsyncSMS(smsSvc, l, smsRepo)
			asyncSMS.cnt = tc.cnt
			err := asyncSMS.Send(ctx, "123", []string{"123", "456"}, "13412345678")
			assert.Equal(t, tc.wantCnt, asyncSMS.cnt)
			crashed := asyncSMS.crashed.Load().(bool)
			assert.Equal(t, tc.wantCrashed, crashed)
			assert.Equal(t, tc.wantError, err)
		})
	}
}

func TestAsyncSMS_asyncSend(t *testing.T) {
	testCases := []struct {
		name        string
		mock        func(controller *gomock.Controller) (*smsmocks.MockSMSService, *limitmocks.MockLimiter, *smsrepomocks.MockSMSRepo)
		timer       time.Duration
		wantCrashed bool
		wantSending bool
	}{
		{
			name: "有缓存sms，发送fail",
			mock: func(controller *gomock.Controller) (*smsmocks.MockSMSService, *limitmocks.MockLimiter, *smsrepomocks.MockSMSRepo) {
				smsService := smsmocks.NewMockSMSService(controller)
				mockLimiter := limitmocks.NewMockLimiter(controller)
				smsRepo := smsrepomocks.NewMockSMSRepo(controller)
				smsRepo.EXPECT().Get(gomock.Any()).AnyTimes().Return(repo.SMSPara{
					Id:      0,
					TplId:   "1",
					Args:    []string{"123"},
					Numbers: []string{"123"},
				}, nil)
				smsService.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Do(func(a, b, c, d any) {
					time.Sleep(600 * time.Millisecond)
				}).Return(errors.New("sms fail"))
				//smsRepo.EXPECT().Del4Success(gomock.Any(), int64(0)).AnyTimes().Return(nil)
				smsRepo.EXPECT().Del4Fail(gomock.Any(), int64(0)).MinTimes(1).Return(nil)
				return smsService, mockLimiter, smsRepo
			},
			timer:       6 * time.Second,
			wantCrashed: false,
			wantSending: true,
		},
		{
			name: "有缓存sms，发送成功",
			mock: func(controller *gomock.Controller) (*smsmocks.MockSMSService, *limitmocks.MockLimiter, *smsrepomocks.MockSMSRepo) {
				smsService := smsmocks.NewMockSMSService(controller)
				mockLimiter := limitmocks.NewMockLimiter(controller)
				smsRepo := smsrepomocks.NewMockSMSRepo(controller)
				smsRepo.EXPECT().Get(gomock.Any()).AnyTimes().Return(repo.SMSPara{
					Id:      0,
					TplId:   "1",
					Args:    []string{"123"},
					Numbers: []string{"123"},
				}, nil)
				smsService.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).MinTimes(1).Do(func(a, b, c, d any) {
					time.Sleep(600 * time.Millisecond)
				}).Return(nil)
				smsRepo.EXPECT().Del4Success(gomock.Any(), int64(0)).MinTimes(1).Return(nil)
				//smsRepo.EXPECT().Del4Fail(gomock.Any(), int64(0)).AnyTimes().Return(nil)
				return smsService, mockLimiter, smsRepo
			},
			timer:       3 * time.Second,
			wantCrashed: false,
			wantSending: true,
		},
		{
			name: "无缓存sms",
			mock: func(controller *gomock.Controller) (*smsmocks.MockSMSService, *limitmocks.MockLimiter, *smsrepomocks.MockSMSRepo) {
				smsService := smsmocks.NewMockSMSService(controller)
				mockLimiter := limitmocks.NewMockLimiter(controller)
				smsRepo := smsrepomocks.NewMockSMSRepo(controller)
				smsRepo.EXPECT().Get(gomock.Any()).AnyTimes().Return(repo.SMSPara{}, repo.ErrNoSMS)
				return smsService, mockLimiter, smsRepo
			},
			timer:       3 * time.Second,
			wantCrashed: false,
			wantSending: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			smsSvc, l, smsRepo := tc.mock(controller)
			asyncSMS := NewAsyncSMS(smsSvc, l, smsRepo)
			time.Sleep(tc.timer)
			crashed := asyncSMS.crashed.Load().(bool)
			assert.Equal(t, tc.wantCrashed, crashed)
			sending := asyncSMS.sending.Load().(bool)
			assert.Equal(t, tc.wantSending, sending)
			controller.Finish()
		})
	}
}
