package blockchain

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// QueryBlockInfo queries and prints information about a block
func QueryBlockInfo(client *ethclient.Client, blockNumber int64) {
	block, err := client.BlockByNumber(context.Background(), big.NewInt(blockNumber))
	if err != nil {
		log.Fatalf("Failed to retrieve block: %v", err)
	}

	fmt.Printf("Block Number: %d\n", block.Number().Uint64())
	fmt.Printf("Block Hash: %s\n", block.Hash().Hex())
	fmt.Printf("Block Timestamp: %s\n", time.Unix(int64(block.Time()), 0))
	fmt.Printf("Transaction Count: %d\n", len(block.Transactions()))
	fmt.Println("--------------------------------------------------")
}
