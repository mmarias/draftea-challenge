package orchestrator

import (
	"context"
	"errors"
	"testing"

	"github.com/mmarias/golearn/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNotifyUserCommand_Notify(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(publisher *MockPublisher)
		expectedError bool
	}{
		{
			name: "publisher fails after retries",
			setupMocks: func(publisher *MockPublisher) {
				publisher.On("Publish", mock.Anything, domain.TopicOrchestratorNotification, mock.Anything).Return(errors.New("publisher error")).Times(3)
			},
			expectedError: true,
		},
		{
			name: "publisher succeeds",
			setupMocks: func(publisher *MockPublisher) {
				publisher.On("Publish", mock.Anything, domain.TopicOrchestratorNotification, mock.Anything).Return(nil).Once()
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisherMock := new(MockPublisher)
			tt.setupMocks(publisherMock)

			cmd := NewNotifyUserCommand(publisherMock)
			err := cmd.Notify(context.Background(), "payment-123", domain.PaymentSuccess)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			publisherMock.AssertExpectations(t)
		})
	}
}
