package eventbus

import (
	"context"
	"log"

	orchestrator "github.com/mmarias/golearn/internal/app/orchestrator/v1"
	"github.com/mmarias/golearn/internal/domain"
	"github.com/mmarias/golearn/internal/infraestructure/publisher"
)

// PaymentSagaHandler holds all the commands needed to orchestrate the payment saga.
// It acts as the brain of the saga, with each method handling a specific event.
type PaymentSagaHandler struct {
	holdFundsCmd    orchestrator.HoldFundsCommand
	releaseFundsCmd orchestrator.ReleaseFundsCommand
	debitFundsCmd   orchestrator.DebitFundsCommand
	authorizeCmd    orchestrator.AuthorizeGatewayCommand
	updateStatusCmd orchestrator.UpdatePaymentStatusCommand
	notifyUserCmd   orchestrator.NotifyUserCommand
	publisher       publisher.Client
}

// NewPaymentSagaHandler creates a new handler with all its dependencies.
func NewPaymentSagaHandler(
	holdFundsCmd orchestrator.HoldFundsCommand,
	releaseFundsCmd orchestrator.ReleaseFundsCommand,
	debitFundsCmd orchestrator.DebitFundsCommand,
	authorizeCmd orchestrator.AuthorizeGatewayCommand,
	updateStatusCmd orchestrator.UpdatePaymentStatusCommand,
	notifyUserCmd orchestrator.NotifyUserCommand,

) *PaymentSagaHandler {
	return &PaymentSagaHandler{
		holdFundsCmd:    holdFundsCmd,
		releaseFundsCmd: releaseFundsCmd,
		debitFundsCmd:   debitFundsCmd,
		authorizeCmd:    authorizeCmd,
		updateStatusCmd: updateStatusCmd,
		notifyUserCmd:   notifyUserCmd,
	}
}

// NOTE: In a real implementation, a main consumer (e.g., Kafka consumer) would receive a message,
// unmarshal it, identify the event type, and call the corresponding method from this handler.

// HandlePaymentCreated is triggered by a `payment.created` event from the Payment Service.
// It starts the saga by attempting to hold funds in the user's wallet.
func (h *PaymentSagaHandler) HandlePaymentCreated(ctx context.Context, event domain.PaymentUpdateStatusEvent) error {
	log.Printf("Handling payment.created for PaymentID: %s. Attempting to hold funds.", event.PaymentID)
	// Here you would add logic to handle errors and trigger compensation (though this is the happy path).
	// For this example, we don't have the other payment fields, so we'll use dummy values.
	return h.holdFundsCmd.Hold(ctx, event.PaymentID, "wallet-123", 100, "USD")
}

// HandleFundsHeld is triggered by a `wallet.hold_funds` event from the Wallet Service.
// It continues the saga by requesting payment authorization from the payment gateway.
func (h *PaymentSagaHandler) HandleFundsHeld(ctx context.Context, event domain.WalletCommandEvent) error {
	log.Printf("Handling wallet.hold_funds for PaymentID: %s. Attempting to authorize with gateway.", event.PaymentID)
	// In a real scenario, you would get the token from the original payment.created event,
	// likely by storing saga state in a database (e.g., Redis, PostgreSQL).
	// For this example, we'll pass the token from the event if available.
	return h.authorizeCmd.Authorize(ctx, event.PaymentID, event.WalletID, event.Amount, event.Currency, event.Token)
}

// HandleGatewayAuthorized is triggered by a `gateway.authorized` event from the Payment Gateway.
// It proceeds to debit the previously held funds.
func (h *PaymentSagaHandler) HandleGatewayAuthorized(ctx context.Context, event domain.GatewayAuthorizedEvent) error {
	log.Printf("Handling gateway.authorized for PaymentID: %s. Attempting to debit funds.", event.PaymentID)
	return h.debitFundsCmd.Debit(ctx, event.PaymentID, event.WalletID, event.Amount, event.Currency)
}

