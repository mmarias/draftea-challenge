package domain

import (
	"context"
)

const (
	TopicOrchestratorWallet       = "orchestrator.wallet"
	TopicOrchestratorPayment      = "orchestrator.payment"
	TopicOrchestratorGateway      = "orchestrator.gateway"
	TopicOrchestratorNotification = "orchestrator.notification"
)

type WalletCommands interface {
	Debit(ctx context.Context, walletId string, amount float64, currency string) error
	Credit(ctx context.Context, walletId string, amount float64, currency string) error
	Release(ctx context.Context, walletId string, amount float64, currency string) error
}

const (
	HoldFundsEventType    = "hold_funds"
	ReleaseFundsEventType = "release_funds"
	DebitFundsEventType   = "debit_funds"
)

const (
	PaymentUpdateStatusEventType = "payment_update_status"
	AuthorizeGatewayEventType    = "authorize_gateway"
	NotifyUserEventType          = "notify_user"
)

// Event from Payment Service
type PaymentCreatedEvent struct {
	ID       string  `json:"id"`
	WalletID string  `json:"wallet_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Token    string  `json:"token"`
}

// Event from Gateway Service
type GatewayAuthorizedEvent struct {
	PaymentID string  `json:"payment_id"`
	WalletID  string  `json:"wallet_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

// Event from Gateway Service on failure
type GatewayAuthorizationFailedEvent struct {
	PaymentID string  `json:"payment_id"`
	WalletID  string  `json:"wallet_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Reason    string  `json:"reason"`
}

type CommandEvent struct {
	EventType            string `json:"event_type"`
	EventVersion         string `json:"event_version"`
	Timestamp            string `json:"timestamp"`
	CommandEventMetadata `json:"metadata"`
}

func BuildDeduplicationId(action, id string) string {
	return action + "." + id
}

type CommandEventMetadata struct {
	TraceID                string `json:"trace_id"`
	MessageGroupID         string `json:"message_group_id"`
	MessageDeduplicationId string `json:"message_deduplication_id"`
}

type WalletCommandEvent struct {
	CommandEvent
	WalletCommandEventPayload `json:"payload"`
}

type WalletCommandEventPayload struct {
	WalletID  string  `json:"wallet_id"`
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Token     string  `json:"token"`
}

type PaymentUpdateStatusEvent struct {
	CommandEvent
	PaymentUpdateStatusEventPayload `json:"payload"`
}

type PaymentUpdateStatusEventPayload struct {
	PaymentID string        `json:"payment_id"`
	Status    PaymentStatus `json:"status"`
}

type NotifyUserEvent struct {
	CommandEvent
	NotifyUserEventPayload `json:"payload"`
}

type NotifyUserEventPayload struct {
	PaymentID    string       `json:"payment_id"`
	Notification Notification `json:"notification"`
}

type MetricEvent struct {
	CommandEvent
	MetricEventPayload `json:"payload"`
}

type MetricEventPayload struct {
	Metric string `json:"metric"`
}

type PaymentCommands interface {
	UpdateStatus(ctx context.Context, paymentId string, status PaymentStatus) error
}

type GatewayCommands interface {
	Authorize(ctx context.Context, paymentId string, amount float64, currency string, token string) error
	Refund(ctx context.Context, paymentId string) error
}

type Notification string

const (
	PaymentFailure Notification = "payment_failure"
	PaymentSuccess Notification = "payment_success"

	RefundFailure Notification = "refund_failure"
	RefundSuccess Notification = "refund_success"
)

type NotifyCommands interface {
	Notify(v Notification)
}
