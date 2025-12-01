package database

import (
	"github.com/mmarias/golearn/internal/domain"
)

type paymentRepository struct{}

func NewPaymentRepository() *paymentRepository {
	return &paymentRepository{}
}

func (r *paymentRepository) Create(t domain.Payment) error {
	return nil
}
