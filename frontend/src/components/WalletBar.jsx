// 导入 wagmi 钱包相关钩子
import { useAccount, useConnect, useDisconnect } from 'wagmi';
// 导入自定义认证钩子
import { useAuth } from '../hooks/useAuth';

/**
 * 钱包连接栏组件
 * 显示钱包连接状态、登录按钮和地址信息
 */
export default function WalletBar() {
  // 获取当前钱包地址和连接状态
  const { address, isConnected } = useAccount();
  // 获取连接方法、可用连接器列表和连接中状态
  const { connect, connectors, isPending } = useConnect();
  // 获取断开连接方法
  const { disconnect } = useDisconnect();
  // 获取认证相关状态和方法
  const { token, login, logout, loading, error, isAuthenticated } = useAuth();

  return (
    <div style={{ marginLeft: 'auto', display: 'flex', gap: '0.75rem', alignItems: 'center', fontSize: '0.875rem' }}>
      {/* 未连接钱包时显示连接按钮 */}
      {!isConnected && (
        <button type="button" onClick={() => connect({ connector: connectors[0] })} disabled={isPending}>
          连接钱包
        </button>
      )}
      {/* 已连接钱包时显示地址和操作按钮 */}
      {isConnected && (
        <>
          {/* 显示钱包地址缩略（前6位...后4位） */}
          <span title={address}>
            {address?.slice(0, 6)}…{address?.slice(-4)}
          </span>
          {/* 未登录时显示 SIWE 登录按钮 */}
          {!token && (
            <button type="button" onClick={login} disabled={loading}>
              {loading ? '登录中…' : 'SIWE 登录'}
            </button>
          )}
          {/* 已认证时显示已登录标记 */}
          {isAuthenticated && <span className="status-ok">已登录</span>}
          {/* 断开连接按钮：同时登出和断开钱包 */}
          <button type="button" onClick={() => { logout(); disconnect(); }}>
            断开
          </button>
        </>
      )}
      {/* 显示错误信息（如有） */}
      {error && <span className="status-error">{error}</span>}
    </div>
  );
}
