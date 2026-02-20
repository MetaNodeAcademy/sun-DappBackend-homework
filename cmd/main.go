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
	mode := flag.String("mode", "", "Mode of operation: 'query', 'tx', 'deploy', 'increment', 'count'")
	blockNum := flag.Int64("block", 0, "Block number to query (default: latest)")
	toAddr := flag.String("to", "", "Recipient address for transaction")
	amount := flag.Float64("amount", 0.0, "Amount of ETH to send")
	contractAddr := flag.String("contract", "", "Contract address for interaction")

	flag.Parse()

	if *mode == "" {
		fmt.Println("Please specify a mode using -mode flag.")
		fmt.Println("Available modes: query, tx, deploy, increment, count")
		fmt.Println("Examples:")
		fmt.Println("  go run cmd/main.go -mode query -block 123456")
		fmt.Println("  go run cmd/main.go -mode tx -to 0xRecipientAddress -amount 0.001")
		fmt.Println("  go run cmd/main.go -mode deploy")
		fmt.Println("  go run cmd/main.go -mode increment -contract 0xContractAddress")
		fmt.Println("  go run cmd/main.go -mode count -contract 0xContractAddress")
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

	case "deploy":
		address, err := blockchain.DeployContract(client, cfg.PrivateKey)
		if err != nil {
			log.Fatalf("Failed to deploy contract: %v", err)
		}
		fmt.Printf("Contract successfully deployed at: %s\n", address)

	case "increment":
		if *contractAddr == "" {
			log.Fatal("Please provide -contract address for increment mode")
		}
		if err := blockchain.IncrementCounter(client, cfg.PrivateKey, *contractAddr); err != nil {
			log.Fatalf("Failed to increment counter: %v", err)
		}
		fmt.Println("Counter incremented successfully")

	case "count":
		if *contractAddr == "" {
			log.Fatal("Please provide -contract address for count mode")
		}
		count, err := blockchain.GetCounterValue(client, *contractAddr)
		if err != nil {
			log.Fatalf("Failed to get counter value: %v", err)
		}
		fmt.Printf("Current Counter Value: %s\n", count)

	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}
}
