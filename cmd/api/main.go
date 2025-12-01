package main

import (
	"log"
	"net/http"
	"time"

	"github.com/mmarias/golearn/internal/app/payment/v1"
	entrypoint "github.com/mmarias/golearn/internal/entrypoint/http"
	"github.com/mmarias/golearn/internal/infraestructure/database"
	"github.com/mmarias/golearn/internal/infraestructure/memcache"
	"github.com/mmarias/golearn/internal/infraestructure/publisher"
)

func main() {
	// load dependencies
	cache := memcache.NewCache(5 * time.Second)
	paymentRepository := database.NewPaymentRepository()
	publisher := publisher.New()

	paymentCreateService := v1.NewCreatePaymentUseCase(paymentRepository, publisher)

	paymentHandler := entrypoint.NewPaymentHandler(paymentCreateService, cache)

	mux := http.NewServeMux()
	entrypoint.RegisterRoutes(mux, paymentHandler)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
