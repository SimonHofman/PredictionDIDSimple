// 导入 React 钩子：副作用和状态管理
import { useEffect, useState } from 'react';
// 导入获取平台统计数据的 API 方法
import { getPlatformStats } from '../services/api';
// 导入 USDC 金额格式化工具
import { formatUsdc } from '../services/contracts';

/**
 * 平台统计页组件
 * 展示平台整体运营数据（成交量、用户数、TVL 等）
 */
export default function Stats() {
  // 平台统计数据
  const [stats, setStats] = useState(null);

  // 组件挂载时加载统计数据
  useEffect(() => {
    getPlatformStats().then(setStats).catch(console.error);
  }, []);

  // 数据加载中显示提示
  if (!stats) return <p>加载中…</p>;

  return (
    <div>
      {/* 页面标题 */}
      <h1>平台统计</h1>
      {/* 统计数据列表 */}
      <ul>
        {/* 成交笔数 */}
        <li>成交笔数：{stats.trade_count}</li>
        {/* 成交总量（USDC 计价） */}
        <li>成交量：{formatUsdc(stats.trade_volume)} mUSDC</li>
        {/* 平台估算手续费收入 */}
        <li>估算手续费：{formatUsdc(stats.fees_collected)} mUSDC</li>
        {/* 活跃用户数 */}
        <li>活跃用户：{stats.active_users}</li>
        {/* 当前开放的市场数量 */}
        <li>开放市场：{stats.open_markets}</li>
        {/* 总锁定价值近似值 */}
        <li>TVL 近似：{formatUsdc(stats.tvl_approx)} mUSDC</li>
      </ul>
    </div>
  );
}
