import { useEffect, useState } from 'react';
import { useAccount } from 'wagmi';
import { listMarkets, getMarketPool } from '../services/api';
import {
  addLiquidityV3,
  approveUsdc,
  formatUsdc,
  parseUsdc,
} from '../services/contracts';

export default function Liquidity() {
  const { isConnected } = useAccount();
  const [markets, setMarkets] = useState([]);
  const [selected, setSelected] = useState('');
  const [pool, setPool] = useState(null);
  const [amount, setAmount] = useState('100');
  const [tx, setTx] = useState(null);

  useEffect(() => {
    listMarkets({ status: 'OPEN' }).then((res) => {
      const rows = res.items || [];
      setMarkets(rows);
      if (rows[0]) setSelected(String(rows[0].id));
    });
  }, []);

  useEffect(() => {
    if (!selected) return;
    getMarketPool(selected).then(setPool).catch(console.error);
  }, [selected]);

  const onAdd = async () => {
    const m = markets.find((x) => String(x.id) === selected);
    if (!m || !isConnected) return;
    const collateral = import.meta.env.VITE_MOCK_USDC_ADDRESS;
    if (!collateral) return;
    setTx('pending');
    try {
      const amt = parseUsdc(amount);
      await approveUsdc(collateral, m.market_address, amt);
      await addLiquidityV3(m.market_address, amt);
      setTx('ok');
      getMarketPool(selected).then(setPool);
    } catch (e) {
      setTx(e.message);
    }
  };

  return (
    <div>
      <h1>流动性 (CPMM)</h1>
      <p>向 V3 二元市场注入流动性；需链上合约为 PredictionMarketV3。</p>
      <label>
        市场{' '}
        <select value={selected} onChange={(e) => setSelected(e.target.value)}>
          {markets.map((m) => (
            <option key={m.id} value={m.id}>
              #{m.id} {m.question}
            </option>
          ))}
        </select>
      </label>
      {pool && (
        <p>
          类型 {pool.market_type} · Yes 储备 {formatUsdc(pool.reserve_yes)} · No 储备{' '}
          {formatUsdc(pool.reserve_no)} · 价格 Yes {pool.price_yes_bps || '—'} bps
        </p>
      )}
      <label style={{ display: 'block', marginTop: '0.5rem' }}>
        注入金额 (mUSDC){' '}
        <input type="number" value={amount} onChange={(e) => setAmount(e.target.value)} />
      </label>
      <button type="button" style={{ marginTop: '0.75rem' }} onClick={onAdd} disabled={!isConnected}>
        Approve + Add Liquidity
      </button>
      {tx && <p>{tx}</p>}
    </div>
  );
}
