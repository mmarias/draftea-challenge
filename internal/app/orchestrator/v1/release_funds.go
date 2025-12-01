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

type ReleaseFundsCommand interface {
	Release(ctx context.Context, paymentId, walletId string, amount float64, currency string) error
}

type releaseFundsCommand struct {
	publisher publisher.Client
}

func NewReleaseFundsCommand(
	publisher publisher.Client,
) *releaseFundsCommand {
	return &releaseFundsCommand{
		publisher,
	}
}

func (c *releaseFundsCommand) Release(ctx context.Context, paymentId, walletId string, amount float64, currency string) error {
	traceID := uuid.NewString()

	b := c.buildReleaseEventV1(traceID, paymentId, walletId, amount, currency)

	return retry.Do(
		func() error {
			return c.publisher.Publish(ctx, domain.TopicOrchestratorWallet, b)
		},
		retry.Attempts(3),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(100*time.Millisecond),
	)
}

func (c *releaseFundsCommand) buildReleaseEventV1(traceID, paymentId, walletId string, amount float64, currency string) []byte {
	event := domain.WalletCommandEvent{
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

	b, err := json.Marshal(event)
	if err != nil {
		log.Printf("ERROR: Failed to marshal buildReleaseCommand event: %v", err)
	}

	return b
}
