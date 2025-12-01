package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	TopicPaymentCreated = "payment.created"
)

type PaymentRepository interface {
	Create(payment Payment) error
}

type Payment struct {
	ID        string
	WalletID  string
	ServiceID string
	Amount    float64
	Currency  string
	Method    string
	Status    PaymentStatus
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "PENDING"
	PaymentStatusFailed    PaymentStatus = "FAILED"
	PaymentStatusCompleted PaymentStatus = "COMPLETED"
)

func (p *Payment) SetID() {
	p.ID = uuid.NewString()
}

func (p *Payment) SetCreatedAt() {
	p.CreatedAt = time.Now().UTC()
}

func (p *Payment) SetUpdatedAt() {
	now := time.Now().UTC()
	p.UpdatedAt = &now
}
