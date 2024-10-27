package localsms

import (
	"context"
	"log"

	"github.com/misakimei123/redbook/internal/service/sms"
)

type LocalSMSService struct {
}

func NewLocalSMSService() sms.SMSService {
	return &LocalSMSService{}
}

func (s *LocalSMSService) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	log.Println("verify code:", args)
	return nil
}
