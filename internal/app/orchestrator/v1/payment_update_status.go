package orchestrator

import (
	"context"
	"encoding/json"
	"log"
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
	publisher publisher.Client
}

func NewUpdatePaymentStatusCommand(
	publisher publisher.Client,
) *updatePaymentStatusCommand {
	return &updatePaymentStatusCommand{
		publisher,
	}
}

func (c *updatePaymentStatusCommand) UpdateStatus(ctx context.Context, paymentId string, status domain.PaymentStatus) error {
	traceID := uuid.NewString()

	b := c.buildEventV1(traceID, paymentId, status)

	return retry.Do(
		func() error {
			return c.publisher.Publish(ctx, domain.TopicOrchestratorPayment, b)
		},
		retry.Attempts(3),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(100*time.Millisecond),
	)

}

func (c *updatePaymentStatusCommand) buildEventV1(traceID, paymentId string, status domain.PaymentStatus) []byte {
	event := domain.PaymentUpdateStatusEvent{
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

	b, err := json.Marshal(event)
	if err != nil {
		log.Printf("ERROR: Failed to marshal updatePaymentStatusCommand event: %v", err)
	}

	return b
}
