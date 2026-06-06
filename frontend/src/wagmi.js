// 导入 wagmi 的 HTTP 传输和配置创建工具
import { http, createConfig } from 'wagmi';
// 导入浏览器注入式钱包连接器（如 MetaMask）
import { injected } from 'wagmi/connectors';
// 导入 viem 的自定义链定义工具
import { defineChain } from 'viem';
// 导入应用配置
import { config as appConfig } from './config';

// 定义 Hardhat 本地开发链配置
export const hardhatLocal = defineChain({
  // 链 ID，从应用配置中获取
  id: appConfig.chainId,
  // 链名称
  name: 'Hardhat Local',
  // 原生代币配置（以太坊 ETH，18 位小数）
  nativeCurrency: { name: 'Ether', symbol: 'ETH', decimals: 18 },
  // RPC 节点地址配置
  rpcUrls: {
    // 默认使用本地 Hardhat 节点
    default: { http: ['http://127.0.0.1:8545'] },
  },
});

// 创建 wagmi 全局配置实例
export const wagmiConfig = createConfig({
  // 支持的区块链列表
  chains: [hardhatLocal],
  // 钱包连接器列表（使用浏览器注入的钱包）
  connectors: [injected()],
  // 各链对应的网络传输方式
  transports: {
    // Hardhat 本地链使用 HTTP 传输
    [hardhatLocal.id]: http(),
  },
});