// HandleFundsDebited is triggered by a `wallet.debit_funds` event from the Wallet Service.
// This is the final step of the happy path. It marks the payment as complete and notifies the user.
func (h *PaymentSagaHandler) HandleFundsDebited(ctx context.Context, event domain.WalletCommandEvent) error {
	log.Printf("Handling wallet.debit_funds for PaymentID: %s. Finalizing payment.", event.PaymentID)

	// Update payment status to COMPLETED
	err := h.updateStatusCmd.UpdateStatus(ctx, event.PaymentID, domain.PaymentStatusCompleted)
	if err != nil {
		// This is a critical error. The payment succeeded but the status update failed.
		// It requires a retry mechanism or manual intervention.
		log.Printf("CRITICAL: Failed to update payment status for PaymentID %s: %v", event.PaymentID, err)
		return err
	}

	return nil
}

func (h *PaymentSagaHandler) HandlePaymentCompleted(ctx context.Context, event domain.PaymentUpdateStatusEvent) error {
	log.Printf("Handling payment.completed for PaymentID: %s. Notifying user and sending metrics.", event.PaymentID)
	// Notify the user of the successful payment
	err := h.notifyUserCmd.Notify(ctx, event.PaymentID, domain.PaymentSuccess)
	if err != nil {
		// This is a non-critical error for the saga itself, as the payment is already complete.
		// Logging the error is sufficient.
		log.Printf("WARN: Failed to notify user for successful PaymentID %s: %v", event.PaymentID, err)
	}

	log.Printf("Payment SAGA for PaymentID %s completed successfully.", event.PaymentID)
	return nil
}

// HandleFundsHoldFailed is triggered by a `wallet.hold_funds_failed` event.
// It terminates the saga, updates the payment status to FAILED, and notifies the user.
func (h *PaymentSagaHandler) HandleFundsHoldFailed(ctx context.Context, event domain.WalletCommandEvent) error {
	log.Printf("Handling wallet.hold_funds_failed for PaymentID: %s. Terminating saga.", event.PaymentID)

	// Update payment status to FAILED
	err := h.updateStatusCmd.UpdateStatus(ctx, event.PaymentID, domain.PaymentStatusFailed)
	if err != nil {
		// This is a critical error. The payment failed but the status update also failed.
		// It requires a retry mechanism or manual intervention.
		log.Printf("CRITICAL: Failed to update payment status for failed PaymentID %s: %v", event.PaymentID, err)
		return err // Return error to allow retry if the event bus supports it.
	}

	// Notify the user of the failure
	err = h.notifyUserCmd.Notify(ctx, event.PaymentID, domain.PaymentFailure)
	if err != nil {
		// This is a non-critical error for the saga itself, as the payment has already failed.
		// Logging the error is sufficient.
		log.Printf("WARN: Failed to notify user for failed PaymentID %s: %v", event.PaymentID, err)
	}

	log.Printf("Payment SAGA for PaymentID %s failed and was terminated.", event.PaymentID)
	return nil
}

// HandleGatewayAuthorizationFailed is triggered by a `gateway.authorization_failed` event.
// It initiates the compensation process by releasing the previously held funds.
func (h *PaymentSagaHandler) HandleGatewayAuthorizationFailed(ctx context.Context, event domain.GatewayAuthorizationFailedEvent) error {
	log.Printf("Handling gateway.authorization_failed for PaymentID: %s. Releasing funds.", event.PaymentID)
	// Trigger compensation: release the funds that were held.
	return h.releaseFundsCmd.Release(ctx, event.PaymentID, event.WalletID, event.Amount, event.Currency)
}

// HandleFundsReleased is triggered by a `wallet.funds_released` event.
// This is a terminal state for a failed saga. It marks the payment as FAILED and notifies the user.
func (h *PaymentSagaHandler) HandleFundsReleased(ctx context.Context, event domain.WalletCommandEvent) error {
	log.Printf("Handling wallet.funds_released for PaymentID: %s. Terminating saga.", event.PaymentID)

	// Update payment status to FAILED
	err := h.updateStatusCmd.UpdateStatus(ctx, event.PaymentID, domain.PaymentStatusFailed)
	if err != nil {
		log.Printf("CRITICAL: Failed to update payment status for released PaymentID %s: %v", event.PaymentID, err)
		return err
	}

	// Notify the user of the failure
	err = h.notifyUserCmd.Notify(ctx, event.PaymentID, domain.PaymentFailure)
	if err != nil {
		log.Printf("WARN: Failed to notify user for released PaymentID %s: %v", event.PaymentID, err)
	}

	log.Printf("Payment SAGA for PaymentID %s was compensated and terminated.", event.PaymentID)
	return nil
}
