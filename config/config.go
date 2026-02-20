package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	InfuraURL  string
	PrivateKey string
}

func LoadConfig() *Config {
	// Try loading from .env in current directory
	err := godotenv.Load()
	if err != nil {
		// If not found, try loading from parent directory (in case running from cmd/)
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("Warning: No .env file found in current or parent directory, reading from environment variables")
		}
	}

	infuraURL := os.Getenv("INFURA_URL")
	if infuraURL == "" {
		log.Fatal("INFURA_URL is not set")
	}

	privateKey := os.Getenv("PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("PRIVATE_KEY is not set")
	}

	return &Config{
		InfuraURL:  infuraURL,
		PrivateKey: privateKey,
	}
}
