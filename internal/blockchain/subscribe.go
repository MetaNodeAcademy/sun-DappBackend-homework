package blockchain

import (
	"context"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// SubscribeNewHead 订阅新区块头并打印其信息。
// 它处理断线重连、优雅退出以及启动时的回放扫描。
func SubscribeNewHead(ctx context.Context, wsURL string, startBlock int64) {
	var client *ethclient.Client
	var sub interface {
		Err() <-chan error
		Unsubscribe()
	}
	// 通道定义
	var headers chan *types.Header
	var err error

	// 记录上一次处理的区块号，用于断点续传
	// 如果 startBlock > 0，则初始化为 startBlock - 1
	var lastProcessedBlock int64 = -1
	if startBlock > 0 {
		lastProcessedBlock = startBlock - 1
	}

	// 初始化连接
	client, err = ethclient.DialContext(ctx, wsURL)
	if err != nil {
		log.Printf("连接 WebSocket 失败: %v。5秒后重试...", err)
	} else {
		log.Println("已连接到 WebSocket")
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

		// 2. 如果尚未订阅，则进行订阅流程
		if sub == nil {
			// 在订阅前，先检查是否需要补数据 (Scanner)
			// 获取当前网络最新区块
			header, err := client.HeaderByNumber(ctx, nil)
			if err != nil {
				log.Printf("获取最新区块头失败: %v。重置连接...", err)
				client.Close()
				client = nil
				time.Sleep(2 * time.Second)
				continue
			}
			latestBlock := header.Number.Int64()

			// 如果有上一次处理的区块记录，且小于最新区块，则进行补漏扫描
			if lastProcessedBlock >= 0 && lastProcessedBlock < latestBlock {
				log.Printf("检测到区块缺口: 本地(%d) -> 网络(%d)。开始补数据...", lastProcessedBlock, latestBlock)
				err := ScanBlocks(ctx, client, lastProcessedBlock+1, latestBlock, func(h *types.Header) {
					PrintBlockInfo(h)
					lastProcessedBlock = h.Number.Int64()
				})
				if err != nil {
					log.Printf("扫描补数据过程中出错: %v。将尝试继续订阅...", err)
				}
			} else {
				// 如果没有 lastProcessedBlock (首次运行且未指定 startBlock)，则将 latestBlock 视为起始点
				if lastProcessedBlock == -1 {
					lastProcessedBlock = latestBlock
					log.Printf("未指定起始区块，将从最新区块 %d 开始监听", latestBlock)
				}
			}

			// 创建新通道并订阅
			headers = make(chan *types.Header)
			sub, err = client.SubscribeNewHead(ctx, headers)
			if err != nil {
				log.Printf("订阅失败: %v。5秒后重试...", err)
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
				log.Println("已订阅新区块头")
			}
		}

		// 3. 处理事件循环
		select {
		case <-ctx.Done():
			log.Println("上下文已取消，正在取消订阅...")
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
			// 重连前等待，避免紧密循环
			time.Sleep(2 * time.Second)

		case header, ok := <-headers:
			if !ok {
				log.Println("Header 通道已关闭。正在重新连接...")
				sub = nil
				client = nil
				time.Sleep(1 * time.Second)
				continue
			}

			// 简单的去重检查
			currentNum := header.Number.Int64()
			if currentNum <= lastProcessedBlock {
				// 可能是扫描过程中已经处理过了，或者是重复推送
				continue
			}

			// 如果订阅收到的区块号与 lastProcessedBlock 不连续（gap > 1），
			// 理想情况下应该触发补漏。但由于 SubscribeNewHead 通常推送最新的，
			// 如果我们错过了中间的，通常是因为断线（上面会处理）。
			// 如果是网络延迟导致的乱序或跳跃，这里简单记录。
			if currentNum > lastProcessedBlock+1 {
				log.Printf("警告: 收到非连续区块 (上一个: %d, 当前: %d)。可能丢失了部分区块。", lastProcessedBlock, currentNum)
				// 可选：在这里也可以触发一次小范围 ScanBlocks
				// err := ScanBlocks(ctx, client, lastProcessedBlock+1, currentNum-1, func(h *types.Header) {
				// 	PrintBlockInfo(h)
				// 	lastProcessedBlock = h.Number.Int64()
				// })
			}

			PrintBlockInfo(header)
			lastProcessedBlock = currentNum
		}
	}
}

// printBlockInfo 移除旧的私有函数，改用 scanner.go 中的 PrintBlockInfo
