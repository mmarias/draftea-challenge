package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

type Transaction struct {
	ID                   string  `json:"id"`
	OriginAccountID      string  `json:"origin_account_id"`
	DestinationAccountID string  `json:"destination_account_id"`
	Amount               float64 `json:"amount"`
}

const ucCreateTransactionRaceCondition = "createTransactionRaceCondition"

func createTransactionRaceCondition() {
	var wg sync.WaitGroup
	requests := []string{
		`{"origin_account_id":"2","destination_account_id":"2","amount":500}`,
		`{"origin_account_id":"2","destination_account_id":"2","amount":500}`,
	}

	for _, reqBody := range requests {
		wg.Add(1)
		go func(body string) {
			defer wg.Done()
			resp, err := http.Post("http://localhost:8080/transactions", "application/json", bytes.NewBuffer([]byte(body)))
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()

			res, err := io.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			var transaction Transaction
			if err = json.Unmarshal(res, &transaction); err != nil {
				panic(err)
			}

			println("Status:", resp.Status, "transaction_id:", transaction.ID)
		}(reqBody)
	}

	wg.Wait()
}
