package blockchain

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	clientInstance *ethclient.Client
	once           sync.Once
)

// GetClient 返回以太坊客户端的单例实例。
// 如果客户端已初始化，则返回现有实例。
// 否则，它将使用提供的 URL 连接到以太坊网络。
func GetClient(url string) *ethclient.Client {
	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		clientInstance, err = ethclient.DialContext(ctx, url)
		if err != nil {
			log.Fatalf("连接以太坊客户端失败: %v", err)
		}
		log.Println("以太坊客户端连接成功 (单例已初始化)")
	})
	return clientInstance
}

// CloseClient 关闭单例客户端连接（如果存在）。
func CloseClient() {
	if clientInstance != nil {
		clientInstance.Close()
		log.Println("以太坊客户端连接已关闭")
	}
}
