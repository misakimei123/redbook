package sms

import "context"

type SMSService interface {
	Send(ctx context.Context, tplId string, args []string, number ...string) error
}

type SMSBuilder interface {
	Create(model SMSServiceType) SMSService
}

type SMSServiceBuilder struct {
}

type SMSServiceType int

const (
	Local SMSServiceType = iota
	Tencent
	Limiter
	FailOver
	Auth
)

func (s *SMSServiceBuilder) Create(model SMSServiceType) SMSService {
	switch model {
	case Local:
		return nil
	case Tencent:
		return nil
	case Limiter:
		return nil
	case FailOver:
		return nil
	case Auth:
		return nil
	default:
		return nil
	}
}
