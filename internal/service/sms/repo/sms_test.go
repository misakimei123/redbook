package repo

import (
	"context"
	"testing"

	"github.com/misakimei123/redbook/internal/service/sms/repo/dao"
	smsdao "github.com/misakimei123/redbook/internal/service/sms/repo/dao/mock"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSMSRepository_Get(t *testing.T) {
	testCases := []struct {
		name        string
		mock        func(controller *gomock.Controller) dao.SMSDao
		wantSMSPara SMSPara
		wantError   error
	}{
		{
			name: "查到缓存的sms",
			mock: func(controller *gomock.Controller) dao.SMSDao {
				smsDao := smsdao.NewMockSMSDao(controller)
				smsDao.EXPECT().QueryAndUpdate(gomock.Any(), gomock.Any(), gomock.Any()).Return(dao.SMS{
					Id:     0,
					Paras:  `{"Id":0,"TplId":"1","Args":["123"],"Numbers":["123"]}`,
					Status: "Processing",
				}, nil)
				return smsDao
			},
			wantSMSPara: SMSPara{
				Id:      0,
				TplId:   "1",
				Args:    []string{"123"},
				Numbers: []string{"123"},
			},
			wantError: nil,
		},
		{
			name: "no sms",
			mock: func(controller *gomock.Controller) dao.SMSDao {
				smsDao := smsdao.NewMockSMSDao(controller)
				smsDao.EXPECT().QueryAndUpdate(gomock.Any(), gomock.Any(), gomock.Any()).Return(dao.SMS{}, dao.ErrNoSMS)
				return smsDao
			},
			wantSMSPara: SMSPara{},
			wantError:   dao.ErrNoSMS,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			smsDao := tc.mock(controller)
			repo := NewSMSRepository(smsDao)
			smsPara, err := repo.Get(context.Background())
			assert.Equal(t, tc.wantSMSPara, smsPara)
			assert.Equal(t, tc.wantError, err)
		})
	}

}
