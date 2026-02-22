# Sepolia DApp Backend (Go)

本项目是一个功能完备的以太坊 Sepolia 测试网交互后端，使用 Go 语言 (Golang) 开发。它展示了如何与区块链进行交互，包括查询区块、发送 EIP-1559 交易、智能合约部署与调用、以及健壮的 WebSocket 订阅和事件监听。

## 🌟 核心功能

*   **区块查询**: 获取最新区块或指定高度区块的详细信息。
*   **ETH 转账**: 支持发送符合 EIP-1559 标准的交易，自动估算 Gas 费用。
*   **智能合约交互**:
    *   **部署**: 将 Solidity 合约编译并部署到网络。
    *   **调用**: 支持写入 (Write) 和 读取 (View/Call) 合约方法。
    *   **类型安全**: 使用 `abigen` 生成 Go 绑定代码。
*   **实时订阅 (WebSocket)**:
    *   **区块头订阅**: 实时监听新区块生成。
    *   **日志事件订阅**: 实时监听指定合约的 Event Logs。
*   **健壮性设计**:
    *   **断点续传**: 订阅模式下自动检测区块缺口并补齐历史数据。
    *   **自动重连**: 网络断开时自动尝试重新连接 WebSocket。
    *   **优雅退出**: 处理 SIGINT/SIGTERM 信号，确保资源正确释放。

## 📂 目录结构

```
.
├── cmd/
│   └── main.go                 # 主程序入口，处理命令行参数
├── internal/
│   ├── blockchain/             # 区块链核心逻辑
│   │   ├── client.go           # 单例模式客户端连接
│   │   ├── query.go            # 区块查询
│   │   ├── transaction.go      # 交易发送
│   │   ├── contract_interaction.go # 合约部署与交互
│   │   ├── subscribe.go        # 区块头订阅 (含断点续传)
│   │   ├── subscribe_logs.go   # 日志事件订阅
│   │   └── scanner.go          # 区块范围扫描器
│   └── contract/               # 智能合约绑定
│       ├── Counter.sol         # Solidity 源码
│       └── counter.go          # abigen 生成的 Go 代码
├── config/
│   └── config.go               # 环境变量配置加载
├── .env.example                # 环境变量配置模板
├── go.mod                      # Go 依赖管理
└── README.md                   # 说明文档
```

## 🚀 快速开始

### 1. 环境准备

*   **Go**: 1.18+
*   **Infura 账户**: 获取 Sepolia 网络的 HTTP 和 WebSocket Endpoint。
    *   [注册 Infura](https://infura.io/) -> Create New Key -> Network: Sepolia
*   **MetaMask 钱包**: 获取账户私钥和 Sepolia 测试币。
    *   [Sepolia Faucet](https://sepoliafaucet.com/)

### 2. 配置环境变量

复制模板文件并填写配置：

```bash
cp .env.example .env
```

编辑 `.env` 文件：

```env
# Infura HTTP 节点地址
INFURA_URL=https://sepolia.infura.io/v3/YOUR_PROJECT_ID

# Infura WebSocket 节点地址 (必须以 wss:// 开头)
INFURA_WS_URL=wss://sepolia.infura.io/ws/v3/YOUR_PROJECT_ID

# 你的账户私钥 (不带 0x 前缀)
PRIVATE_KEY=YOUR_PRIVATE_KEY_WITHOUT_0x_PREFIX
```

### 3. 运行项目

项目通过 `cmd/main.go` 运行，使用 `-mode` 参数指定功能模式。

#### 🔍 基础查询与交易

*   **查询最新区块**:
    ```bash
    go run cmd/main.go -mode query
    ```
*   **查询指定区块**:
    ```bash
    go run cmd/main.go -mode query -block 5432100
    ```
*   **发送 ETH 交易**:
    ```bash
    go run cmd/main.go -mode tx -to 0xRecipientAddress -amount 0.001
    ```

#### 📜 智能合约交互

*   **部署合约 (Counter)**:
    ```bash
    go run cmd/main.go -mode deploy
    # 输出: 合约部署已启动。合约地址: 0x...
    ```
*   **读取计数 (View)**:
    ```bash
    go run cmd/main.go -mode count -contract 0xDeployedContractAddress
    ```
*   **增加计数 (Transaction)**:
    ```bash
    go run cmd/main.go -mode increment -contract 0xDeployedContractAddress
    ```

#### 📡 实时订阅与监听

*   **订阅新区块 (实时)**:
    ```bash
    go run cmd/main.go -mode subscribe
    ```
*   **订阅新区块 (带追赶/回放)**:
    从指定高度开始扫描，追赶到最新高度后继续实时监听。适合补齐缺失数据。
    ```bash
    go run cmd/main.go -mode subscribe -block 5430000
    ```
*   **订阅合约日志事件**:
    实时监听指定合约的所有事件。
    ```bash
    go run cmd/main.go -mode subscribe-logs -contract 0xDeployedContractAddress
    ```

## 🛠 开发指南

### 添加新合约

1.  将 Solidity 文件放入 `internal/contract/`。
2.  安装 `abigen` 工具。
3.  编译并生成 Go 绑定：
    ```bash
    solc --abi --bin internal/contract/MyContract.sol -o internal/contract/build --overwrite
    abigen --bin=internal/contract/build/MyContract.bin --abi=internal/contract/build/MyContract.abi --pkg=contract --out=internal/contract/my_contract.go
    ```

### 常见问题

*   **`notifications not supported`**: 确保 `.env` 中的 `INFURA_WS_URL` 配置正确，且必须以 `wss://` 开头。
*   **交易一直 Pending**: 检查 Gas 费是否过低，或者 Sepolia 网络是否拥堵。
*   **`insufficient funds`**: 确保账户有足够的 Sepolia ETH。

## ⚠️ 免责声明

本项目仅供学习和测试目的。请勿将包含真实资产的私钥上传到代码仓库或在生产环境中使用不安全的配置。
