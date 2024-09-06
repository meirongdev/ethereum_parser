package ethereum

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	ethNodeURL = "https://cloudflare-eth.com"
)

type EthereumAPI interface {
	GetCurrentBlock() (string, error)
	GetTransactions(blockNumber string) ([]interface{}, error)
}

type ethereumAPI struct {
}

func NewEthereumAPI() EthereumAPI {
	return &ethereumAPI{}
}

func (e *ethereumAPI) GetCurrentBlock() (string, error) {
	reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":"%s"}`, generateID())
	resp, err := http.Post(ethNodeURL, "application/json", strings.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("StatusCode: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	hexBlock := result["result"].(string)
	return hexBlock, nil
}

func (e *ethereumAPI) GetTransactions(blockNumber string) ([]interface{}, error) {
	// curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x1b4", true],"id":1}'
	var txList []interface{}
	reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["%s", true],"id":"%s"}`, blockNumber, generateID())
	resp, err := http.Post(ethNodeURL, "application/json", strings.NewReader(reqBody))
	if err != nil {
		return txList, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return txList, fmt.Errorf("StatusCode: %s", resp.Status)
	}

	// Parse the JSON-RPC response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return txList, err
	}

	block, ok := result["result"].(map[string]interface{})
	if !ok {
		return txList, fmt.Errorf("invalid block data for block %s", blockNumber)
	}
	transactions, ok := block["transactions"].([]interface{})
	if !ok || len(transactions) == 0 {
		return txList, fmt.Errorf("no transactions found in block %s", blockNumber)
	}
	return transactions, nil
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
