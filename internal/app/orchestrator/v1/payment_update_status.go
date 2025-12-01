package orchestrator

import (
	"context"
	"encoding/json"
	"time"

	"github.com/avast/retry-go"
	"github.com/google/uuid"
	"github.com/mmarias/golearn/internal/domain"
	"github.com/mmarias/golearn/internal/infraestructure/publisher"
)

type UpdatePaymentStatusCommand interface {
	UpdateStatus(ctx context.Context, paymentId string, status domain.PaymentStatus) error
}

type updatePaymentStatusCommand struct {
	publisher  publisher.Client
	repository domain.CommandRepository
}

func NewUpdatePaymentStatusCommand(
	publisher publisher.Client,
	repository domain.CommandRepository,
) *updatePaymentStatusCommand {
	return &updatePaymentStatusCommand{
		publisher,
		repository,
	}
}

func (c *updatePaymentStatusCommand) UpdateStatus(ctx context.Context, paymentId string, status domain.PaymentStatus) error {
	traceID := uuid.NewString()

	event := c.buildEventV1(traceID, paymentId, status)

	b, err := json.Marshal(event)
	if err != nil {
		_ = c.repository.Save(
			ctx,
			domain.CommandStatusFailed,
			event.CommandEvent,
			event.PaymentUpdateStatusEventPayload,
		)
		return err
	}

	err = retry.Do(
		func() error {
			return c.publisher.Publish(ctx, domain.TopicOrchestratorPayment, b)
		},
		retry.Attempts(3),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(100*time.Millisecond),
	)

	if err != nil {
		_ = c.repository.Save(
			ctx,
			domain.CommandStatusFailed,
			event.CommandEvent,
			event.PaymentUpdateStatusEventPayload,
		)
		return err
	}

	return nil
}

func (c *updatePaymentStatusCommand) buildEventV1(traceID, paymentId string, status domain.PaymentStatus) domain.PaymentUpdateStatusEvent {
	return domain.PaymentUpdateStatusEvent{
		CommandEvent: domain.CommandEvent{
			EventType:    domain.PaymentUpdateStatusEventType,
			EventVersion: "1",
			Timestamp:    time.Now().UTC().Format(time.RFC3339),
			CommandEventMetadata: domain.CommandEventMetadata{
				TraceID:        traceID,
				MessageGroupID: paymentId,
				MessageDeduplicationId: domain.BuildDeduplicationId(
					domain.PaymentUpdateStatusEventType,
					paymentId,
				),
			},
		},
		PaymentUpdateStatusEventPayload: domain.PaymentUpdateStatusEventPayload{
			PaymentID: paymentId,
			Status:    status,
		},
	}
}
