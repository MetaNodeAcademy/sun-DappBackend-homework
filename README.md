# Sepolia 区块链交互项目

本项目演示了如何使用 Go 语言与 Sepolia 测试网络进行交互，包括查询区块信息和发送以太币交易。

## 目录结构

```
.
├── cmd/
│   └── main.go           # 主程序入口
├── internal/
│   └── blockchain/       # 区块链相关逻辑
│       ├── client.go     # 连接客户端
│       ├── query.go      # 查询区块
│       └── transaction.go # 发送交易
├── config/
│   └── config.go         # 配置加载
├── .env.example          # 环境变量示例
├── go.mod                # Go 模块文件
└── README.md             # 说明文档
```

## 环境准备

1.  **安装 Go 语言环境**：确保已安装 Go 1.18+。
2.  **获取 Infura API Key**：
    *   注册 [Infura](https://infura.io/) 账户。
    *   创建一个新项目 (Web3 API)。
    *   获取 Sepolia 网络的 HTTPS Endpoint (例如: `https://sepolia.infura.io/v3/YOUR_PROJECT_ID`)。
3.  **准备以太坊账户**：
    *   安装 MetaMask 钱包。
    *   切换到 Sepolia 测试网络。
    *   获取账户私钥 (注意保密，不要上传到 GitHub)。
    *   从 [Sepolia Faucet](https://sepoliafaucet.com/) 获取测试币。

## 配置

1.  复制 `.env.example` 为 `.env`：
    ```bash
    cp .env .env
    ```
2.  编辑 `.env` 文件，填入你的 Infura URL 和 私钥：
    ```env
    INFURA_URL=https://sepolia.infura.io/v3/YOUR_INFURA_PROJECT_ID
    PRIVATE_KEY=YOUR_PRIVATE_KEY_WITHOUT_0x_PREFIX
    ```
    *注意：私钥不需要带 `0x` 前缀。*

## 运行

### 1. 查询区块信息

查询最新区块：
```bash
go run cmd/main.go -mode query
```

查询指定区块号 (例如 123456)：
```bash
go run cmd/main.go -mode query -block 123456
```

### 2. 发送交易

发送 ETH 给指定地址：
```bash
go run cmd/main.go -mode tx -to 0xRecipientAddress -amount 0.001
```
*   `-to`: 接收方地址 (以 `0x` 开头)。
*   `-amount`: 发送金额 (单位 ETH)。

## 注意事项

*   请确保账户中有足够的 Sepolia ETH 用于支付 Gas 费。
*   私钥非常敏感，请勿泄露。
*   本项目仅供学习使用，请勿在主网使用真实资产测试。
