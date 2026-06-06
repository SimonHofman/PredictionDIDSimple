// 从环境变量中获取 API 基础地址
const apiUrl = import.meta.env.VITE_API_URL;
// 从环境变量中获取区块链网络 ID（字符串类型）
const chainIdRaw = import.meta.env.VITE_CHAIN_ID;

// 导出应用全局配置对象
export const config = {
  // API 基础地址，默认为本地 8080 端口
  apiUrl: apiUrl || 'http://localhost:8080',
  // 区块链网络 ID，默认为 Hardhat 本地网络 31337
  chainId: chainIdRaw ? Number(chainIdRaw) : 31337,
  // 标记 API URL 是否已通过环境变量配置
  apiUrlConfigured: Boolean(apiUrl),
  // MockUSDC 代币合约地址
  mockUsdc: import.meta.env.VITE_MOCK_USDC_ADDRESS || '',
  // 市场工厂合约地址
  marketFactory: import.meta.env.VITE_MARKET_FACTORY_ADDRESS || '',
  // SIWE（以太坊签名登录）的域名
  siweDomain: import.meta.env.VITE_SIWE_DOMAIN || 'localhost',
  // SIWE 的 URI 地址
  siweUri: import.meta.env.VITE_SIWE_URI || 'http://localhost:5173',
};
