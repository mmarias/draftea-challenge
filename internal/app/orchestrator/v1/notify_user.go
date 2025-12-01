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

type NotifyUserCommand interface {
	Notify(ctx context.Context, paymentId string, notificationType domain.Notification) error
}

type notifyUserCommand struct {
	publisher  publisher.Client
	repository domain.CommandRepository
}

func NewNotifyUserCommand(
	publisher publisher.Client,
	repository domain.CommandRepository,
) *notifyUserCommand {
	return &notifyUserCommand{
		publisher,
		repository,
	}
}

func (c *notifyUserCommand) Notify(ctx context.Context, paymentId string, notificationType domain.Notification) error {
	traceID := uuid.NewString()

	event := c.buildEventV1(traceID, paymentId, notificationType)

	b, err := json.Marshal(event)
	if err != nil {
		_ = c.repository.Save(
			ctx,
			domain.CommandStatusFailed,
			event.CommandEvent,
			event.NotifyUserEventPayload,
		)
		return err
	}

	err = retry.Do(
		func() error {
			return c.publisher.Publish(ctx, domain.TopicOrchestratorNotification, b)
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
			event.NotifyUserEventPayload,
		)
		return err
	}

	return nil
}

func (c *notifyUserCommand) buildEventV1(traceID, paymentId string, notificationType domain.Notification) domain.NotifyUserEvent {
	return domain.NotifyUserEvent{
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
}
