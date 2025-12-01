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

func TestPaymentSaga_HappyPath(t *testing.T) {
	// 1. SETUP: Initialize a self-contained environment for the test.
	// This is similar to the dependency injection in main.go.
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
	wg.Add(2) // We expect two final events: status update and user notification

	var receivedStatusUpdate domain.PaymentUpdateStatusEvent
	var receivedUserNotification domain.NotifyUserEvent

	// 2. SPY: Create a subscriber that spies on the final events to assert success.
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

	// 3. SIMULATOR & DISPATCHER: This dispatcher simulates external services and routes events.
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
			// Simulate Gateway Service success
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)

			// The gateway would publish this event upon success
			gatewaySuccessEvent := domain.GatewayAuthorizedEvent{
				PaymentID: ev.PaymentID,
				WalletID:  ev.WalletID,
				Amount:    ev.Amount,
				Currency:  ev.Currency,
			}
			type eventForDispatch struct {
				domain.GatewayAuthorizedEvent
				EventType string `json:"event_type"`
			}
			msgBody, _ := json.Marshal(eventForDispatch{
				GatewayAuthorizedEvent: gatewaySuccessEvent,
				EventType:              "gateway.authorized",
			})
			bus.Publish(ctx, "gateway.authorized", msgBody)

		case "gateway.authorized":
			var ev domain.GatewayAuthorizedEvent
			json.Unmarshal(msg, &ev)
			err := sagaHandler.HandleGatewayAuthorized(ctx, ev)
			assert.NoError(t, err)

		case domain.DebitFundsEventType:
			// Simulate Wallet Service debit success
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			err := sagaHandler.HandleFundsDebited(ctx, ev)
			assert.NoError(t, err)
		}
	}

	bus.Subscribe("payment.created", dispatcher)
	bus.Subscribe(domain.TopicOrchestratorWallet, dispatcher)
	bus.Subscribe(domain.TopicOrchestratorGateway, dispatcher)
	bus.Subscribe("gateway.authorized", dispatcher)

	// 4. EXECUTION: Kick off the saga.
	paymentID := uuid.New().String()
	initialEvent := domain.PaymentCreatedEvent{
		ID:       paymentID,
		WalletID: "wallet-test-123",
		Amount:   150.75,
		Currency: "EUR",
		Token:    "tok_test_visa",
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

	// 5. ASSERTION: Wait for the final events and verify their content.
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All good, the WaitGroup was unlocked.
		assert.Equal(t, paymentID, receivedStatusUpdate.PaymentID)
		assert.Equal(t, domain.PaymentStatusCompleted, receivedStatusUpdate.Status)

		assert.Equal(t, paymentID, receivedUserNotification.PaymentID)
		assert.Equal(t, domain.PaymentSuccess, receivedUserNotification.Notification)
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out: did not receive the final SAGA events.")
	}
}
