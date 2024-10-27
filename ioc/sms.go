package ioc

import (
	"github.com/misakimei123/redbook/internal/service/sms"
	"github.com/misakimei123/redbook/internal/service/sms/localsms"
)

func InitSMSService() sms.SMSService {
	return localsms.NewLocalSMSService()
}
