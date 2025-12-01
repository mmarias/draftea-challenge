package orchestrator

import (
	"context"
	"errors"
	"testing"

	"github.com/mmarias/golearn/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPublisher is a mock for the publisher.Client interface
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, topic string, body []byte) error {
	args := m.Called(ctx, topic, body)
	return args.Error(0)
}

func TestHoldFundsCommand_Hold(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(publisher *MockPublisher)
		expectedError bool
	}{
		{
			name: "publisher fails after retries",
			setupMocks: func(publisher *MockPublisher) {
				publisher.On("Publish", mock.Anything, domain.TopicOrchestratorWallet, mock.Anything).Return(errors.New("publisher error")).Times(3)
			},
			expectedError: true,
		},
		{
			name: "publisher succeeds",
			setupMocks: func(publisher *MockPublisher) {
				publisher.On("Publish", mock.Anything, domain.TopicOrchestratorWallet, mock.Anything).Return(nil).Once()
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisherMock := new(MockPublisher)
			tt.setupMocks(publisherMock)

			cmd := NewHoldFundsCommand(publisherMock)
			err := cmd.Hold(context.Background(), "payment-123", "wallet-456", 100.0, "USD")

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			publisherMock.AssertExpectations(t)
		})
	}
}
