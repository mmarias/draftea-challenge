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

type AuthorizeGatewayCommand interface {
	Authorize(ctx context.Context, paymentId, walletId string, amount float64, currency, token string) error
}

type authorizeGatewayCommand struct {
	publisher  publisher.Client
	repository domain.CommandRepository
}

func NewAuthorizeGatewayCommand(
	publisher publisher.Client,
	repository domain.CommandRepository,
) *authorizeGatewayCommand {
	return &authorizeGatewayCommand{
		publisher,
		repository,
	}
}

func (c *authorizeGatewayCommand) Authorize(ctx context.Context, paymentId, walletId string, amount float64, currency, token string) error {
	traceID := uuid.NewString()

	event := c.buildEventV1(traceID, paymentId, walletId, amount, currency, token)

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
			return c.publisher.Publish(ctx, domain.TopicOrchestratorGateway, b)
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

func (c *authorizeGatewayCommand) buildEventV1(traceID, paymentId, walletId string, amount float64, currency, token string) domain.WalletCommandEvent {
	return domain.WalletCommandEvent{
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
}
