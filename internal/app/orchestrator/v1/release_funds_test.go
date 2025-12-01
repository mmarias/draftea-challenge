package orchestrator

import (
	"context"
	"errors"
	"testing"

	"github.com/mmarias/golearn/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReleaseFundsCommand_Release(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(publisher *MockPublisher, repository *MockCommandRepository)
		expectedError bool
	}{
		{
			name: "publisher fails after retries",
			setupMocks: func(publisher *MockPublisher, repository *MockCommandRepository) {
				publisher.On("Publish", mock.Anything, domain.TopicOrchestratorWallet, mock.Anything).Return(errors.New("publisher error")).Times(3)
				repository.On("Save", mock.Anything, domain.CommandStatusFailed, mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: true,
		},
		{
			name: "publisher succeeds",
			setupMocks: func(publisher *MockPublisher, repository *MockCommandRepository) {
				publisher.On("Publish", mock.Anything, domain.TopicOrchestratorWallet, mock.Anything).Return(nil).Once()
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisherMock := new(MockPublisher)
			repositoryMock := new(MockCommandRepository)
			tt.setupMocks(publisherMock, repositoryMock)

			cmd := NewReleaseFundsCommand(publisherMock, repositoryMock)
			err := cmd.Release(context.Background(), "payment-123", "wallet-456", 100.0, "USD")

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			publisherMock.AssertExpectations(t)
			repositoryMock.AssertExpectations(t)
		})
	}
}
