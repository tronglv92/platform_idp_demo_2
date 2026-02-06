package consumer

import (
	"context"
	"go-service-template/helper/client/incident"
	"go-service-template/helper/queue/consumer"
	"go-service-template/helper/recovery"
	"go-service-template/internal/config"
	"go-service-template/internal/registry"
)

type (
	Consumer interface {
		Process(ctx context.Context, payload consumer.MessageContext) error
	}
	Handler struct {
		appContext registry.ConsumerContext
		config     config.Config
		consumers  map[string]Consumer
	}
)

func NewHandler(appContext registry.ConsumerContext) Handler {
	return Handler{
		appContext: appContext,
		config:     appContext.GetConfig(),
		consumers:  map[string]Consumer{},
	}
}

func (h Handler) Consume(ctx context.Context, payload consumer.MessageContext, _ map[string]any) error {
	defer recovery.RecoverWithReporter(ctx,
		recovery.NewReporter(
			h.config.GetServer().GetName(),
			incident.New(h.config.GetInternal()),
		),
	)

	if v, ok := h.consumers[payload.GetQueueName()]; ok {
		return v.Process(ctx, payload)
	}
	return nil
}
