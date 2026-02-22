package blockchain

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// SendTransaction 从与私钥关联的账户发送交易
func SendTransaction(client *ethclient.Client, privateKeyHex string, toAddressHex string, amount float64) {
	// 1. 加载私钥
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("解析私钥失败: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("无法断言类型: 公钥不是 *ecdsa.PublicKey 类型")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 2. 获取 Nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("获取 nonce 失败: %v", err)
	}

	// 3. 设置金额 (单位: Wei)
	value := big.NewInt(0)
	// 1 ETH = 10^18 Wei. 我们使用 big.Float 转换助手或简单的乘法（如果金额是整数）。
	// 为简单起见，我们假设金额单位是 ETH 并转换为 Wei。
	// 1 ETH = 1000000000000000000 Wei
	amountWei := new(big.Float).Mul(big.NewFloat(amount), big.NewFloat(1e18))
	amountWei.Int(value) // 将 big.Float 转换为 big.Int

	// 4. EIP-1559 动态费用
	gasTipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		log.Fatalf("获取 gas tip cap 建议失败: %v", err)
	}

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		log.Fatalf("获取最新区块头失败: %v", err)
	}

	baseFee := header.BaseFee
	// 计算 GasFeeCap: BaseFee * 2 + GasTipCap
	gasFeeCap := new(big.Int).Add(
		new(big.Int).Mul(baseFee, big.NewInt(2)),
		gasTipCap,
	)

	// 5. Gas 限制
	gasLimit := uint64(21000) // ETH 转账的标准限制

	toAddress := common.HexToAddress(toAddressHex)

	// 6. 创建交易 (EIP-1559)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("获取链 ID 失败: %v", err)
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

	// 7. 签名交易
	signer := types.NewLondonSigner(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		log.Fatalf("签名交易失败: %v", err)
	}

	// 8. 发送交易
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatalf("发送交易失败: %v", err)
	}
}
