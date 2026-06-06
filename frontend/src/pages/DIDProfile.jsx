// 导入 React 钩子：副作用和状态管理
import { useEffect, useState } from 'react';
// 导入 wagmi 钱包账户钩子
import { useAccount } from 'wagmi';
// 导入获取用户凭证的 API 方法
import { myCredentials } from '../services/api';
// 导入自定义认证钩子
import { useAuth } from '../hooks/useAuth';
// 导入可验证凭证卡片组件
import VCCard from '../components/VCCard';

/**
 * DID 身份档案页组件
 * 展示用户的去中心化身份（DID）和关联的可验证凭证（VC）
 */
export default function DIDProfile() {
  // 获取钱包地址和连接状态
  const { address, isConnected } = useAccount();
  // 获取认证状态和登录方法
  const { isAuthenticated, login } = useAuth();
  // 可验证凭证列表
  const [items, setItems] = useState([]);

  // 认证成功后加载用户凭证列表
  useEffect(() => {
    if (isAuthenticated) {
      // 调用 API 获取凭证，失败时设置为空数组
      myCredentials().then((r) => setItems(r.items || [])).catch(() => setItems([]));
    }
  }, [isAuthenticated]);

  // 根据钱包地址构造 DID 标识符
  const did = address
    ? `did:pkh:eip155:${import.meta.env.VITE_CHAIN_ID || 31337}:${address.toLowerCase()}`
    : '';

  return (
    <div>
      {/* 页面标题 */}
      <h1>DID 身份</h1>
      {/* 未连接钱包时提示 */}
      {!isConnected && <p>请连接钱包</p>}
      {/* 已连接钱包时显示 DID 和凭证 */}
      {isConnected && (
        <>
          {/* 显示用户 DID 标识符 */}
          <p>
            <code>{did}</code>
          </p>
          {/* 未认证时显示登录按钮 */}
          {!isAuthenticated && (
            <button type="button" onClick={login}>
              SIWE 登录以查看凭证
            </button>
          )}
          {/* 可验证凭证列表标题 */}
          <h2>可验证凭证</h2>
          {/* 无凭证时的提示 */}
          {items.length === 0 && <p>暂无 VC（可由管理员签发 VerifiedFan）</p>}
          {/* 渲染每个凭证卡片 */}
          {items.map((c) => (
            <VCCard key={c.id} credential={c} />
          ))}
        </>
      )}
    </div>
  );
}
