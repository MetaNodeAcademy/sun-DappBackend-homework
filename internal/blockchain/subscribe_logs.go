package blockchain

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// SubscribeFilterLogs 订阅合约日志事件并打印
// 支持断线重连和优雅退出
func SubscribeFilterLogs(ctx context.Context, wsURL string, contractAddr string) {
	var client *ethclient.Client
	var sub interface {
		Err() <-chan error
		Unsubscribe()
	}
	var logs chan types.Log
	var err error

	// 解析合约地址
	var addresses []common.Address
	if contractAddr != "" {
		if !common.IsHexAddress(contractAddr) {
			log.Fatalf("无效的合约地址: %s", contractAddr)
		}
		addresses = []common.Address{common.HexToAddress(contractAddr)}
		log.Printf("正在监听合约地址: %s", contractAddr)
	} else {
		log.Println("未指定合约地址，将监听所有事件 (注意：流量可能很大)")
	}

	// 构建过滤条件
	// 注意：SubscribeFilterLogs 通常只使用 Addresses 和 Topics
	// FromBlock 和 ToBlock 在订阅模式下通常被忽略（只监听新事件）
	query := ethereum.FilterQuery{
		Addresses: addresses,
	}

	// 初始化连接
	client, err = ethclient.DialContext(ctx, wsURL)
	if err != nil {
		log.Printf("连接 WebSocket 失败: %v。5秒后重试...", err)
	} else {
		log.Println("已连接到 WebSocket (Log Subscription)")
	}

	for {
		// 1. 确保客户端已连接
		if client == nil {
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				log.Println("正在重新连接 WebSocket...")
				client, err = ethclient.DialContext(ctx, wsURL)
				if err != nil {
					log.Printf("连接失败: %v。正在重试...", err)
					client = nil
					continue
				}
				log.Println("已连接到 WebSocket")
			}
		}

		// 2. 如果尚未订阅，则进行订阅
		if sub == nil {
			logs = make(chan types.Log)
			sub, err = client.SubscribeFilterLogs(ctx, query, logs)
			if err != nil {
				log.Printf("订阅日志失败: %v。5秒后重试...", err)
				client.Close()
				client = nil
				sub = nil
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			} else {
				log.Println("已成功订阅日志事件")
			}
		}

		// 3. 处理事件循环
		select {
		case <-ctx.Done():
			log.Println("上下文已取消，正在取消日志订阅...")
			if sub != nil {
				sub.Unsubscribe()
			}
			if client != nil {
				client.Close()
			}
			return

		case err := <-sub.Err():
			log.Printf("订阅错误: %v。正在重新连接...", err)
			sub.Unsubscribe()
			sub = nil
			client.Close()
			client = nil
			time.Sleep(2 * time.Second)

		case vLog := <-logs:
			printLogInfo(vLog)
		}
	}
}

// printLogInfo 打印日志详细信息
func printLogInfo(vLog types.Log) {
	fmt.Println("================================================")
	fmt.Printf("区块号:     %d\n", vLog.BlockNumber)
	fmt.Printf("交易哈希:   %s\n", vLog.TxHash.Hex())
	fmt.Printf("日志索引:   %d\n", vLog.Index)
	fmt.Printf("合约地址:   %s\n", vLog.Address.Hex())
	
	fmt.Println("Topics:")
	for i, topic := range vLog.Topics {
		fmt.Printf("  [%d] %s\n", i, topic.Hex())
	}

	fmt.Printf("Data (Hex): %x\n", vLog.Data)
	fmt.Println("================================================")
}
