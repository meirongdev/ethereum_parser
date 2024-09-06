package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/meirongdev/ethereum_parser/internal/ethereum"
	"github.com/meirongdev/ethereum_parser/internal/parser"
)

type Response struct {
	Data interface{} `json:"data"`
}

func main() {
	eAPI := ethereum.NewEthereumAPI()
	eParser := parser.NewEthereumParser(eAPI)
	go eParser.Start()

	// get the latest block number
	http.HandleFunc("/currentBlock", func(w http.ResponseWriter, r *http.Request) {
		blockNumber := eParser.GetCurrentBlock()
		response := Response{
			Data: struct {
				BlockNumber int `json:"blockNumber"`
			}{
				BlockNumber: blockNumber,
			},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, "Failed to encode block number", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Query().Get("address")
		if address == "" {
			http.Error(w, "Address is required", http.StatusBadRequest)
			return
		}
		subscribed := eParser.Subscribe(address)
		msg := "Subscribed to address: " + address
		if !subscribed {
			msg = "Already subscribe to address: " + address
		}
		response := Response{
			Data: struct {
				Message string `json:"message"`
			}{
				Message: msg,
			},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Query().Get("address")
		if address == "" {
			http.Error(w, "Address is required", http.StatusBadRequest)
			return
		}
		transactions := eParser.GetTransactions(address)
		response := Response{
			Data: struct {
				Transactions []parser.Transaction `json:"transactions"`
			}{
				Transactions: transactions,
			},
		}
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			http.Error(w, "Failed to encode transactions", http.StatusInternalServerError)
			return
		}
	})

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
