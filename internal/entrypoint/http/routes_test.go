package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterRoutes(t *testing.T) {
	mux := http.NewServeMux()
	paymentHandler := &PaymentHandler{} // Using a dummy handler

	RegisterRoutes(mux, paymentHandler)

	// Test that the route is registered
	req := httptest.NewRequest(http.MethodPost, "/payments", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	// We expect a call to the handler, which is nil in this dummy, so it will panic.
	// We are not testing the handler here, but that the route exists.
	// A more sophisticated test could involve a mock handler that sets a flag.
	// For now, we'll just check that it doesn't return 404.
	assert.NotEqual(t, http.StatusNotFound, rr.Code, "should have registered the /payments route")
}
