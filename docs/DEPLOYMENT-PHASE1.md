# Phase 1 本地部署与演示

## 合约地址

本地部署后地址写在 `contracts/deployments.local.json`：

```json
{
  "mockUSDC": "0x...",
  "marketFactory": "0x...",
  "oracle": "0x..."
}
```

将 `marketFactory`、`mockUSDC` 填入根目录 `.env` 与 `frontend/.env`。

## 完整演示流程

### 1. 基础设施

```powershell
docker compose up -d
$env:DATABASE_URL="postgres://prediction:prediction@localhost:5432/prediction?sslmode=disable"
cd backend
go run ./cmd/migrate
go run ./cmd/seed
```

### 2. 本地区块链

```powershell
cd contracts
npx hardhat node
```

新终端：

```powershell
cd contracts
npx hardhat run scripts/deploy.js --network localhost
npx hardhat run scripts/seed-markets.js --network localhost
.\scripts\sync-abi.ps1
```

记录地址到 `.env` / `frontend/.env`。

### 3. 后端 + 前端

```powershell
# 根目录 .env 设置 MARKET_FACTORY_ADDRESS, MOCK_USDC_ADDRESS, JWT_SECRET 等
cd backend
go run ./cmd/api
```

```powershell
cd frontend
npm run dev
```

### 4. 用户操作

1. 浏览器 http://localhost:5173
2. MetaMask 连接 Hardhat 本地链 (31337)，导入 Hardhat 默认账户
3. 在合约脚本中已为 deployer mint mUSDC；可用 `MockUSDC.mint` 给其他账户
4. SIWE 登录 → 市场页 → 选择 Yes/No → Approve + Buy
5. 运维结算：

```powershell
$env:MARKET_ADDRESS="0x..."
$env:WINNING_OUTCOME="0"
npx hardhat run scripts/resolve.js --network localhost
```

6. 我的 → Claim

## 测试网

将 `hardhat.config.js` 增加 sepolia 网络，配置 `DEPLOYER_PRIVATE_KEY` 与 `ETH_RPC_URL`，执行：

```powershell
npx hardhat run scripts/deploy.js --network sepolia
```

地址写入 `docs/deployments/sepolia.json`（自行创建）。
