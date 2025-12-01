package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type HoldResponse struct {
	ID string `json:"id"`
}

var ucCreateHoldConcurrency = "createHoldConcurrency"
func createHoldConcurrency() {
	accountID := "1" // cuenta inicial con 1000
	baseURL := "http://localhost:8080"

	cycles := 5       // cantidad de ciclos de estrés
	parallelism := 10 // cantidad de goroutines en paralelo por paso

	var mu sync.Mutex
	holdIDs := []string{}

	for cycle := 1; cycle <= cycles; cycle++ {
		fmt.Printf("\n--- CICLO %d ---\n", cycle)

		var wg sync.WaitGroup

		// 1. Crear holds en paralelo
		for i := 0; i < parallelism; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				url := fmt.Sprintf("%s/accounts/%s/holds", baseURL, accountID)
				body := map[string]interface{}{
					"amount": 400,
					"reason": fmt.Sprintf("Hold stress test c%d-%d", cycle, n),
				}
				data, _ := json.Marshal(body)
				resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
				if err != nil {
					fmt.Println("[HOLD ERROR]", err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusCreated {
					var hr HoldResponse
					if err := json.NewDecoder(resp.Body).Decode(&hr); err == nil && hr.ID != "" {
						mu.Lock()
						holdIDs = append(holdIDs, hr.ID)
						mu.Unlock()
						fmt.Println("[HOLD] Created:", hr.ID)
					}
				} else {
					bodyBytes, _ := io.ReadAll(resp.Body)
					fmt.Println("[HOLD FAIL]", resp.Status, string(bodyBytes))
				}
			}(i + 1)
		}
		wg.Wait()

		// 2. Crear transacciones en paralelo
		for i := 0; i < parallelism; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				url := fmt.Sprintf("%s/transactions", baseURL)
				body := map[string]interface{}{
					"origin_account_id":      accountID,
					"destination_account_id": "2",
					"amount":                 500,
				}
				data, _ := json.Marshal(body)
				resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
				if err != nil {
					fmt.Println("[TX ERROR]", err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusCreated {
					fmt.Println("[TX] Success")
				} else {
					bodyBytes, _ := io.ReadAll(resp.Body)
					fmt.Println("[TX FAIL]", resp.Status, string(bodyBytes))
				}
			}(i + 1)
		}
		wg.Wait()

		// 3. Liberar holds en paralelo
		mu.Lock()
		toRelease := holdIDs
		holdIDs = []string{}
		mu.Unlock()

		for _, holdID := range toRelease {
			wg.Add(1)
			go func(id string) {
				defer wg.Done()
				url := fmt.Sprintf("%s/holds/%s/release", baseURL, id)
				req, err := http.NewRequest(http.MethodPatch, url, nil)
				if err != nil {
					fmt.Println("[RELEASE ERROR]", err)
					return
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					fmt.Println("[RELEASE ERROR]", err)
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					fmt.Println("[RELEASE] Success:", id)
				} else {
					bodyBytes, _ := io.ReadAll(resp.Body)
					fmt.Println("[RELEASE FAIL]", resp.Status, string(bodyBytes))
				}
			}(holdID)
		}
		wg.Wait()

		// Pausa breve para que se estabilicen los saldos antes del siguiente ciclo
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Stress test terminado")
}
