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

type ReleaseFundsCommand interface {
	Release(ctx context.Context, paymentId, walletId string, amount float64, currency string) error
}

type releaseFundsCommand struct {
	publisher  publisher.Client
	repository domain.CommandRepository
}

func NewReleaseFundsCommand(
	publisher publisher.Client,
	repository domain.CommandRepository,
) *releaseFundsCommand {
	return &releaseFundsCommand{
		publisher,
		repository,
	}
}

func (c *releaseFundsCommand) Release(ctx context.Context, paymentId, walletId string, amount float64, currency string) error {
	traceID := uuid.NewString()

	event := c.buildReleaseEventV1(traceID, paymentId, walletId, amount, currency)

	b, err := json.Marshal(event)
	if err != nil {
		_ = c.repository.Save(
			ctx,
			domain.CommandStatusFailed,
			event.CommandEvent,
			event.WalletCommandEventPayload,
		)
		return err
	}

	err = retry.Do(
		func() error {
			return c.publisher.Publish(ctx, domain.TopicOrchestratorWallet, b)
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
			event.WalletCommandEventPayload,
		)
		return err
	}

	return nil
}

func (c *releaseFundsCommand) buildReleaseEventV1(traceID, paymentId, walletId string, amount float64, currency string) domain.WalletCommandEvent {
	return domain.WalletCommandEvent{
		CommandEvent: domain.CommandEvent{
			EventType:    domain.ReleaseFundsEventType,
			EventVersion: "1",
			Timestamp:    time.Now().UTC().Format(time.RFC3339),
			CommandEventMetadata: domain.CommandEventMetadata{
				TraceID:        traceID,
				MessageGroupID: paymentId,
				MessageDeduplicationId: domain.BuildDeduplicationId(
					domain.ReleaseFundsEventType,
					paymentId,
				),
			},
		},
		WalletCommandEventPayload: domain.WalletCommandEventPayload{
			WalletID:  walletId,
			PaymentID: paymentId,
			Amount:    amount,
			Currency:  currency,
		},
	}
}
