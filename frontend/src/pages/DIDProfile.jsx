import { useEffect, useState } from 'react';
import { useAccount } from 'wagmi';
import { myCredentials } from '../services/api';
import { useAuth } from '../hooks/useAuth';
import VCCard from '../components/VCCard';

export default function DIDProfile() {
  const { address, isConnected } = useAccount();
  const { isAuthenticated, login } = useAuth();
  const [items, setItems] = useState([]);

  useEffect(() => {
    if (isAuthenticated) {
      myCredentials().then((r) => setItems(r.items || [])).catch(() => setItems([]));
    }
  }, [isAuthenticated]);

  const did = address
    ? `did:pkh:eip155:${import.meta.env.VITE_CHAIN_ID || 31337}:${address.toLowerCase()}`
    : '';

  return (
    <div>
      <h1>DID 身份</h1>
      {!isConnected && <p>请连接钱包</p>}
      {isConnected && (
        <>
          <p>
            <code>{did}</code>
          </p>
          {!isAuthenticated && (
            <button type="button" onClick={login}>
              SIWE 登录以查看凭证
            </button>
          )}
          <h2>可验证凭证</h2>
          {items.length === 0 && <p>暂无 VC（可由管理员签发 VerifiedFan）</p>}
          {items.map((c) => (
            <VCCard key={c.id} credential={c} />
          ))}
        </>
      )}
    </div>
  );
}
