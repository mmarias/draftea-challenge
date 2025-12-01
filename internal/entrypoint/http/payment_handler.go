package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mmarias/golearn/internal/domain"
	"github.com/mmarias/golearn/internal/infraestructure/memcache"
)

type createPaymentImpl interface {
	Execute(ctx context.Context, pay domain.Payment) (string, error)
}

// PaymentHandler holds the dependencies for the handlers.
type PaymentHandler struct {
	createPayment createPaymentImpl
	cache         memcache.Cache
}

func NewPaymentHandler(createPayment createPaymentImpl, cache memcache.Cache) *PaymentHandler {
	return &PaymentHandler{
		createPayment: createPayment,
		cache:         cache,
	}
}

func (h *PaymentHandler) CreatePaymentHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idempotentKey := r.Header.Get("X-Idempotent-Key")

	if idempotentKey == "" {
		http.Error(w, "missing idempotent key", http.StatusBadRequest)
		return
	}

	tx := fmt.Sprintf("payment.%s", idempotentKey)

	if err := h.cache.SetNX(tx); err != nil {
		http.Error(w, "payment already processed or in progress", http.StatusConflict)
		return
	}

	defer h.cache.Delete(tx)

	var req PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.createPayment.Execute(r.Context(), req.ToDomain())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	response := map[string]string{"id": id}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		log.Printf("could not encode response: %v", err)
	}
}
