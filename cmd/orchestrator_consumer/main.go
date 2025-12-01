package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mmarias/golearn/internal/app/orchestrator/v1"
	"github.com/mmarias/golearn/internal/domain"
	"github.com/mmarias/golearn/internal/infraestructure/eventbus"
	"github.com/mmarias/golearn/mocks"
)

func main() {
	// 1. Initialize all dependencies (Dependency Injection)
	sagaHandler, bus := wireDependencies()

	// 2. Setup the dispatcher to route events to the correct handlers
	setupSagaDispatcher(bus, sagaHandler)

	// 3. Simulate the start of the SAGA by publishing a `payment.created` event
	log.Println("--- Kicking off a new Payment SAGA ---")
	startSaga(bus)

	// Wait for a moment to allow events to be processed
	time.Sleep(5 * time.Second)
	log.Println("--- Simulation finished ---")
}

// wireDependencies sets up all the application's components and their dependencies.
func wireDependencies() (*orchestrator.PaymentSagaHandler, *eventbus.MemoryBus) {
	// Infrastructure
	bus := eventbus.NewMemoryBus()
	repo := mocks.NewCommandRepository()

	// Commands (the producers)
	holdFundsCmd := orchestrator.NewHoldFundsCommand(bus, repo)
	releaseFundsCmd := orchestrator.NewReleaseFundsCommand(bus, repo)
	debitFundsCmd := orchestrator.NewDebitFundsCommand(bus, repo)
	authorizeCmd := orchestrator.NewAuthorizeGatewayCommand(bus, repo)
	updateStatusCmd := orchestrator.NewUpdatePaymentStatusCommand(bus, repo)
	notifyUserCmd := orchestrator.NewNotifyUserCommand(bus, repo)

	// Saga Handler (the consumer logic)
	sagaHandler := orchestrator.NewPaymentSagaHandler(
		holdFundsCmd,
		releaseFundsCmd,
		debitFundsCmd,
		authorizeCmd,
		updateStatusCmd,
		notifyUserCmd,
	)

	return sagaHandler, bus
}

// setupSagaDispatcher subscribes a generic event dispatcher to all relevant topics.
func setupSagaDispatcher(bus *eventbus.MemoryBus, handler *orchestrator.PaymentSagaHandler) {
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
		case "payment.created": // This is a simulated event type
			var ev domain.PaymentCreatedEvent
			json.Unmarshal(msg, &ev)
			handler.HandlePaymentCreated(ctx, ev)
		case "gateway.authorized": // This is a simulated event type
			var ev domain.GatewayAuthorizedEvent
			json.Unmarshal(msg, &ev)
			handler.HandleGatewayAuthorized(ctx, ev)

		// Events from our own orchestrator commands (e.g., from wallet)
		case domain.HoldFundsEventType:
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			// SIMULATION: We fake the "wallet.hold_funds" success event here.
			// In a real system, the Wallet service would publish this.
			log.Printf("[SIMULATION] Wallet service would process '%s' and publish success.", domain.HoldFundsEventType)
			// We immediately call the next handler in the happy path.
			handler.HandleFundsHeld(ctx, ev)

		case domain.DebitFundsEventType:
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			// SIMULATION: We fake the "wallet.debit" success event here.
			log.Printf("[SIMULATION] Wallet service would process '%s' and publish success.", domain.DebitFundsEventType)
			handler.HandleFundsDebited(ctx, ev)

		default:
			log.Printf("WARN: Unknown event type received: %s", genericEvent.EventType)
		}
	}

	// Subscribe the dispatcher to all topics the orchestrator listens to.
	bus.Subscribe("payment.created", dispatcher)
	bus.Subscribe(domain.TopicOrchestratorWallet, dispatcher)
	bus.Subscribe("gateway.authorized", dispatcher) // This would be a real topic from the gateway
}

// startSaga simulates an external service (like the API) publishing the first event.
func startSaga(bus *eventbus.MemoryBus) {
	// This event would typically come from your API service
	paymentCreatedEvent := domain.PaymentCreatedEvent{
		ID:       uuid.New().String(),
		WalletID: "wallet-123",
		Amount:   99.99,
		Currency: "USD",
		Token:    "tok_mastercard_1234",
	}
	
	// We manually add the "event_type" for the dispatcher to work
	type eventForDispatch struct {
		domain.PaymentCreatedEvent
		EventType string `json:"event_type"`
	}

	msgBody, _ := json.Marshal(eventForDispatch{
		PaymentCreatedEvent: paymentCreatedEvent,
		EventType: "payment.created",
	})

	bus.Publish(context.Background(), "payment.created", msgBody)
}
