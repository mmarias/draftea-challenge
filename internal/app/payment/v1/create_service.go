package v1

import (
	"context"
	"encoding/json"
	"time"

	"github.com/avast/retry-go"
	"github.com/google/uuid"
	"github.com/mmarias/golearn/internal/domain"
	"github.com/mmarias/golearn/internal/infraestructure/publisher"
)

type createPaymentUseCase struct {
	repository domain.PaymentRepository
	publisher  publisher.Client
}

func NewCreatePaymentUseCase(
	repository domain.PaymentRepository,
	publisher publisher.Client,
) *createPaymentUseCase {
	return &createPaymentUseCase{
		repository,
		publisher,
	}
}

func (uc *createPaymentUseCase) Execute(ctx context.Context, pay domain.Payment) (string, error) {
	traceID := uuid.NewString() // extracted from context implementation of otel for example

	pay.SetID()
	pay.SetCreatedAt()
	pay.SetStatus(domain.PaymentStatusPending)

	err := uc.repository.Create(pay)
	if err != nil {
		return "", err
	}

	event := uc.buildEventV1(traceID, pay.ID, pay.Status)

	b, err := json.Marshal(event)
	if err != nil {
		//publish metric errors
	}

	err = retry.Do(
		func() error {
			return uc.publisher.Publish(ctx, domain.TopicPaymentCreated, b)
		},
		retry.Attempts(3),
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(100*time.Millisecond),
	)

	if err != nil {
		// publish metric error
	}

	return pay.ID, nil
}

func (c *createPaymentUseCase) buildEventV1(traceID, paymentId string, status domain.PaymentStatus) domain.PaymentUpdateStatusEvent {
	return domain.PaymentUpdateStatusEvent{
		CommandEvent: domain.CommandEvent{
			EventType:    domain.TopicPaymentCreated,
			EventVersion: "1",
			Timestamp:    time.Now().UTC().Format(time.RFC3339),
			CommandEventMetadata: domain.CommandEventMetadata{
				TraceID:        traceID,
				MessageGroupID: paymentId,
				MessageDeduplicationId: domain.BuildDeduplicationId(
					domain.TopicPaymentCreated,
					paymentId,
				),
			},
		},
		PaymentUpdateStatusEventPayload: domain.PaymentUpdateStatusEventPayload{
			PaymentID: paymentId,
			Status:    status,
		},
	}
}
