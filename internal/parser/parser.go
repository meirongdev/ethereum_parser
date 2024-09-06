package parser

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/meirongdev/ethereum_parser/internal/ethereum"
)

type Transaction struct {
	Hash        string
	From        string
	To          string
	Value       string
	BlockNumber int
}

type Parser interface {
	// last parsed block
	GetCurrentBlock() int
	// add address to observer
	Subscribe(address string) bool
	// list of inbound or outbound transactions for an address
	GetTransactions(address string) []Transaction
}

type EthereumParser struct {
	api          ethereum.EthereumAPI
	currentBlock int
	// The addresses which are being subscribed
	addresses map[string]struct{}
	// The transactions for each address
	transactions map[string][]Transaction
	mutex        sync.Mutex
}

func hexToInt(hexStr string) (int, error) {
	var result int //0x11c37937e08000
	_, err := fmt.Sscanf(hexStr, "0x%x", &result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func NewEthereumParser(api ethereum.EthereumAPI) *EthereumParser {
	return &EthereumParser{
		api:          api,
		currentBlock: -1,
		addresses:    make(map[string]struct{}),
		transactions: make(map[string][]Transaction),
	}
}

// GetCurrentBlock returns the last parsed block
func (p *EthereumParser) GetCurrentBlock() int {
	return p.currentBlock
}

func (p *EthereumParser) Subscribe(address string) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if _, exists := p.addresses[address]; exists {
		return false
	}
	// Add the address to the list of addresses
	p.addresses[address] = struct{}{}
	return true
}

func (p *EthereumParser) GetTransactions(address string) []Transaction {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if _, exists := p.addresses[address]; !exists {
		return []Transaction{}
	}
	return p.transactions[address]
}

func (p *EthereumParser) Start() {
	for {
		log.Println("show me all the addresses")
		for addr := range p.transactions {
			log.Println(addr)
		}

		// Get the current block number
		blockNumberStr, err := p.api.GetCurrentBlock()
		if err != nil {
			fmt.Println("Error getting current block", err)
			continue
		}
		blockNumber, err := hexToInt(blockNumberStr)
		if err != nil {
			fmt.Println("Error converting block number", err)
			continue
		}
		if blockNumber <= p.currentBlock {
			continue
		}
		if p.currentBlock < 0 {
			p.currentBlock = blockNumber - 1
		}
		for i := p.currentBlock + 1; i <= blockNumber; i++ {
			err := p.processBlock(i)
			if err != nil {
				log.Println("Error processing block", i, err)
				break
			}
			p.currentBlock = i
			// To avoid 429 error
			time.Sleep(time.Second * 30)
		}
	}
}

func (p *EthereumParser) processBlock(blockNumber int) error {
	blockNumberStr := fmt.Sprintf("0x%x", blockNumber)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	log.Printf("Processing block %d", blockNumber)
	txList, err := p.api.GetTransactions(blockNumberStr)
	if err != nil {
		return err
	}
	log.Printf("Found %d transactions in block %d", len(txList), blockNumber)

	// convert the tx to Transaction struct
	for _, tx := range txList {
		txMap, ok := tx.(map[string]interface{})
		if !ok {
			continue
		}

		hash, ok := txMap["hash"].(string)
		if !ok {
			log.Printf("Error getting hash for transaction %v", txMap["hash"])
			continue
		}
		from, ok := txMap["from"].(string)
		if !ok {
			log.Printf("Error getting from address for transaction %s %v", hash, txMap["from"])
			continue
		}
		// some transactions don't have a "to" field
		to, ok := txMap["to"].(string)
		if !ok {
			log.Printf("Error getting to address for transaction %s %v", hash, txMap["to"])
			continue
		}

		value, ok := txMap["value"].(string)
		if !ok {
			log.Printf("Error getting value for transaction %s %v", hash, txMap["value"])
			continue
		}

		// Check if the address is involved in the transaction (either as sender or receiver)
		transaction := Transaction{
			Hash:        hash,
			From:        from,
			To:          to,
			Value:       value,
			BlockNumber: blockNumber,
		}
		p.transactions[from] = append(p.transactions[from], transaction)
		p.transactions[to] = append(p.transactions[to], transaction)
	}

	return nil
}
