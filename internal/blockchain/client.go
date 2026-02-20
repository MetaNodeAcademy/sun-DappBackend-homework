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

// GetClient returns a singleton instance of the Ethereum client.
// If the client is already initialized, it returns the existing instance.
// Otherwise, it connects to the Ethereum network using the provided URL.
func GetClient(url string) *ethclient.Client {
	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error
		clientInstance, err = ethclient.DialContext(ctx, url)
		if err != nil {
			log.Fatalf("Failed to connect to the Ethereum client: %v", err)
		}
		log.Println("Ethereum client connected successfully (Singleton initialized)")
	})
	return clientInstance
}

// CloseClient closes the singleton client connection if it exists.
func CloseClient() {
	if clientInstance != nil {
		clientInstance.Close()
		log.Println("Ethereum client connection closed")
	}
}
