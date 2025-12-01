package http

import (
	"fmt"

	"github.com/mmarias/golearn/internal/domain"
)

type PaymentRequest struct {
	WalletID  string  `json:"wallet_id"`
	ServiceID string  `json:"service_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Method    string  `json:"method"`
}

func (t *PaymentRequest) Validate() error {
	if t.WalletID == "" {
		return fmt.Errorf("missing wallet_id")
	}

	if t.ServiceID == "" {
		return fmt.Errorf("missing service_id")
	}

	if t.Amount <= 0 {
		return fmt.Errorf("invalid amount")
	}

	if t.Currency == "" {
		return fmt.Errorf("missing currency")
	}

	if t.Method == "" {
		return fmt.Errorf("missing method")
	}

	return nil
}

func (t *PaymentRequest) ToDomain() domain.Payment {
	return domain.Payment{
		WalletID:  t.WalletID,
		ServiceID: t.ServiceID,
		Amount:    t.Amount,
		Currency:  t.Currency,
		Method:    t.Method,
	}
}
