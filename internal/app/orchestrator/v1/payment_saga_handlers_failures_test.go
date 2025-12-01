
package orchestrator_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	orchestrator "github.com/mmarias/golearn/internal/app/orchestrator/v1"
	"github.com/mmarias/golearn/internal/domain"
	"github.com/mmarias/golearn/internal/infraestructure/eventbus"
	"github.com/mmarias/golearn/mocks"
	"github.com/stretchr/testify/assert"
)

func TestPaymentSaga_InsufficientFunds(t *testing.T) {
	// 1. SETUP
	bus := eventbus.NewMemoryBus()
	repo := mocks.NewCommandRepository()

	holdFundsCmd := orchestrator.NewHoldFundsCommand(bus, repo)
	releaseFundsCmd := orchestrator.NewReleaseFundsCommand(bus, repo)
	debitFundsCmd := orchestrator.NewDebitFundsCommand(bus, repo)
	authorizeCmd := orchestrator.NewAuthorizeGatewayCommand(bus, repo)
	updateStatusCmd := orchestrator.NewUpdatePaymentStatusCommand(bus, repo)
	notifyUserCmd := orchestrator.NewNotifyUserCommand(bus, repo)

	sagaHandler := orchestrator.NewPaymentSagaHandler(
		holdFundsCmd,
		releaseFundsCmd,
		debitFundsCmd,
		authorizeCmd,
		updateStatusCmd,
		notifyUserCmd,
	)

	// --- Test-specific setup ---
	var wg sync.WaitGroup
	wg.Add(2) // Expect two final events: status update and user notification

	var receivedStatusUpdate domain.PaymentUpdateStatusEvent
	var receivedUserNotification domain.NotifyUserEvent

	// 2. SPY: Spy on final events
	spySubscriber := func(ctx context.Context, msg []byte) {
		var genericEvent domain.CommandEvent
		json.Unmarshal(msg, &genericEvent)

		switch genericEvent.EventType {
		case domain.PaymentUpdateStatusEventType:
			json.Unmarshal(msg, &receivedStatusUpdate)
			wg.Done()
		case domain.NotifyUserEventType:
			json.Unmarshal(msg, &receivedUserNotification)
			wg.Done()
		}
	}
	bus.Subscribe(domain.TopicOrchestratorPayment, spySubscriber)
	bus.Subscribe(domain.TopicOrchestratorNotification, spySubscriber)

	// 3. SIMULATOR & DISPATCHER
	dispatcher := func(ctx context.Context, msg []byte) {
		var genericEvent domain.CommandEvent
		json.Unmarshal(msg, &genericEvent)

		switch genericEvent.EventType {
		case "payment.created":
			var ev domain.PaymentCreatedEvent
			json.Unmarshal(msg, &ev)
			err := sagaHandler.HandlePaymentCreated(ctx, ev)
			assert.NoError(t, err)

		case domain.HoldFundsEventType:
			// Simulate Wallet Service failure due to insufficient funds
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)

			// The wallet service would publish this event upon failure
			holdFailedEvent := domain.WalletCommandEvent{
				CommandEvent: domain.CommandEvent{EventType: "wallet.hold_funds_failed"},
				WalletCommandEventPayload: domain.WalletCommandEventPayload{
					PaymentID: ev.PaymentID,
					WalletID:  ev.WalletID,
					Amount:    ev.Amount,
					Currency:  ev.Currency,
				},
			}
			msgBody, _ := json.Marshal(holdFailedEvent)
			bus.Publish(ctx, "wallet.hold_funds_failed", msgBody)

		case "wallet.hold_funds_failed":
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			err := sagaHandler.HandleFundsHoldFailed(ctx, ev)
			assert.NoError(t, err)
		}
	}

	bus.Subscribe("payment.created", dispatcher)
	bus.Subscribe(domain.TopicOrchestratorWallet, dispatcher)
	bus.Subscribe("wallet.hold_funds_failed", dispatcher)

	// 4. EXECUTION
	paymentID := uuid.New().String()
	initialEvent := domain.PaymentCreatedEvent{
		ID:       paymentID,
		WalletID: "wallet-insufficient-funds-123",
		Amount:   200.00,
		Currency: "USD",
	}
	type eventForDispatch struct {
		domain.PaymentCreatedEvent
		EventType string `json:"event_type"`
	}
	msgBody, _ := json.Marshal(eventForDispatch{
		PaymentCreatedEvent: initialEvent,
		EventType:           "payment.created",
	})
	bus.Publish(context.Background(), "payment.created", msgBody)

	// 5. ASSERTION
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		assert.Equal(t, paymentID, receivedStatusUpdate.PaymentID)
		assert.Equal(t, domain.PaymentStatusFailed, receivedStatusUpdate.Status)

		assert.Equal(t, paymentID, receivedUserNotification.PaymentID)
		assert.Equal(t, domain.PaymentFailure, receivedUserNotification.Notification)
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out: did not receive the final SAGA events for failure.")
	}
}

