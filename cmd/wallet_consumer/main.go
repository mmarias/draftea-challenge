package wallet_consumer

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/mmarias/golearn/internal/domain"
	"github.com/mmarias/golearn/internal/infraestructure/eventbus"
)

const (
	HoldFunds    = "wallet.hold_funds"
	ReleaseFunds = "wallet.release_funds"
	DebitFunds   = "wallet.debit_funds"
)

func Setup(bus eventbus.Client) {
	dispatcher := func(ctx context.Context, msg []byte) {
		var genericEvent domain.CommandEvent
		if err := json.Unmarshal(msg, &genericEvent); err != nil {
			log.Printf("ERROR: could not unmarshal generic event: %v", err)
			return
		}

		switch genericEvent.EventType {
		case domain.HoldFundsEventType:
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			log.Printf("[Wallet] Holding funds for payment %s", ev.PaymentID)
			time.Sleep(100 * time.Millisecond)
			log.Printf("[Wallet] Funds held for payment %s", ev.PaymentID)

			ev.EventType = HoldFunds
			msgBody, _ := json.Marshal(ev)
			bus.Publish(ctx, HoldFunds, msgBody)

		case domain.ReleaseFundsEventType:
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			log.Printf("[Wallet] Releasing funds for payment %s", ev.PaymentID)
			time.Sleep(100 * time.Millisecond)
			log.Printf("[Wallet] Funds released for payment %s", ev.PaymentID)

			ev.EventType = ReleaseFunds
			msgBody, _ := json.Marshal(ev)
			bus.Publish(ctx, ReleaseFunds, msgBody)

		case domain.DebitFundsEventType:
			var ev domain.WalletCommandEvent
			json.Unmarshal(msg, &ev)
			log.Printf("[Wallet] Debiting funds for payment %s", ev.PaymentID)
			time.Sleep(100 * time.Millisecond)
			log.Printf("[Wallet] Funds debited for payment %s", ev.PaymentID)

			ev.EventType = DebitFunds
			msgBody, _ := json.Marshal(ev)
			bus.Publish(ctx, DebitFunds, msgBody)
		}
	}
	bus.Subscribe(domain.TopicOrchestratorWallet, dispatcher)
}
