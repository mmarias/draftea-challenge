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
			if err := json.Unmarshal(msg, &ev); err != nil {
				log.Printf("ERROR: could not unmarshal PaymentUpdateStatusEvent for %s: %v", domain.TopicPaymentCreated, err)
				return
			}
			handler.HandlePaymentCreated(ctx, ev)
		case domain.TopicPaymentCompleted:
			var ev domain.PaymentUpdateStatusEvent
			if err := json.Unmarshal(msg, &ev); err != nil {
				log.Printf("ERROR: could not unmarshal PaymentUpdateStatusEvent for %s: %v", domain.TopicPaymentCompleted, err)
				return
			}
			handler.HandlePaymentCompleted(ctx, ev)

		// Events that consume orchestrator from gateway service
		case domain.TopicGatewayAuthorized:
			var ev domain.GatewayAuthorizedEvent
			if err := json.Unmarshal(msg, &ev); err != nil {
				log.Printf("ERROR: could not unmarshal GatewayAuthorizedEvent for %s: %v", domain.TopicGatewayAuthorized, err)
				return
			}
			handler.HandleGatewayAuthorized(ctx, ev)
		case domain.TopicGatewayAuthorizationFailed:
			var ev domain.GatewayAuthorizationFailedEvent
			if err := json.Unmarshal(msg, &ev); err != nil {
				log.Printf("ERROR: could not unmarshal GatewayAuthorizationFailedEvent for %s: %v", domain.TopicGatewayAuthorizationFailed, err)
				return
			}
			handler.HandleGatewayAuthorizationFailed(ctx, ev)

		// Events that consume orchestrator from wallet service
		case domain.TopicWalletFunds:
			var ev domain.WalletCommandEvent
			if err := json.Unmarshal(msg, &ev); err != nil {
				log.Printf("ERROR: could not unmarshal WalletCommandEvent for %s: %v", domain.TopicWalletFunds, err)
				return
			}
			handler.HandleFundsHeld(ctx, ev)
		case domain.TopicWalletDebitFunds:
			var ev domain.WalletCommandEvent
			if err := json.Unmarshal(msg, &ev); err != nil {
				log.Printf("ERROR: could not unmarshal WalletCommandEvent for %s: %v", domain.TopicWalletDebitFunds, err)
				return
			}
			handler.HandleFundsDebited(ctx, ev)
		case domain.TopicWalletHoldFundsFailed:
			var ev domain.WalletCommandEvent
			if err := json.Unmarshal(msg, &ev); err != nil {
				log.Printf("ERROR: could not unmarshal WalletCommandEvent for %s: %v", domain.TopicWalletHoldFundsFailed, err)
				return
			}
			handler.HandleFundsHoldFailed(ctx, ev)
		case domain.TopicWalletFundsReleased:
			var ev domain.WalletCommandEvent
			if err := json.Unmarshal(msg, &ev); err != nil {
				log.Printf("ERROR: could not unmarshal WalletCommandEvent for %s: %v", domain.TopicWalletFundsReleased, err)
				return
			}
			handler.HandleFundsReleased(ctx, ev)
		}
	}

	// Subscribe the dispatcher to all topics the orchestrator listens to.
	bus.Subscribe(domain.TopicPaymentCreated, dispatcher)
	bus.Subscribe(domain.TopicGatewayAuthorized, dispatcher)
	bus.Subscribe(domain.TopicGatewayAuthorizationFailed, dispatcher)
	bus.Subscribe(domain.TopicPaymentCompleted, dispatcher)
	bus.Subscribe(domain.TopicWalletFunds, dispatcher)
	bus.Subscribe(domain.TopicWalletDebitFunds, dispatcher)
	bus.Subscribe(domain.TopicWalletHoldFundsFailed, dispatcher)
	bus.Subscribe(domain.TopicWalletFundsReleased, dispatcher)
}
