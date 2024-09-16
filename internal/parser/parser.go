package parser

import (
	"context"
	"fmt"
	"log"
	"strings"
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
	api          ethereum.API
	currentBlock int
	// The addresses which are being subscribed
	addresses map[string]struct{}
	// The transactions for each address
	transactions map[string][]Transaction
	mutex        sync.Mutex
	// closing channel is for elegent stop the go routine
	stopChannel chan struct{}
	doneChannel chan struct{}
	waitTime    time.Duration
}

type Option func(*EthereumParser)

func WithWaitTime(duration time.Duration) Option {
	return func(p *EthereumParser) {
		p.waitTime = duration
	}
}

func hexToInt(hexStr string) (int, error) {
	var result int //0x11c37937e08000
	_, err := fmt.Sscanf(hexStr, "0x%x", &result)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func NewEthereumParser(api ethereum.API, options ...Option) *EthereumParser {
	p := &EthereumParser{
		api:          api,
		currentBlock: -1,
		addresses:    make(map[string]struct{}),
		transactions: make(map[string][]Transaction),
		stopChannel:  make(chan struct{}),
		doneChannel:  make(chan struct{}),
	}
	for _, option := range options {
		option(p)
	}
	return p
}

// GetCurrentBlock returns the last parsed block
func (p *EthereumParser) GetCurrentBlock() int {
	return p.currentBlock
}

func (p *EthereumParser) Subscribe(address string) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	address = strings.ToLower(address)
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
	address = strings.ToLower(address)
	if _, exists := p.addresses[address]; !exists {
		return []Transaction{}
	}
	return p.transactions[address]
}

func (p *EthereumParser) Start() {
	for {
		select {
		case <-p.stopChannel:
			p.doneChannel <- struct{}{}
			return
		default:
			// Get the current block number
			// To avoid 429 error
			err := p.retrieveBlockDatas()
			if err != nil {
				log.Println("error retrieveBlockDatas %w", err)
				time.Sleep(p.waitTime)
			}
		}
	}
}

func (p *EthereumParser) retrieveBlockDatas() error {
	log.Println("show me all the addresses")
	for addr := range p.transactions {
		log.Println(addr)
	}

	blockNumberStr, err := p.api.GetCurrentBlock()
	if err != nil {
		return fmt.Errorf("error getting current block %w", err)
	}
	blockNumber, err := hexToInt(blockNumberStr)
	if err != nil {
		return fmt.Errorf("error converting block number %w", err)
	}
	if blockNumber <= p.currentBlock {
		log.Printf("blockNumer %d is less or equals then currentBlock%d \n", blockNumber, p.currentBlock)
		return nil
	}
	if p.currentBlock < 0 {
		p.currentBlock = blockNumber - 1
	}
	log.Printf("have %d block to process\n", blockNumber-p.currentBlock)
	for i := p.currentBlock + 1; i <= blockNumber; i++ {
		err := p.processBlock(i)
		if err != nil {
			return fmt.Errorf("error proccing block %d %w", i, err)
		}
		p.currentBlock = i

		time.Sleep(p.waitTime)
	}
	return nil
}

func (p *EthereumParser) processBlock(blockNumber int) error {
	blockNumberStr := fmt.Sprintf("0x%x", blockNumber)
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
		p.mutex.Lock()
		defer p.mutex.Unlock()
		p.transactions[from] = append(p.transactions[from], transaction)
		p.transactions[to] = append(p.transactions[to], transaction)
	}

	return nil
}

func (p *EthereumParser) Stop() {
	log.Println("Parser is closing")
	close(p.stopChannel)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	select {
	case <-ctx.Done():
		log.Println("Stop timeout")
	case <-p.doneChannel:
		log.Println("Parser closed")
	}
}
