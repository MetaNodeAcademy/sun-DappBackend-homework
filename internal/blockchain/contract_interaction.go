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

// DeployContract 将 Counter 合约部署到网络
func DeployContract(client *ethclient.Client, privateKeyHex string) (string, error) {
	auth, err := getTransactOpts(client, privateKeyHex)
	if err != nil {
		return "", err
	}

	address, tx, _, err := contract.DeployContract(auth, client)
	if err != nil {
		return "", fmt.Errorf("部署合约失败: %v", err)
	}

	fmt.Printf("合约部署已启动。交易哈希: %s\n", tx.Hash().Hex())
	fmt.Printf("合约地址: %s\n", address.Hex())
	return address.Hex(), nil
}

// IncrementCounter 调用 Counter 合约的 increment 函数
func IncrementCounter(client *ethclient.Client, privateKeyHex string, contractAddressHex string) error {
	auth, err := getTransactOpts(client, privateKeyHex)
	if err != nil {
		return err
	}

	contractAddress := common.HexToAddress(contractAddressHex)
	counter, err := contract.NewContract(contractAddress, client)
	if err != nil {
		return fmt.Errorf("加载合约失败: %v", err)
	}

	tx, err := counter.Increment(auth)
	if err != nil {
		return fmt.Errorf("增加计数器失败: %v", err)
	}

	fmt.Printf("增加交易已发送。交易哈希: %s\n", tx.Hash().Hex())
	return nil
}

// GetCounterValue 读取 Counter 合约的当前计数值
func GetCounterValue(client *ethclient.Client, contractAddressHex string) (string, error) {
	contractAddress := common.HexToAddress(contractAddressHex)
	counter, err := contract.NewContract(contractAddress, client)
	if err != nil {
		return "", fmt.Errorf("加载合约失败: %v", err)
	}

	count, err := counter.GetCount(nil)
	if err != nil {
		return "", fmt.Errorf("获取计数值失败: %v", err)
	}

	return count.String(), nil
}

// 创建交易选项的辅助函数
func getTransactOpts(client *ethclient.Client, privateKeyHex string) (*bind.TransactOpts, error) {
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("无效的私钥: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("无法将公钥转换为 ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("获取 nonce 失败: %v", err)
	}

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("获取链 ID 失败: %v", err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("创建 transactor 失败: %v", err)
	}

	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // 单位: wei
	auth.GasLimit = uint64(300000) // 单位: units

	// EIP-1559 动态费用
	gasTipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, fmt.Errorf("获取 gas tip cap 失败: %v", err)
	}

	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("获取最新区块头失败: %v", err)
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