func TestPaymentSaga_GatewayTimeout(t *testing.T) {
	// 1. SETUP
	bus := eventbus.NewMemoryBus()
	repo := mocks.NewCommandRepository()

	holdFundsCmd := orchestrator.NewHoldFundsCommand(bus, repo)
	releaseFundsCmd := orchestrator.NewReleaseFundsCommand(bus, repo)
	debitFundsCmd := orchestrator.NewDebitFundsCommand(bus, repo)
	authorizeCmd := orchestrator.NewAuthorizeGatewayCommand(bus, repo)
	updateStatusCmd := orchestrator.NewUpdatePaymentStatusCommand(bus, repo)
	notifyUserCmd := orchestrator.NewNotifyUserCommand(bus, repo)

	sagaHandler := orchestrator.NewPaymentSagaHandler(
		holdFundsCmd,
		releaseFundsCmd,
		debitFundsCmd,
		authorizeCmd,
		updateStatusCmd,
		notifyUserCmd,
	)

	// --- Test-specific setup ---
	var wg sync.WaitGroup
	wg.Add(2) // Expect two final events: status update and user notification

	var receivedStatusUpdate domain.PaymentUpdateStatusEvent
	var receivedUserNotification domain.NotifyUserEvent

	// 2. SPY: Spy on final events
	spySubscriber := func(ctx context.Context, msg []byte) {
		var genericEvent domain.CommandEvent
		json.Unmarshal(msg, &genericEvent)

		switch genericEvent.EventType {
		case domain.PaymentUpdateStatusEventType:
			json.Unmarshal(msg, &receivedStatusUpdate)
			wg.Done()
		case domain.NotifyUserEventType:
			json.Unmarshal(msg, &receivedUserNotification)
			wg.Done()
		}
	}
	bus.Subscribe(domain.TopicOrchestratorPayment, spySubscriber)
	bus.Subscribe(domain.TopicOrchestratorNotification, spySubscriber)

	// 3. SIMULATOR & DISPATCHER
	dispatcher := func(ctx context.Context, msg []byte) {
		var genericEvent domain.CommandEvent
		json.Unmarshal(msg, &genericEvent)

		switch genericEvent.EventType {
		case "payment.created":
			var ev domain.PaymentCreatedEvent
			json.Unmarshal(msg, &ev)
			err := sagaHandler.HandlePaymentCreated(ctx, ev)
			assert.NoError(t, err)

		case domain.HoldFundsEventType:
			// Simulate Wallet Service success
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			err := sagaHandler.HandleFundsHeld(ctx, ev)
			assert.NoError(t, err)

		case domain.AuthorizeGatewayEventType:
			// Simulate Gateway failure
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)

			gatewayFailureEvent := domain.GatewayAuthorizationFailedEvent{
				PaymentID: ev.PaymentID,
				WalletID:  ev.WalletID,
				Amount:    ev.Amount,
				Currency:  ev.Currency,
				Reason:    "Gateway timed out",
			}
			type eventForDispatch struct {
				domain.GatewayAuthorizationFailedEvent
				EventType string `json:"event_type"`
			}
			msgBody, _ := json.Marshal(eventForDispatch{
				GatewayAuthorizationFailedEvent: gatewayFailureEvent,
				EventType:                       "gateway.authorization_failed",
			})
			bus.Publish(ctx, "gateway.authorization_failed", msgBody)

		case "gateway.authorization_failed":
			var ev domain.GatewayAuthorizationFailedEvent
			json.Unmarshal(msg, &ev)
			err := sagaHandler.HandleGatewayAuthorizationFailed(ctx, ev)
			assert.NoError(t, err)

		case domain.ReleaseFundsEventType:
			// Simulate Wallet Service success
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)

			fundsReleasedEvent := domain.WalletCommandEvent{
				CommandEvent:              domain.CommandEvent{EventType: "wallet.funds_released"},
				WalletCommandEventPayload: ev.WalletCommandEventPayload,
			}
			msgBody, _ := json.Marshal(fundsReleasedEvent)
			bus.Publish(ctx, "wallet.funds_released", msgBody)

		case "wallet.funds_released":
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			err := sagaHandler.HandleFundsReleased(ctx, ev)
			assert.NoError(t, err)
		}
	}

	bus.Subscribe("payment.created", dispatcher)
	bus.Subscribe(domain.TopicOrchestratorWallet, dispatcher)
	bus.Subscribe(domain.TopicOrchestratorGateway, dispatcher)
	bus.Subscribe("gateway.authorization_failed", dispatcher)
	bus.Subscribe("wallet.funds_released", dispatcher)

	// 4. EXECUTION
	paymentID := uuid.New().String()
	initialEvent := domain.PaymentCreatedEvent{
		ID:       paymentID,
		WalletID: "wallet-gateway-timeout-456",
		Amount:   50.00,
		Currency: "EUR",
		Token:    "tok_mastercard_timeout",
	}
	type eventForDispatch struct {
		domain.PaymentCreatedEvent
		EventType string `json:"event_type"`
	}
	msgBody, _ := json.Marshal(eventForDispatch{
		PaymentCreatedEvent: initialEvent,
		EventType:           "payment.created",
	})
	bus.Publish(context.Background(), "payment.created", msgBody)

	// 5. ASSERTION
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		assert.Equal(t, paymentID, receivedStatusUpdate.PaymentID)
		assert.Equal(t, domain.PaymentStatusFailed, receivedStatusUpdate.Status)

		assert.Equal(t, paymentID, receivedUserNotification.PaymentID)
		assert.Equal(t, domain.PaymentFailure, receivedUserNotification.Notification)
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out: did not receive the final SAGA events for gateway failure.")
	}
}
