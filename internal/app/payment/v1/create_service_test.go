package v1

import (
	"context"
	"testing"

	"github.com/mmarias/golearn/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPaymentRepository struct {
	mock.Mock
}

func (m *mockPaymentRepository) Create(payment domain.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

type mockPublisher struct {
	mock.Mock
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, msg []byte) error {
	args := m.Called(ctx, topic, msg)
	return args.Error(0)
}

func TestCreatePaymentUseCase_Execute(t *testing.T) {
	mockRepo := new(mockPaymentRepository)
	mockPub := new(mockPublisher)

	uc := NewCreatePaymentUseCase(mockRepo, mockPub)

	payment := domain.Payment{
		Amount:   100,
		WalletID: "user-123",
	}

	mockRepo.On("Create", mock.AnythingOfType("domain.Payment")).Return(nil)
	mockPub.On("Publish", context.Background(), domain.TopicPaymentCreated, mock.AnythingOfType("[]uint8")).Return(nil)

	id, err := uc.Execute(context.Background(), payment)

	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}
