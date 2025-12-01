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

type HoldFundsCommand interface {
	Hold(ctx context.Context, paymentId, walletId string, amount float64, currency string) error
}

type holdFundsCommand struct {
	publisher publisher.Client
}

func NewHoldFundsCommand(
	publisher publisher.Client,
) *holdFundsCommand {
	return &holdFundsCommand{
		publisher,
	}
}

func (c *holdFundsCommand) Hold(ctx context.Context, paymentId, walletId string, amount float64, currency string) error {
	// extracted from context implementation of otel for example
	traceID := uuid.NewString()

	b := c.buildHoldEventV1(traceID, paymentId, walletId, amount, currency)

	return retry.Do(
		func() error {
			return c.publisher.Publish(ctx, domain.TopicOrchestratorWallet, b)
		},
		retry.Attempts(3),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(100*time.Millisecond),
	)
}

func (c *holdFundsCommand) buildHoldEventV1(traceID, paymentId, walletId string, amount float64, currency string) []byte {
	event := domain.WalletCommandEvent{
		CommandEvent: domain.CommandEvent{
			EventType:    domain.HoldFundsEventType,
			EventVersion: "1",
			Timestamp:    time.Now().UTC().Format(time.RFC3339),
			CommandEventMetadata: domain.CommandEventMetadata{
				TraceID:        traceID,
				MessageGroupID: paymentId,
				MessageDeduplicationId: domain.BuildDeduplicationId(
					domain.HoldFundsEventType,
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
		log.Printf("ERROR: Failed to marshal holdFundsCommand event: %v", err)
	}

	return b
}
