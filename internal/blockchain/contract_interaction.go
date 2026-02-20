package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"sun-DappBackend-homework/internal/contract"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// DeployContract deploys the Counter contract to the network
func DeployContract(client *ethclient.Client, privateKeyHex string) (string, error) {
	auth, err := getTransactOpts(client, privateKeyHex)
	if err != nil {
		return "", err
	}

	address, tx, _, err := contract.DeployContract(auth, client)
	if err != nil {
		return "", fmt.Errorf("failed to deploy contract: %v", err)
	}

	fmt.Printf("Contract deployment initiated. Tx Hash: %s\n", tx.Hash().Hex())
	fmt.Printf("Contract Address: %s\n", address.Hex())
	return address.Hex(), nil
}

// IncrementCounter calls the increment function of the Counter contract
func IncrementCounter(client *ethclient.Client, privateKeyHex string, contractAddressHex string) error {
	auth, err := getTransactOpts(client, privateKeyHex)
	if err != nil {
		return err
	}

	contractAddress := common.HexToAddress(contractAddressHex)
	counter, err := contract.NewContract(contractAddress, client)
	if err != nil {
		return fmt.Errorf("failed to load contract: %v", err)
	}

	tx, err := counter.Increment(auth)
	if err != nil {
		return fmt.Errorf("failed to increment counter: %v", err)
	}

	fmt.Printf("Increment transaction sent. Tx Hash: %s\n", tx.Hash().Hex())
	return nil
}

// GetCounterValue reads the current count from the Counter contract
func GetCounterValue(client *ethclient.Client, contractAddressHex string) (string, error) {
	contractAddress := common.HexToAddress(contractAddressHex)
	counter, err := contract.NewContract(contractAddress, client)
	if err != nil {
		return "", fmt.Errorf("failed to load contract: %v", err)
	}

	count, err := counter.GetCount(nil)
	if err != nil {
		return "", fmt.Errorf("failed to get count: %v", err)
	}

	return count.String(), nil
}

// Helper to create transaction options
func getTransactOpts(client *ethclient.Client, privateKeyHex string) (*bind.TransactOpts, error) {
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %v", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain id: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %v", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units

	// EIP-1559 Dynamic Fees
	gasTipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get gas tip cap: %v", err)
	}

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get header: %v", err)
	}

	baseFee := header.BaseFee
	gasFeeCap := new(big.Int).Add(
		new(big.Int).Mul(baseFee, big.NewInt(2)),
		gasTipCap,
	)

	auth.GasFeeCap = gasFeeCap
	auth.GasTipCap = gasTipCap

	return auth, nil
}
