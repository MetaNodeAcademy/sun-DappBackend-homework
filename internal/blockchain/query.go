package blockchain

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

// QueryBlockInfo 查询并打印区块信息
func QueryBlockInfo(client *ethclient.Client, blockNumber int64) {
	block, err := client.BlockByNumber(context.Background(), big.NewInt(blockNumber))
	if err != nil {
		log.Fatalf("获取区块失败: %v", err)
	}

	fmt.Printf("区块号: %d\n", block.Number().Uint64())
	fmt.Printf("区块哈希: %s\n", block.Hash().Hex())
	fmt.Printf("区块时间戳: %s\n", time.Unix(int64(block.Time()), 0))
	fmt.Printf("交易数量: %d\n", len(block.Transactions()))
	fmt.Println("--------------------------------------------------")
}
