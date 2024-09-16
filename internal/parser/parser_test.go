package parser

import (
	"fmt"
	"testing"

	"github.com/meirongdev/ethereum_parser/internal/ethereum/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		addresses    map[string]struct{}
		transactions map[string][]Transaction
		queryAddress string
		expected     []Transaction
	}{
		{
			name: "Address not subscribed",
			addresses: map[string]struct{}{
				"0x123": {},
			},
			transactions: map[string][]Transaction{
				"0x123": {
					{Hash: "0xabc", From: "0x123", To: "0x456", Value: "100", BlockNumber: 1},
				},
			},
			queryAddress: "0x789",
			expected:     []Transaction{},
		},
		{
			name: "Address subscribed with transactions",
			addresses: map[string]struct{}{
				"0x123": {},
			},
			transactions: map[string][]Transaction{
				"0x123": {
					{Hash: "0xabc", From: "0x123", To: "0x456", Value: "100", BlockNumber: 1},
					{Hash: "0xdef", From: "0x123", To: "0x789", Value: "200", BlockNumber: 2},
				},
			},
			queryAddress: "0x123",
			expected: []Transaction{
				{Hash: "0xabc", From: "0x123", To: "0x456", Value: "100", BlockNumber: 1},
				{Hash: "0xdef", From: "0x123", To: "0x789", Value: "200", BlockNumber: 2},
			},
		},
		{
			name: "Address subscribed with no transactions",
			addresses: map[string]struct{}{
				"0x123": {},
			},
			transactions: map[string][]Transaction{},
			queryAddress: "0x123",
			expected:     []Transaction{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := &EthereumParser{
				addresses:    test.addresses,
				transactions: test.transactions,
			}
			result := parser.GetTransactions(test.queryAddress)
			if len(result) != len(test.expected) {
				t.Errorf("GetTransactions(%s) = %v, expected %v", test.queryAddress, result, test.expected)
			}
			for i, tx := range result {
				if tx != test.expected[i] {
					t.Errorf("GetTransactions(%s)[%d] = %v, expected %v", test.queryAddress, i, tx, test.expected[i])
				}
			}
		})
	}
}

func TestProcessBlock(t *testing.T) {
	tests := []struct {
		name           string
		blockNumber    int
		blockNumberStr string
		mockReturn     []interface{}
		expectedError  error
		expectedTxs    map[string][]Transaction
	}{
		{
			name:           "Valid block with transactions",
			blockNumber:    123456,
			blockNumberStr: "0x1e240",
			mockReturn: []interface{}{
				map[string]interface{}{
					"hash":  "0x123",
					"from":  "0xabc",
					"to":    "0xdef",
					"value": "0x100",
				},
			},
			expectedError: nil,
			expectedTxs: map[string][]Transaction{
				"0xabc": {
					{
						Hash:        "0x123",
						From:        "0xabc",
						To:          "0xdef",
						Value:       "0x100",
						BlockNumber: 123456,
					},
				},
				"0xdef": {
					{
						Hash:        "0x123",
						From:        "0xabc",
						To:          "0xdef",
						Value:       "0x100",
						BlockNumber: 123456,
					},
				},
			},
		},
		// Add more test cases as needed
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := new(mocks.API)
			eParser := NewEthereumParser(mockAPI)

			// Mock the GetTransactions method
			mockAPI.On("GetTransactions", tt.blockNumberStr).Return(tt.mockReturn, nil)

			err := eParser.processBlock(tt.blockNumber)
			assert.Equal(t, tt.expectedError, err)

			for addr, txs := range tt.expectedTxs {
				assert.Len(t, eParser.transactions[addr], len(txs))
				for i, tx := range txs {
					assert.Equal(t, tx.Hash, eParser.transactions[addr][i].Hash)
					assert.Equal(t, tx.From, eParser.transactions[addr][i].From)
					assert.Equal(t, tx.To, eParser.transactions[addr][i].To)
					assert.Equal(t, tx.Value, eParser.transactions[addr][i].Value)
					assert.Equal(t, tx.BlockNumber, eParser.transactions[addr][i].BlockNumber)
				}
			}

			// Assert that the expectations were met
			mockAPI.AssertExpectations(t)
		})
	}
}
func TestRetrieveBlockDatas(t *testing.T) {
	tests := []struct {
		name            string
		currentBlock    int
		mockBlockNum    string
		mockBlockNumErr error
		mockTxs         map[string][]Transaction
		mockProcessErr  error
		expectedError   error
		expectedBlock   int
	}{
		{
			name:            "Error getting current block",
			currentBlock:    0,
			mockBlockNum:    "",
			mockBlockNumErr: fmt.Errorf("error"),
			expectedError:   fmt.Errorf("error getting current block"),
			expectedBlock:   0,
		},
		{
			name:          "Block number less than or equal to current block",
			currentBlock:  10,
			mockBlockNum:  "0xa",
			expectedError: nil,
			expectedBlock: 10,
		},
		{
			name:           "Process block error",
			currentBlock:   0,
			mockBlockNum:   "0x2",
			mockProcessErr: fmt.Errorf("process error"),
			expectedError:  fmt.Errorf("error proccing block 1 %w", fmt.Errorf("process error")),
			expectedBlock:  0,
		},
		{
			name:          "Successful block processing",
			currentBlock:  0,
			mockBlockNum:  "0x2",
			expectedError: nil,
			expectedBlock: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAPI := new(mocks.API)
			eParser := NewEthereumParser(mockAPI, WithWaitTime(0))
			eParser.currentBlock = tt.currentBlock

			// Mock the GetCurrentBlock method
			mockAPI.On("GetCurrentBlock").Return(tt.mockBlockNum, tt.mockBlockNumErr)

			// Mock the processBlock method
			if tt.mockBlockNumErr == nil && (tt.currentBlock < tt.expectedBlock || tt.currentBlock == 0) {
				if tt.mockProcessErr != nil {
					mockAPI.On("GetTransactions", mock.Anything).Return(nil, tt.mockProcessErr)
				} else {
					mockAPI.On("GetTransactions", mock.Anything).Return([]interface{}{}, nil)
				}
			}

			err := eParser.retrieveBlockDatas()
			if err != nil {
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			}

			assert.Equal(t, tt.expectedBlock, eParser.GetCurrentBlock())

			// Assert that the expectations were met
			mockAPI.AssertExpectations(t)
		})
	}
}
