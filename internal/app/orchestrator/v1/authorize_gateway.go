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

type AuthorizeGatewayCommand interface {
	Authorize(ctx context.Context, paymentId, walletId string, amount float64, currency, token string) error
}

type authorizeGatewayCommand struct {
	publisher publisher.Client
}

func NewAuthorizeGatewayCommand(
	publisher publisher.Client,
) *authorizeGatewayCommand {
	return &authorizeGatewayCommand{
		publisher,
	}
}

func (c *authorizeGatewayCommand) Authorize(ctx context.Context, paymentId, walletId string, amount float64, currency, token string) error {
	traceID := uuid.NewString()

	b := c.buildEventV1(traceID, paymentId, walletId, amount, currency, token)

	return retry.Do(
		func() error {
			return c.publisher.Publish(ctx, domain.TopicOrchestratorGateway, b)
		},
		retry.Attempts(3),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(100*time.Millisecond),
	)
}

func (c *authorizeGatewayCommand) buildEventV1(traceID, paymentId, walletId string, amount float64, currency, token string) []byte {
	event := domain.WalletCommandEvent{
		CommandEvent: domain.CommandEvent{
			EventType:    domain.AuthorizeGatewayEventType,
			EventVersion: "1",
			Timestamp:    time.Now().UTC().Format(time.RFC3339),
			CommandEventMetadata: domain.CommandEventMetadata{
				TraceID:        traceID,
				MessageGroupID: paymentId,
				MessageDeduplicationId: domain.BuildDeduplicationId(
					domain.AuthorizeGatewayEventType,
					paymentId,
				),
			},
		},
		WalletCommandEventPayload: domain.WalletCommandEventPayload{
			WalletID:  walletId,
			PaymentID: paymentId,
			Amount:    amount,
			Currency:  currency,
			Token:     token,
		},
	}

	b, err := json.Marshal(event)
	if err != nil {
		log.Printf("ERROR: Failed to marshal authorizeGatewayCommand event: %v", err)

	}
	return b
}
