package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"sun-DappBackend-homework/config"
	"sun-DappBackend-homework/internal/blockchain"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 解析命令行参数
	mode := flag.String("mode", "", "运行模式: 'query', 'tx', 'deploy', 'increment', 'count', 'subscribe', 'subscribe-logs'")
	blockNum := flag.Int64("block", 0, "要查询的区块号 (默认: 最新区块) 或 订阅模式的起始扫描高度")
	toAddr := flag.String("to", "", "交易接收方地址")
	amount := flag.Float64("amount", 0.0, "发送的 ETH 金额")
	contractAddr := flag.String("contract", "", "交互的合约地址")

	flag.Parse()

	if *mode == "" {
		fmt.Println("请使用 -mode 参数指定运行模式。")
		fmt.Println("可用模式: query, tx, deploy, increment, count, subscribe, subscribe-logs")
		fmt.Println("示例:")
		fmt.Println("  go run cmd/main.go -mode query -block 123456")
		fmt.Println("  go run cmd/main.go -mode tx -to 0xRecipientAddress -amount 0.001")
		fmt.Println("  go run cmd/main.go -mode deploy")
		fmt.Println("  go run cmd/main.go -mode increment -contract 0xContractAddress")
		fmt.Println("  go run cmd/main.go -mode count -contract 0xContractAddress")
		fmt.Println("  go run cmd/main.go -mode subscribe -block 5430000 (可选: 指定起始高度进行追赶)")
		fmt.Println("  go run cmd/main.go -mode subscribe-logs -contract 0xContractAddress")
		os.Exit(1)
	}

	// 对于订阅模式，我们不需要立即初始化标准的 HTTP 客户端，
	// 并且我们需要以不同方式处理信号。
	if *mode == "subscribe" || *mode == "subscribe-logs" {
		if cfg.InfuraWSURL == "" {
			log.Fatal("订阅模式需要 INFURA_WS_URL。请检查您的 .env 文件。")
		}

		// 检查 URL 前缀是否为 ws:// 或 wss://
		// 虽然 ethclient.DialContext 会根据 URL 方案选择传输方式，
		// 但明确的检查可以帮助用户避免使用 HTTP URL 进行订阅。
		if len(cfg.InfuraWSURL) < 5 || (cfg.InfuraWSURL[:5] != "ws://" && cfg.InfuraWSURL[:6] != "wss://") {
			log.Fatal("INFURA_WS_URL 必须以 ws:// 或 wss:// 开头。请检查您的 .env 文件。")
		}

		// 创建一个在接收到中断信号时取消的上下文
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			sig := <-sigCh
			log.Printf("接收到信号: %v。正在关闭...", sig)
			cancel()
		}()

		if *mode == "subscribe" {
			blockchain.SubscribeNewHead(ctx, cfg.InfuraWSURL, *blockNum)
		} else if *mode == "subscribe-logs" {
			blockchain.SubscribeFilterLogs(ctx, cfg.InfuraWSURL, *contractAddr)
		}
		return
	}

	// 为其他模式连接到以太坊客户端 (HTTP)
	client := blockchain.GetClient(cfg.InfuraURL)
	defer blockchain.CloseClient()

	switch *mode {
	case "query":
		if *blockNum == 0 {
			// 如果未提供区块号，则查询最新区块
			header, err := client.HeaderByNumber(nil, nil)
			if err != nil {
				log.Fatalf("获取最新区块号失败: %v", err)
			}
			*blockNum = header.Number.Int64()
			fmt.Printf("正在查询最新区块: %d\n", *blockNum)
		}
		blockchain.QueryBlockInfo(client, *blockNum)

	case "tx":
		if *toAddr == "" || *amount == 0.0 {
			log.Fatal("交易模式请提供 -to 和 -amount 参数")
		}
		// 假设私钥在配置中
		blockchain.SendTransaction(client, cfg.PrivateKey, *toAddr, *amount)

	case "deploy":
		address, err := blockchain.DeployContract(client, cfg.PrivateKey)
		if err != nil {
			log.Fatalf("部署合约失败: %v", err)
		}
		fmt.Printf("合约部署已启动。交易哈希: %s\n", address)

	case "increment":
		if *contractAddr == "" {
			log.Fatal("增加计数模式请提供 -contract 地址参数")
		}
		if err := blockchain.IncrementCounter(client, cfg.PrivateKey, *contractAddr); err != nil {
			log.Fatalf("增加计数器失败: %v", err)
		}
		fmt.Println("计数器增加成功")

	case "count":
		if *contractAddr == "" {
			log.Fatal("查询计数模式请提供 -contract 地址参数")
		}
		count, err := blockchain.GetCounterValue(client, *contractAddr)
		if err != nil {
			log.Fatalf("获取计数器值失败: %v", err)
		}
		fmt.Printf("当前计数器值: %s\n", count)

	default:
		log.Fatalf("未知模式: %s", *mode)
	}
}
