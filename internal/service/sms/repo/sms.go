package repo

import (
	"context"
	"encoding/json"

	"github.com/misakimei123/redbook/internal/service/sms/repo/dao"
)

type SMSStatus int

func (s SMSStatus) String() string {
	if str, ok := SMSStatusStr[s]; ok {
		return str
	}
	return "Unknown"
}

const (
	Pending SMSStatus = iota
	Processing
	Success
	Fail
)

var (
	ErrNoSMS     = dao.ErrNoSMS
	SMSStatusStr = map[SMSStatus]string{
		Pending:    "Pending",
		Processing: "Processing",
		Success:    "Success",
		Fail:       "Fail",
	}
)

type SMSPara struct {
	Id      int64
	TplId   string
	Args    []string
	Numbers []string
}

type SMSRepo interface {
	Put(ctx context.Context, para SMSPara) error
	Get(ctx context.Context) (SMSPara, error)
	Del4Success(ctx context.Context, id int64) error
	Del4Fail(ctx context.Context, id int64) error
}

type SMSRepository struct {
	dao dao.SMSDao
}

func (s *SMSRepository) Put(ctx context.Context, para SMSPara) error {
	paraJson, err := json.Marshal(para)
	if err != nil {
		return err
	}
	return s.dao.Insert(ctx, dao.SMS{
		Paras:  string(paraJson),
		Status: Pending.String(),
	})
}

func (s *SMSRepository) Get(ctx context.Context) (SMSPara, error) {
	sms, err := s.dao.QueryAndUpdate(ctx, Pending.String(), Processing.String())
	switch err {
	case nil:
		var smsPara SMSPara
		err := json.Unmarshal([]byte(sms.Paras), &smsPara)
		if err != nil {
			return SMSPara{}, err
		}
		return SMSPara{
			Id:      sms.Id,
			TplId:   smsPara.TplId,
			Args:    smsPara.Args,
			Numbers: smsPara.Numbers,
		}, nil
	default:
		return SMSPara{}, err
	}
}

func (s *SMSRepository) Del4Success(ctx context.Context, id int64) error {
	return s.dao.Update(ctx, id, Success.String())
}

func (s *SMSRepository) Del4Fail(ctx context.Context, id int64) error {
	return s.dao.Update(ctx, id, Fail.String())
}

func NewSMSRepository(smsDao dao.SMSDao) SMSRepo {
	return &SMSRepository{
		dao: smsDao,
	}
}
