package main

import (
	"log"
	"net/http"
	"time"

	"github.com/mmarias/golearn/cmd/gateway_consumer"
	"github.com/mmarias/golearn/cmd/notification_consumer"
	"github.com/mmarias/golearn/cmd/orchestrator_consumer"
	"github.com/mmarias/golearn/cmd/payment_consumer"
	"github.com/mmarias/golearn/cmd/wallet_consumer"
	v1 "github.com/mmarias/golearn/internal/app/payment/v1"
	entrypoint "github.com/mmarias/golearn/internal/entrypoint/http"
	"github.com/mmarias/golearn/internal/infraestructure/database"
	"github.com/mmarias/golearn/internal/infraestructure/eventbus"
	"github.com/mmarias/golearn/internal/infraestructure/memcache"
	"github.com/mmarias/golearn/internal/infraestructure/publisher"
)

func main() {
	// load dependencies
	cache := memcache.NewCache(5 * time.Second)
	paymentRepository := database.NewPaymentRepository()

	bus := eventbus.New()
	publisher := publisher.New(bus)

	orchestrator_consumer.Setup(bus)
	gateway_consumer.Setup(bus)
	notification_consumer.Setup(bus)
	wallet_consumer.Setup(bus)
	payment_consumer.Setup(bus)

	paymentCreateService := v1.NewCreatePaymentUseCase(paymentRepository, publisher)

	paymentHandler := entrypoint.NewPaymentHandler(paymentCreateService, cache)

	mux := http.NewServeMux()
	entrypoint.RegisterRoutes(mux, paymentHandler)

	log.Println("Starting server on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
