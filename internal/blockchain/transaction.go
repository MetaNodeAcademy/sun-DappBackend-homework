package blockchain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// SendTransaction sends a transaction from the account associated with the private key
func SendTransaction(client *ethclient.Client, privateKeyHex string, toAddressHex string, amount float64) {
	// 1. Load Private Key
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 2. Get Nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}

	// 3. Set Amount (in Wei)
	value := big.NewInt(0)
	// 1 ETH = 10^18 Wei. We use a float to big int conversion helper or simple multiplication if amount is integer.
	// For simplicity, let's assume amount is in ETH and convert to Wei.
	// 1 ETH = 1000000000000000000 Wei
	amountWei := new(big.Float).Mul(big.NewFloat(amount), big.NewFloat(1e18))
	amountWei.Int(value) // Convert big.Float to big.Int

	// 4. EIP-1559 Dynamic Fees
	gasTipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		log.Fatalf("Failed to suggest gas tip cap: %v", err)
	}

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to get latest block header: %v", err)
	}

	baseFee := header.BaseFee
	// Calculate GasFeeCap: BaseFee * 2 + GasTipCap
	gasFeeCap := new(big.Int).Add(
		new(big.Int).Mul(baseFee, big.NewInt(2)),
		gasTipCap,
	)

	// 5. Gas Limit
	gasLimit := uint64(21000) // Standard limit for ETH transfer

	toAddress := common.HexToAddress(toAddressHex)

	// 6. Create Transaction (EIP-1559)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get chain ID: %v", err)
	}

	txData := &types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &toAddress,
		Value:     value,
		Data:      nil,
	}
	tx := types.NewTx(txData)

	// 7. Sign Transaction
	signer := types.NewLondonSigner(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		log.Fatalf("Failed to sign transaction: %v", err)
	}

	// 8. Send Transaction
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("Failed to send transaction: %v", err)
	}

	fmt.Printf("Transaction sent! Hash: %s\n", signedTx.Hash().Hex())
}
