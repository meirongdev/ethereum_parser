package ethereum

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetCurrentBlock(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   string
		mockStatusCode int
		expectedBlock  string
		expectError    bool
	}{
		{
			name:           "Successful response",
			mockResponse:   `{"jsonrpc":"2.0","result":"0x1b4","id":"1"}`,
			mockStatusCode: http.StatusOK,
			expectedBlock:  "0x1b4",
			expectError:    false,
		},
		{
			name:           "Error response from server",
			mockResponse:   `{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error"}}`,
			mockStatusCode: http.StatusInternalServerError,
			expectedBlock:  "",
			expectError:    true,
		},
		{
			name:           "Invalid JSON response",
			mockResponse:   `{"jsonrpc":"2.0","result":`,
			mockStatusCode: http.StatusOK,
			expectedBlock:  "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatusCode)
				fmt.Fprintln(w, tt.mockResponse)
			}))
			defer server.Close()

			// Override the ethNodeURL with the mock server URL
			ethNodeURL = server.URL

			api := NewEthereumAPI()
			block, err := api.GetCurrentBlock()

			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}

			if block != tt.expectedBlock {
				t.Errorf("expected block: %s, got: %s", tt.expectedBlock, block)
			}
		})
	}
}

// test GetTransactions with the file in the testdata/eth_getblockbynumber.json
func TestGetTransactionsWithResponseJsonFile(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/eth_getblockbynumber.json")
	}))
	defer server.Close()

	// Override the ethNodeURL with the mock server URL
	ethNodeURL = server.URL

	api := NewEthereumAPI()
	txs, err := api.GetTransactions("0x13bb16e")

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(txs) != 134 {
		t.Errorf("expected 134 transaction, got %d", len(txs))
	}
}

func TestGetTransactions(t *testing.T) {
	tests := []struct {
		name           string
		blockNumber    string
		mockResponse   string
		mockStatusCode int
		expectedTxs    int
		expectError    bool
	}{
		{
			name:           "Successful response with transactions",
			blockNumber:    "0x1b4",
			mockResponse:   `{"jsonrpc":"2.0","result":{"transactions":[{}, {}, {}]},"id":"1"}`,
			mockStatusCode: http.StatusOK,
			expectedTxs:    3,
			expectError:    false,
		},
		{
			name:           "Successful response with no transactions",
			blockNumber:    "0x1b4",
			mockResponse:   `{"jsonrpc":"2.0","result":{"transactions":[]},"id":"1"}`,
			mockStatusCode: http.StatusOK,
			expectedTxs:    0,
			expectError:    true,
		},
		{
			name:           "Error response from server",
			blockNumber:    "0x1b4",
			mockResponse:   `{"jsonrpc":"2.0","error":{"code":-32603,"message":"Internal error"}}`,
			mockStatusCode: http.StatusInternalServerError,
			expectedTxs:    0,
			expectError:    true,
		},
		{
			name:           "Invalid JSON response",
			blockNumber:    "0x1b4",
			mockResponse:   `{"jsonrpc":"2.0","result":`,
			mockStatusCode: http.StatusOK,
			expectedTxs:    0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatusCode)
				fmt.Fprintln(w, tt.mockResponse)
			}))
			defer server.Close()

			// Override the ethNodeURL with the mock server URL
			ethNodeURL = server.URL

			api := NewEthereumAPI()
			txs, err := api.GetTransactions(tt.blockNumber)

			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}

			if len(txs) != tt.expectedTxs {
				t.Errorf("expected %d transactions, got %d", tt.expectedTxs, len(txs))
			}
		})
	}
}
