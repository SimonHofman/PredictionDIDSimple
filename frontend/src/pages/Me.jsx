import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAccount } from 'wagmi';
import { myPositions } from '../services/api';
import { claimMarket, formatUsdc } from '../services/contracts';
import { useAuth } from '../hooks/useAuth';
import TxStatus from '../components/TxStatus';
import MarketStatusBadge from '../components/MarketStatusBadge';

export default function Me() {
  const { isConnected } = useAccount();
  const { isAuthenticated, login, loading } = useAuth();
  const [items, setItems] = useState([]);
  const [error, setError] = useState(null);
  const [tx, setTx] = useState({ status: null, error: null, hash: null });

  const load = () => {
    myPositions()
      .then((res) => setItems(res.items || []))
      .catch((e) => setError(e.message));
  };

  useEffect(() => {
    if (isAuthenticated) load();
  }, [isAuthenticated]);

  const onClaim = async (marketAddress) => {
    setTx({ status: 'pending', error: null, hash: null });
    try {
      const receipt = await claimMarket(marketAddress);
      setTx({ status: 'success', error: null, hash: receipt.transactionHash });
      load();
    } catch (e) {
      setTx({ status: 'error', error: e.shortMessage || e.message, hash: null });
    }
  };

  return (
    <div>
      <h1>我的持仓</h1>
      {!isConnected && <p>请先连接钱包</p>}
      {isConnected && !isAuthenticated && (
        <button type="button" onClick={login} disabled={loading}>
          SIWE 登录后查看持仓
        </button>
      )}
      {error && <p className="status-error">{error}</p>}
      {isAuthenticated && items.length === 0 && <p>暂无持仓</p>}
      {items.map((p) => (
        <div key={p.id} className="card">
          <Link to={`/markets/${p.market_id}`}>{p.market?.question || `Market #${p.market_id}`}</Link>
          <p>
            Yes {formatUsdc(p.yes_amount)} / No {formatUsdc(p.no_amount)}
          </p>
          <p>市场状态：<MarketStatusBadge status={p.market?.status} /></p>
          {(p.market?.status === 'RESOLVED' || p.market?.status === 'VOID') && !p.claimed && (
            <button type="button" onClick={() => onClaim(p.market.market_address)}>
              {p.market?.status === 'VOID' ? '退款 Claim' : '领取 Claim'}
            </button>
          )}
          {p.claimed && <span className="status-ok"> 已领取</span>}
        </div>
      ))}
      <TxStatus {...tx} />
    </div>
  );
}
