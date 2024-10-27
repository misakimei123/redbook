package ratelimit

import (
	"context"
	"errors"
	"testing"

	"github.com/misakimei123/redbook/internal/service/sms"
	smsmocks "github.com/misakimei123/redbook/internal/service/sms/mocks"
	"github.com/misakimei123/redbook/pkg/limiter"
	limitmocks "github.com/misakimei123/redbook/pkg/limiter/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewRateLimitSMSService_Send(t *testing.T) {
	errLimiter := errors.New("limiter error")
	testCases := []struct {
		name      string
		mock      func(controller *gomock.Controller) (sms.SMSService, limiter.Limiter)
		ctx       context.Context
		wantError error
	}{
		{
			name: "send success",
			mock: func(controller *gomock.Controller) (sms.SMSService, limiter.Limiter) {
				mockLimiter := limitmocks.NewMockLimiter(controller)
				mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(false, nil)
				smsService := smsmocks.NewMockSMSService(controller)
				smsService.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				return smsService, mockLimiter
			},
			wantError: nil,
		},
		{
			name: "limited",
			mock: func(controller *gomock.Controller) (sms.SMSService, limiter.Limiter) {
				mockLimiter := limitmocks.NewMockLimiter(controller)
				mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, nil)
				smsService := smsmocks.NewMockSMSService(controller)
				return smsService, mockLimiter
			},
			wantError: ErrorLimited,
		},
		{
			name: "limiter error",
			mock: func(controller *gomock.Controller) (sms.SMSService, limiter.Limiter) {
				mockLimiter := limitmocks.NewMockLimiter(controller)
				mockLimiter.EXPECT().Limit(gomock.Any(), gomock.Any()).Return(true, errLimiter)
				smsService := smsmocks.NewMockSMSService(controller)
				return smsService, mockLimiter
			},
			wantError: errLimiter,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()
			smsSVC, l := tc.mock(controller)
			limitSMSService := NewRateLimitSMSService(smsSVC, l)
			err := limitSMSService.Send(tc.ctx, "123", []string{"123"}, "123456789")
			assert.Equal(t, tc.wantError, err)
		})
	}
}
