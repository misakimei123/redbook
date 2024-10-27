package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/misakimei123/redbook/internal/repository"
	"github.com/misakimei123/redbook/internal/service/sms"
)

var ErrCodeVerifyTooMany = repository.ErrCodeVerifyTooMany

type CodeService interface {
	Send(ctx context.Context, biz, phone string) error
	Verify(ctx context.Context, biz, phone, code string) (bool, error)
}

type codeService struct {
	codeRepository repository.CodeRepository
	sms            sms.SMSService
}

func NewCodeService(repo repository.CodeRepository, sms sms.SMSService) CodeService {
	return &codeService{codeRepository: repo,
		sms: sms}
}

func (c *codeService) Send(ctx context.Context, biz, phone string) error {
	code := c.generate()
	err := c.codeRepository.Set(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	const codeTplId = "1877556"
	return c.sms.Send(ctx, codeTplId, []string{code}, phone)
}

func (c *codeService) Verify(ctx context.Context, biz, phone, code string) (bool, error) {
	result, err := c.codeRepository.Verify(ctx, biz, phone, code)
	if errors.Is(err, repository.ErrCodeVerifyTooMany) {
		return false, nil
	}
	return result, err
}

func (c *codeService) generate() string {
	code := rand.Intn(1000000)
	return fmt.Sprintf("%06d", code)
}
