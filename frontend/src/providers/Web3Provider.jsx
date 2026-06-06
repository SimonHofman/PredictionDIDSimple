// 导入 React Query 客户端和提供者组件
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
// 导入 wagmi 提供者组件
import { WagmiProvider } from 'wagmi';
// 导入 wagmi 配置实例
import { wagmiConfig } from '../wagmi';

// 创建 React Query 客户端实例（用于缓存和管理异步数据请求）
const queryClient = new QueryClient();

/**
 * Web3 提供者组件
 * 为整个应用提供 wagmi（区块链交互）和 React Query（数据缓存）能力
 * @param {{ children: React.ReactNode }} props - 子组件
 */
export default function Web3Provider({ children }) {
  return (
    // WagmiProvider：提供区块链钱包连接和合约交互能力
    <WagmiProvider config={wagmiConfig}>
      {/* QueryClientProvider：提供异步数据请求缓存管理 */}
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    </WagmiProvider>
  );
}
