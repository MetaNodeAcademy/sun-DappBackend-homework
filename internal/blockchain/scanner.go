package blockchain

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ScanBlocks 扫描指定范围的区块头并调用回调函数处理
// start: 起始区块号 (包含)
// end: 结束区块号 (包含)
// onBlock: 处理每个区块的回调函数
func ScanBlocks(ctx context.Context, client *ethclient.Client, start int64, end int64, onBlock func(*types.Header)) error {
	log.Printf("开始扫描区块: %d -> %d", start, end)

	for i := start; i <= end; i++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 获取区块头
		// 使用重试机制增强健壮性
		var header *types.Header
		var err error
		maxRetries := 3

		for retry := 0; retry < maxRetries; retry++ {
			header, err = client.HeaderByNumber(ctx, big.NewInt(i))
			if err == nil {
				break
			}
			log.Printf("获取区块 %d 失败 (尝试 %d/%d): %v", i, retry+1, maxRetries, err)
			time.Sleep(1 * time.Second)
		}

		if err != nil {
			log.Printf("跳过区块 %d: 无法获取 (%v)", i, err)
			continue
		}

		// 调用回调处理
		onBlock(header)
	}

	log.Printf("扫描完成: %d -> %d", start, end)
	return nil
}

// PrintBlockInfo 打印区块头的基本信息
// 这是一个公共辅助函数，供 Scanner 和 Subscriber 使用以保持输出格式一致
func PrintBlockInfo(header *types.Header) {
	fmt.Println("------------------------------------------------")
	fmt.Printf("区块高度:   %s\n", header.Number.String())
	fmt.Printf("区块哈希:   %s\n", header.Hash().Hex())
	fmt.Printf("父哈希:     %s\n", header.ParentHash.Hex())
	fmt.Printf("时间戳:     %d\n", header.Time)
	fmt.Printf("Nonce:      %d\n", header.Nonce.Uint64())
	fmt.Println("------------------------------------------------")
}
