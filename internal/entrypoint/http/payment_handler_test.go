package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mmarias/golearn/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCreatePayment is a mock for the createPaymentImpl interface
type MockCreatePayment struct {
	mock.Mock
}

func (m *MockCreatePayment) Execute(ctx context.Context, pay domain.Payment) (string, error) {
	args := m.Called(ctx, pay)
	return args.String(0), args.Error(1)
}

// MockCache is a mock for the memcache.Cache interface
type MockCache struct {
	mock.Mock
}

func (m *MockCache) SetNX(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *MockCache) Delete(key string) {
	m.Called(key)
}

func TestPaymentHandler_CreatePaymentHandler(t *testing.T) {
	tests := []struct {
		name                 string
		idempotentKey        string
		requestBody          interface{}
		setupMocks           func(createPayment *MockCreatePayment, cache *MockCache)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:                 "missing idempotent key",
			idempotentKey:        "",
			requestBody:          nil,
			setupMocks:           func(createPayment *MockCreatePayment, cache *MockCache) {},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "missing idempotent key\n",
		},
		{
			name:          "payment already processed",
			idempotentKey: "test-key",
			requestBody:   nil,
			setupMocks: func(createPayment *MockCreatePayment, cache *MockCache) {
				cache.On("SetNX", "payment.test-key").Return(errors.New("key already exists"))
			},
			expectedStatusCode:   http.StatusConflict,
			expectedResponseBody: "payment already processed or in progress\n",
		},
		{
			name:          "invalid request body",
			idempotentKey: "test-key",
			requestBody:   "invalid-json",
			setupMocks: func(createPayment *MockCreatePayment, cache *MockCache) {
				cache.On("SetNX", "payment.test-key").Return(nil)
				cache.On("Delete", "payment.test-key").Return()
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: "invalid request body\n",
		},
		{
			name:          "create payment fails",
			idempotentKey: "test-key",
			requestBody: PaymentRequest{
				WalletID:  "wallet-123",
				ServiceID: "service-456",
				Amount:    100,
				Currency:  "USD",
				Method:    "credit_card",
			},
			setupMocks: func(createPayment *MockCreatePayment, cache *MockCache) {
				cache.On("SetNX", "payment.test-key").Return(nil)
				cache.On("Delete", "payment.test-key").Return()
				createPayment.On("Execute", mock.Anything, mock.AnythingOfType("domain.Payment")).Return("", errors.New("internal server error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: "internal server error\n",
		},
		{
			name:          "successful payment creation",
			idempotentKey: "test-key",
			requestBody: PaymentRequest{
				WalletID:  "wallet-123",
				ServiceID: "service-456",
				Amount:    100,
				Currency:  "USD",
				Method:    "credit_card",
			},
			setupMocks: func(createPayment *MockCreatePayment, cache *MockCache) {
				cache.On("SetNX", "payment.test-key").Return(nil)
				createPayment.On("Execute", mock.Anything, mock.AnythingOfType("domain.Payment")).Return("payment-id-123", nil)
				cache.On("Delete", "payment.test-key").Return()
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: "{\"id\":\"payment-id-123\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createPaymentMock := new(MockCreatePayment)
			cacheMock := new(MockCache)
			tt.setupMocks(createPaymentMock, cacheMock)

			handler := NewPaymentHandler(createPaymentMock, cacheMock)

			var body []byte
			if tt.requestBody != nil {
				if s, ok := tt.requestBody.(string); ok {
					body = []byte(s)
				} else {
					body, _ = json.Marshal(tt.requestBody)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewReader(body))
			if tt.idempotentKey != "" {
				req.Header.Set("X-Idempotent-Key", tt.idempotentKey)
			}
			rr := httptest.NewRecorder()

			handler.CreatePaymentHandler(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			assert.Equal(t, tt.expectedResponseBody, rr.Body.String())

			createPaymentMock.AssertExpectations(t)
			cacheMock.AssertExpectations(t)
		})
	}
}
