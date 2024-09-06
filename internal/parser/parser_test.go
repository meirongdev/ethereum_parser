package parser

import (
	"testing"
)

func TestHexToInt(t *testing.T) {
	tests := []struct {
		hexStr   string
		expected int
		hasError bool
	}{
		{"0x1", 1, false},
		{"0xA", 10, false},
		{"0x10", 16, false},
		{"0xFF", 255, false},
		{"0x100", 256, false},
		{"0x7FFFFFFFFFFFFFFF", 9223372036854775807, false},
		{"0x8000000000000000", 0, true}, // value out of range
		{"0xG", 0, true},                // Invalid hex character
		{"", 0, true},                   // Empty string
	}

	for _, test := range tests {
		result, err := hexToInt(test.hexStr)
		if (err != nil) != test.hasError {
			t.Errorf("hexToInt(%s) error = %v, expected error = %v", test.hexStr, err, test.hasError)
		}
		if result != test.expected {
			t.Errorf("hexToInt(%s) = %d, expected %d", test.hexStr, result, test.expected)
		}
	}
}
func TestGetCurrentBlock(t *testing.T) {
	tests := []struct {
		name         string
		currentBlock int
		expected     int
	}{
		{"Initial block", -1, -1},
		{"Block 0", 0, 0},
		{"Block 100", 100, 100},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := &EthereumParser{
				currentBlock: test.currentBlock,
			}
			result := parser.GetCurrentBlock()
			if result != test.expected {
				t.Errorf("GetCurrentBlock() = %d, expected %d", result, test.expected)
			}
		})
	}
}
func TestSubscribe(t *testing.T) {
	tests := []struct {
		name          string
		initialAddrs  map[string]struct{}
		subscribeAddr string
		expected      bool
	}{
		{
			name:          "Subscribe new address",
			initialAddrs:  map[string]struct{}{},
			subscribeAddr: "0x123",
			expected:      true,
		},
		{
			name: "Subscribe existing address",
			initialAddrs: map[string]struct{}{
				"0x123": {},
			},
			subscribeAddr: "0x123",
			expected:      false,
		},
		{
			name: "Subscribe another new address",
			initialAddrs: map[string]struct{}{
				"0x123": {},
			},
			subscribeAddr: "0x456",
			expected:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := &EthereumParser{
				addresses: test.initialAddrs,
			}
			result := parser.Subscribe(test.subscribeAddr)
			if result != test.expected {
				t.Errorf("Subscribe(%s) = %v, expected %v", test.subscribeAddr, result, test.expected)
			}
		})
	}
}
func TestGetTransactions(t *testing.T) {
	tests := []struct {
		name         string
		address      string
		transactions map[string][]Transaction
		expected     []Transaction
	}{
		{
			name:    "No transactions for address",
			address: "0x123",
			transactions: map[string][]Transaction{
				"0x456": {
					{Hash: "0x1", From: "0x456", To: "0x789", Value: 100, BlockNumber: 1},
				},
			},
			expected: []Transaction{},
		},
		{
			name:    "Single transaction for address",
			address: "0x123",
			transactions: map[string][]Transaction{
				"0x123": {
					{Hash: "0x1", From: "0x123", To: "0x456", Value: 100, BlockNumber: 1},
				},
			},
			expected: []Transaction{
				{Hash: "0x1", From: "0x123", To: "0x456", Value: 100, BlockNumber: 1},
			},
		},
		{
			name:    "Multiple transactions for address",
			address: "0x123",
			transactions: map[string][]Transaction{
				"0x123": {
					{Hash: "0x1", From: "0x123", To: "0x456", Value: 100, BlockNumber: 1},
					{Hash: "0x2", From: "0x789", To: "0x123", Value: 200, BlockNumber: 2},
				},
			},
			expected: []Transaction{
				{Hash: "0x1", From: "0x123", To: "0x456", Value: 100, BlockNumber: 1},
				{Hash: "0x2", From: "0x789", To: "0x123", Value: 200, BlockNumber: 2},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := &EthereumParser{
				transactions: test.transactions,
			}
			result := parser.GetTransactions(test.address)
			if len(result) != len(test.expected) {
				t.Errorf("GetTransactions(%s) length = %d, expected %d", test.address, len(result), len(test.expected))
			}
			for i, tx := range result {
				if tx != test.expected[i] {
					t.Errorf("GetTransactions(%s)[%d] = %v, expected %v", test.address, i, tx, test.expected[i])
				}
			}
		})
	}
}
