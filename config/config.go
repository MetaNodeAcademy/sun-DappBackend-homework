package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	InfuraURL   string
	InfuraWSURL string
	PrivateKey  string
}

func LoadConfig() *Config {
	// 尝试从当前目录加载 .env
	err := godotenv.Load()
	if err != nil {
		// 如果未找到，尝试从父目录加载 (以防从 cmd/ 运行)
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("警告: 当前或父目录未找到 .env 文件，正在从环境变量读取")
		}
	}

	infuraURL := os.Getenv("INFURA_URL")
	if infuraURL == "" {
		log.Fatal("未设置 INFURA_URL")
	}

	infuraWSURL := os.Getenv("INFURA_WS_URL")
	// 如果缺少 WS URL，则发出警告但不是致命错误，因为某些模式不需要它
	if infuraWSURL == "" {
		log.Println("警告: 未设置 INFURA_WS_URL (订阅模式需要此项)")
	}

	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("未设置 PRIVATE_KEY")
	}

	return &Config{
		InfuraURL:   infuraURL,
		InfuraWSURL: infuraWSURL,
		PrivateKey:  privateKey,
	}
}
