package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/meirongdev/ethereum_parser/internal/ethereum"
	"github.com/meirongdev/ethereum_parser/internal/parser"
)

type Response struct {
	Data interface{} `json:"data"`
}

func main() {
	var wg sync.WaitGroup

	eAPI := ethereum.NewEthereumAPI()
	eParser := parser.NewEthereumParser(eAPI)
	go eParser.Start()

	mux := http.NewServeMux()

	// get the latest block number
	mux.HandleFunc("/currentBlock", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
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

	closeCh := make(chan struct{})
	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	wg.Add(1)
	server.RegisterOnShutdown(func() {
		defer wg.Done()
		eParser.Stop()
	})
	log.Println("Server started at :8080")
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
		<-sigChan
		log.Println("Received stop signal, shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
		close(closeCh)
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	<-closeCh
	log.Println("Server shutdown finished")
	wg.Wait()
	log.Println("Clear all resources")
}
