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

type NotifyUserCommand interface {
	Notify(ctx context.Context, paymentId string, notificationType domain.Notification) error
}

type notifyUserCommand struct {
	publisher publisher.Client
}

func NewNotifyUserCommand(
	publisher publisher.Client,
) *notifyUserCommand {
	return &notifyUserCommand{
		publisher,
	}
}

func (c *notifyUserCommand) Notify(ctx context.Context, paymentId string, notificationType domain.Notification) error {
	traceID := uuid.NewString()

	b := c.buildEventV1(traceID, paymentId, notificationType)

	return retry.Do(
		func() error {
			return c.publisher.Publish(ctx, domain.TopicOrchestratorNotification, b)
		},
		retry.Attempts(3),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(100*time.Millisecond),
	)
}

func (c *notifyUserCommand) buildEventV1(traceID, paymentId string, notificationType domain.Notification) []byte {
	event := domain.NotifyUserEvent{
		CommandEvent: domain.CommandEvent{
			EventType:    domain.NotifyUserEventType,
			EventVersion: "1",
			Timestamp:    time.Now().UTC().Format(time.RFC3339),
			CommandEventMetadata: domain.CommandEventMetadata{
				TraceID:        traceID,
				MessageGroupID: paymentId,
				MessageDeduplicationId: domain.BuildDeduplicationId(
					domain.NotifyUserEventType,
					paymentId,
				),
			},
		},
		NotifyUserEventPayload: domain.NotifyUserEventPayload{
			PaymentID:    paymentId,
			Notification: notificationType,
		},
	}

	b, err := json.Marshal(event)
	if err != nil {
		log.Printf("ERROR: Failed to marshal notifyUserCommand event: %v", err)
	}

	return b
}
