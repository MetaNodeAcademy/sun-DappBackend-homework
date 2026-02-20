package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"sun-DappBackend-homework/config"
	"sun-DappBackend-homework/internal/blockchain"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to Ethereum client
	client := blockchain.GetClient(cfg.InfuraURL)
	defer blockchain.CloseClient()

	// Parse command line flags
	mode := flag.String("mode", "", "Mode of operation: 'query' or 'tx'")
	blockNum := flag.Int64("block", 0, "Block number to query (default: latest)")
	toAddr := flag.String("to", "", "Recipient address for transaction")
	amount := flag.Float64("amount", 0.0, "Amount of ETH to send")

	flag.Parse()

	if *mode == "" {
		fmt.Println("Please specify a mode using -mode flag.")
		fmt.Println("Available modes: query, tx")
		fmt.Println("Examples:")
		fmt.Println("  go run cmd/main.go -mode query -block 123456")
		fmt.Println("  go run cmd/main.go -mode tx -to 0xRecipientAddress -amount 0.001")
		os.Exit(1)
	}

	switch *mode {
	case "query":
		if *blockNum == 0 {
			// If no block number provided, query the latest block
			header, err := client.HeaderByNumber(nil, nil)
			if err != nil {
				log.Fatalf("Failed to get latest block number: %v", err)
			}
			*blockNum = header.Number.Int64()
			fmt.Printf("Querying latest block: %d\n", *blockNum)
		}
		blockchain.QueryBlockInfo(client, *blockNum)

	case "tx":
		if *toAddr == "" || *amount == 0.0 {
			log.Fatal("Please provide -to and -amount for transaction mode")
		}
		// Assuming private key is in config
		blockchain.SendTransaction(client, cfg.PrivateKey, *toAddr, *amount)

	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}
