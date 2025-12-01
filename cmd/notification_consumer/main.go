package notification_consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/mmarias/golearn/internal/domain"
	"github.com/mmarias/golearn/internal/infraestructure/eventbus"
)

func Setup(bus eventbus.Client) {
	dispatcher := func(ctx context.Context, msg []byte) {
		var genericEvent domain.CommandEvent
		if err := json.Unmarshal(msg, &genericEvent); err != nil {
			log.Printf("ERROR: could not unmarshal generic event: %v", err)
			return
		}

		if genericEvent.EventType == domain.NotifyUserEventType {
			var ev domain.NotifyUserEvent
			json.Unmarshal(msg, &ev)
			log.Printf("[Notification] Sending notification '%s' for payment %s", ev.Notification, ev.PaymentID)
		}
	}
	bus.Subscribe(domain.TopicOrchestratorNotification, dispatcher)
}
