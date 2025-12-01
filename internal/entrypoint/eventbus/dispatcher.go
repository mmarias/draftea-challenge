package eventbus

import (
	"context"
	"encoding/json"
	"log"

	"github.com/mmarias/golearn/internal/domain"
	"github.com/mmarias/golearn/internal/infraestructure/eventbus"
)

func SetupSagaDispatcher(bus eventbus.Client, handler *OrchestratorSagaHandler) {
	// The dispatcher is a single function that knows how to route events.
	dispatcher := func(ctx context.Context, msg []byte) {
		var genericEvent domain.CommandEvent
		if err := json.Unmarshal(msg, &genericEvent); err != nil {
			log.Printf("ERROR: could not unmarshal generic event: %v", err)
			return
		}

		// Route event to the correct handler based on EventType
		switch genericEvent.EventType {
		// Events consumed from other services
		case domain.TopicPaymentCreated:
			var ev domain.PaymentUpdateStatusEvent
			json.Unmarshal(msg, &ev)
			handler.HandlePaymentCreated(ctx, ev)
		case "gateway.authorized":
			var ev domain.GatewayAuthorizedEvent
			json.Unmarshal(msg, &ev)
			handler.HandleGatewayAuthorized(ctx, ev)
		case "gateway.authorization_failed":
			var ev domain.GatewayAuthorizationFailedEvent
			json.Unmarshal(msg, &ev)
			handler.HandleGatewayAuthorizationFailed(ctx, ev)

		// Events from our own orchestrator commands (e.g., from wallet)
		case "wallet.hold_funds":
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			handler.HandleFundsHeld(ctx, ev)
		case "wallet.debit_funds":
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			handler.HandleFundsDebited(ctx, ev)
		case "wallet.hold_funds_failed":
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			handler.HandleFundsHoldFailed(ctx, ev)
		case "wallet.funds_released":
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			handler.HandleFundsReleased(ctx, ev)

		case domain.TopicPaymentCompleted:
			var ev domain.PaymentUpdateStatusEvent
			json.Unmarshal(msg, &ev)
			handler.HandlePaymentCompleted(ctx, ev)
		}
	}

	// Subscribe the dispatcher to all topics the orchestrator listens to.
	bus.Subscribe(domain.TopicPaymentCreated, dispatcher)
	bus.Subscribe(domain.TopicOrchestratorWallet, dispatcher)
	bus.Subscribe("gateway.authorized", dispatcher)
	bus.Subscribe("gateway.authorization_failed", dispatcher)
	bus.Subscribe(domain.TopicPaymentCompleted, dispatcher)
	bus.Subscribe("wallet.hold_funds", dispatcher)
	bus.Subscribe("wallet.debit_funds", dispatcher)
	bus.Subscribe("wallet.hold_funds_failed", dispatcher)
	bus.Subscribe("wallet.funds_released", dispatcher)
}
