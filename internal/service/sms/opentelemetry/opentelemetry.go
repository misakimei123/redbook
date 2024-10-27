package opentelemetry

import (
	"context"

	"github.com/misakimei123/redbook/internal/service/sms"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Decorator struct {
	svc    sms.SMSService
	tracer trace.Tracer
}

func NewDecorator(svc sms.SMSService, tracer trace.Tracer) *Decorator {
	return &Decorator{svc: svc, tracer: tracer}
}

func (d *Decorator) Send(ctx context.Context, tplId string, args []string, number ...string) error {
	ctx, span := d.tracer.Start(ctx, "sms")
	span.SetAttributes(attribute.String("tplId", tplId))
	span.AddEvent("send sms")
	defer span.End()
	err := d.svc.Send(ctx, tplId, args, number...)
	if err != nil {
		span.RecordError(err)
	}
	return err
}
