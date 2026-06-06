import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { useAccount } from 'wagmi';
import { getMarket } from '../services/api';
import {
  approveUsdc,
  buyOutcome,
  buyV3,
  buyMulti,
  formatUsdc,
  parseUsdc,
  readMarketStatus,
  readPoolStateV3,
} from '../services/contracts';
import { getMarketPool } from '../services/api';
import TxStatus from '../components/TxStatus';
import MarketStatusBadge from '../components/MarketStatusBadge';

export default function MarketDetail() {
  const { id } = useParams();
  const { isConnected } = useAccount();
  const [data, setData] = useState(null);
  const [amount, setAmount] = useState('10');
  const [outcome, setOutcome] = useState(0);
  const [pool, setPool] = useState(null);
  const [tx, setTx] = useState({ status: null, error: null, hash: null });

  const collateral = data?.collateral_address;
  const market = data?.market;
  const access = data?.access;

  const refresh = async () => {
    const res = await getMarket(id);
    setData(res);
    if (res?.market?.market_address) {
      await readMarketStatus(res.market.market_address);
      if (res.market.market_type === 'BINARY_V3') {
        setPool(await readPoolStateV3(res.market.market_address));
      }
      getMarketPool(id).then(setPool).catch(() => {});
    }
  };

  useEffect(() => {
    refresh().catch(console.error);
  }, [id]);

  const onBuy = async () => {
    if (!isConnected || !market || !collateral) return;
    if (access && access.requires_vc && !access.allowed) return;
    setTx({ status: 'pending', error: null, hash: null });
    try {
      const amt = parseUsdc(amount);
      await approveUsdc(collateral, market.market_address, amt);
      let receipt;
      if (market.market_type === 'BINARY_V3') {
        receipt = await buyV3(market.market_address, outcome, amt);
      } else if (market.market_type === 'MULTI') {
        receipt = await buyMulti(market.market_address, outcome, amt);
      } else {
        receipt = await buyOutcome(market.market_address, outcome, amt);
      }
      setTx({ status: 'success', error: null, hash: receipt.transactionHash });
      await refresh();
    } catch (e) {
      setTx({ status: 'error', error: e.shortMessage || e.message, hash: null });
    }
  };

  if (!data) return <p>加载中…</p>;
  if (!market) return <p>市场不存在</p>;

  const displayStatus = market.match?.status === 'ORACLE_PENDING' ? 'ORACLE_PENDING' : market.status;

  return (
    <div>
      <h1>{market.question}</h1>
      {market.match && (
        <p>
          {market.match.home_team} vs {market.match.away_team}{' '}
          <MarketStatusBadge status={displayStatus} />
        </p>
      )}
      <p>截止：{new Date(market.end_time).toLocaleString()}</p>
      <p>
        Yes 池 {formatUsdc(market.yes_pool)} · No 池 {formatUsdc(market.no_pool)}
        {market.outcome_count > 2 && ` · ${market.outcome_count} outcomes`}
      </p>
      {pool?.priceYesBps != null && (
        <p>CPMM Yes 价格约 {(Number(pool.priceYesBps) / 100).toFixed(2)}%</p>
      )}
      {pool?.price_yes_bps && (
        <p>API 池快照 Yes {(Number(pool.price_yes_bps) / 100).toFixed(2)}%</p>
      )}
      {access?.requires_vc && !access.allowed && (
        <p className="status-error">
          受限市场：需要 {access.credential_type} 凭证，请前往 DID 页登录并联系管理员签发。
        </p>
      )}

      {market.status === 'OPEN' && (!access?.requires_vc || access.allowed) && (
        <div className="card">
          <h2>下注</h2>
          <label>
            结果{' '}
            <select value={outcome} onChange={(e) => setOutcome(Number(e.target.value))}>
              {Array.from({ length: market.outcome_count || 2 }, (_, i) => (
                <option key={i} value={i}>
                  Outcome {i}
                </option>
              ))}
            </select>
          </label>
          <label style={{ display: 'block', marginTop: '0.5rem' }}>
            金额 (mUSDC){' '}
            <input value={amount} onChange={(e) => setAmount(e.target.value)} type="number" min="1" />
          </label>
          <button type="button" style={{ marginTop: '0.75rem' }} onClick={onBuy} disabled={!isConnected || !collateral}>
            Approve + Buy
          </button>
        </div>
      )}

      <TxStatus {...tx} />
    </div>
  );
}
