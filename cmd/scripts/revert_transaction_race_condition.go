package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

const ucRevertTransactionRaceCondition = "revertTransactionRaceCondition"

func revertTransactionRaceCondition() {
	// 1️⃣ Crear la transacción inicial y capturar su ID
	transactionID := createTransaction()
	fmt.Println("[INFO] Transaction creada con ID:", transactionID)

	var wg sync.WaitGroup

	// 2️⃣ Definir las requests concurrentes
	requests := []func(){
		// PATCH revert transaction (en paralelo varias veces)
		func() {
			url := fmt.Sprintf("http://localhost:8080/transactions/revert/%s", transactionID)
			req, err := http.NewRequest(http.MethodPatch, url, nil)
			if err != nil {
				panic(err)
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			fmt.Println("[PATCH] Status:", resp.Status)
		},
		// POST nueva transacción para forzar conflictos de saldo
		func() {
			url := "http://localhost:8080/transactions"
			body := []byte(`{"origin_account_id":"1","destination_account_id":"2","amount":50}`)
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			fmt.Println("[POST] Status:", resp.Status)
		},
	}

	// 3️⃣ Ejecutar concurrencia
	for i := 0; i < 5; i++ {
		for _, reqFn := range requests {
			wg.Add(1)
			go func(fn func()) {
				defer wg.Done()
				fn()
			}(reqFn)
		}
	}

	wg.Wait()
	fmt.Println("[INFO] Test terminado")
}

// createTransaction hace el POST inicial y devuelve el ID
func createTransaction() string {
	url := "http://localhost:8080/transactions"
	body := []byte(`{"origin_account_id":"1","destination_account_id":"2","amount":100}`)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		panic(err)
	}

	return result.ID
}
