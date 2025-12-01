package gateway_consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

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

		if genericEvent.EventType == domain.AuthorizeGatewayEventType {
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			log.Printf("[Gateway] Processing authorization for payment %s", ev.PaymentID)

			// Simulate calling an external payment provider
			time.Sleep(200 * time.Millisecond)
			log.Printf("[Gateway] Payment %s authorized by external provider", ev.PaymentID)

			// The gateway would publish this event upon success
			gatewaySuccessEvent := domain.GatewayAuthorizedEvent{
				PaymentID: ev.PaymentID,
				WalletID:  ev.WalletID,
				Amount:    ev.Amount,
				Currency:  ev.Currency,
			}
			type eventForDispatch struct {
				domain.GatewayAuthorizedEvent
				domain.CommandEvent
			}
			msgBody, _ := json.Marshal(eventForDispatch{
				GatewayAuthorizedEvent: gatewaySuccessEvent,
				CommandEvent: domain.CommandEvent{
					EventType: "gateway.authorized",
				},
			})
			bus.Publish(ctx, "gateway.authorized", msgBody)
		}
	}
	bus.Subscribe(domain.TopicOrchestratorGateway, dispatcher)
}
