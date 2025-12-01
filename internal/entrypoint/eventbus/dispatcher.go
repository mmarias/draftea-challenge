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

		// Route events
		switch genericEvent.EventType {
		// Events that consume orchestrator from payment service
		case domain.TopicPaymentCreated:
			var ev domain.PaymentUpdateStatusEvent
			json.Unmarshal(msg, &ev)
			handler.HandlePaymentCreated(ctx, ev)
		case domain.TopicPaymentCompleted:
			var ev domain.PaymentUpdateStatusEvent
			json.Unmarshal(msg, &ev)
			handler.HandlePaymentCompleted(ctx, ev)

		// Events that consume orchestrator from gateway service
		case domain.TopicGatewayAuthorized:
			var ev domain.GatewayAuthorizedEvent
			json.Unmarshal(msg, &ev)
			handler.HandleGatewayAuthorized(ctx, ev)
		case domain.TopicGatewayAuthorizationFailed:
			var ev domain.GatewayAuthorizationFailedEvent
			json.Unmarshal(msg, &ev)
			handler.HandleGatewayAuthorizationFailed(ctx, ev)

		// Events that consume orchestrator from wallet service
		case domain.TopicWalletFunds:
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			handler.HandleFundsHeld(ctx, ev)
		case domain.TopicWalletDebitFunds:
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			handler.HandleFundsDebited(ctx, ev)
		case domain.TopicWalletHoldFundsFailed:
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			handler.HandleFundsHoldFailed(ctx, ev)
		case domain.TopicWalletFundsReleased:
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			handler.HandleFundsReleased(ctx, ev)
		}
	}

	// Subscribe the dispatcher to all topics the orchestrator listens to.
	log.Printf("Subscribing to topic: %s", domain.TopicPaymentCreated)
	bus.Subscribe(domain.TopicPaymentCreated, dispatcher)
	log.Printf("Subscribing to topic: %s", domain.TopicOrchestratorWallet)
	bus.Subscribe(domain.TopicOrchestratorWallet, dispatcher)
	log.Printf("Subscribing to topic: %s", domain.TopicGatewayAuthorized)
	bus.Subscribe(domain.TopicGatewayAuthorized, dispatcher)
	log.Printf("Subscribing to topic: %s", domain.TopicGatewayAuthorizationFailed)
	bus.Subscribe(domain.TopicGatewayAuthorizationFailed, dispatcher)
	log.Printf("Subscribing to topic: %s", domain.TopicPaymentCompleted)
	bus.Subscribe(domain.TopicPaymentCompleted, dispatcher)
	log.Printf("Subscribing to topic: %s", domain.TopicWalletFunds)
	bus.Subscribe(domain.TopicWalletFunds, dispatcher)
	log.Printf("Subscribing to topic: %s", domain.TopicWalletDebitFunds)
	bus.Subscribe(domain.TopicWalletDebitFunds, dispatcher)
	log.Printf("Subscribing to topic: %s", domain.TopicWalletHoldFundsFailed)
	bus.Subscribe(domain.TopicWalletHoldFundsFailed, dispatcher)
	log.Printf("Subscribing to topic: %s", domain.TopicWalletFundsReleased)
	bus.Subscribe(domain.TopicWalletFundsReleased, dispatcher)
}
