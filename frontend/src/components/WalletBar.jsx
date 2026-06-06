import { useAccount, useConnect, useDisconnect } from 'wagmi';
import { useAuth } from '../hooks/useAuth';

export default function WalletBar() {
  const { address, isConnected } = useAccount();
  const { connect, connectors, isPending } = useConnect();
  const { disconnect } = useDisconnect();
  const { token, login, logout, loading, error, isAuthenticated } = useAuth();

  return (
    <div style={{ marginLeft: 'auto', display: 'flex', gap: '0.75rem', alignItems: 'center', fontSize: '0.875rem' }}>
      {!isConnected && (
        <button type="button" onClick={() => connect({ connector: connectors[0] })} disabled={isPending}>
          连接钱包
        </button>
      )}
      {isConnected && (
        <>
          <span title={address}>
            {address?.slice(0, 6)}…{address?.slice(-4)}
          </span>
          {!token && (
            <button type="button" onClick={login} disabled={loading}>
              {loading ? '登录中…' : 'SIWE 登录'}
            </button>
          )}
          {isAuthenticated && <span className="status-ok">已登录</span>}
          <button type="button" onClick={() => { logout(); disconnect(); }}>
            断开
          </button>
        </>
      )}
      {error && <span className="status-error">{error}</span>}
    </div>
  );
}
