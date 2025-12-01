package http

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, paymentHandler *PaymentHandler) {
	mux.HandleFunc("POST /payments", paymentHandler.CreatePaymentHandler)
}
