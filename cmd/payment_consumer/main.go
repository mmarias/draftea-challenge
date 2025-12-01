package payment_consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/mmarias/golearn/internal/domain"
	"github.com/mmarias/golearn/internal/infraestructure/eventbus"
)

const (
	PaymentCompleted = "payment.completed"
	PaymentFailed    = "payment.failed"
)

func Setup(bus eventbus.Client) {
	dispatcher := func(ctx context.Context, msg []byte) {
		var genericEvent domain.CommandEvent
		if err := json.Unmarshal(msg, &genericEvent); err != nil {
			log.Printf("ERROR: could not unmarshal generic event: %v", err)
			return
		}

		switch genericEvent.EventType {
		case domain.PaymentUpdateStatusEventType:
			var ev domain.PaymentUpdateStatusEvent
			if err := json.Unmarshal(msg, &ev); err != nil {
				log.Printf("ERROR: could not unmarshal PaymentUpdateStatusEvent: %v", err)
				return
			}

			log.Printf("[PaymentConsumer] Received PaymentUpdateStatusEvent for PaymentID: %s with status: %s", ev.PaymentID, ev.Status)

			switch ev.PaymentUpdateStatusEventPayload.Status {
			case domain.PaymentStatusCompleted:
				log.Printf("[PaymentConsumer] Handling PaymentStatusCompleted for PaymentID: %s", ev.PaymentID)
				time.Sleep(50 * time.Millisecond) // Simulate some work

				// Publish payment.completed event
				ev.EventType = PaymentCompleted
				msgBody, _ := json.Marshal(ev)
				bus.Publish(ctx, PaymentCompleted, msgBody)

			case domain.PaymentStatusFailed:
				log.Printf("[PaymentConsumer] Handling PaymentStatusFailed for PaymentID: %s", ev.PaymentID)
				time.Sleep(50 * time.Millisecond) // Simulate some work

				// Publish payment.failed event
				ev.EventType = PaymentFailed
				msgBody, _ := json.Marshal(ev)
				bus.Publish(ctx, PaymentFailed, msgBody)

			default:
				log.Printf("[PaymentConsumer] Unknown payment status received for PaymentID %s: %s", ev.PaymentID, ev.PaymentUpdateStatusEventPayload.Status)
			}
		default:
			log.Printf("[PaymentConsumer] Unknown event type received: %s", genericEvent.EventType)
		}
	}
	bus.Subscribe(domain.TopicOrchestratorPayment, dispatcher)
}
