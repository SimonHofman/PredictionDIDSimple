import { useEffect, useState } from 'react';
import { getPlatformStats } from '../services/api';
import { formatUsdc } from '../services/contracts';

export default function Stats() {
  const [stats, setStats] = useState(null);

  useEffect(() => {
    getPlatformStats().then(setStats).catch(console.error);
  }, []);

  if (!stats) return <p>加载中…</p>;

  return (
    <div>
      <h1>平台统计</h1>
      <ul>
        <li>成交笔数：{stats.trade_count}</li>
        <li>成交量：{formatUsdc(stats.trade_volume)} mUSDC</li>
        <li>估算手续费：{formatUsdc(stats.fees_collected)} mUSDC</li>
        <li>活跃用户：{stats.active_users}</li>
        <li>开放市场：{stats.open_markets}</li>
        <li>TVL 近似：{formatUsdc(stats.tvl_approx)} mUSDC</li>
      </ul>
    </div>
  );
}
