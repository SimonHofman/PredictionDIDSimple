// 导入 React 钩子：副作用和状态管理
import { useEffect, useState } from 'react';
// 导入路由链接组件
import { Link } from 'react-router-dom';
// 导入获取市场列表的 API 方法
import { listMarkets } from '../services/api';
// 导入 USDC 金额格式化工具
import { formatUsdc } from '../services/contracts';

/**
 * 预测市场列表页组件
 * 展示所有已创建的预测市场及其池信息
 */
export default function Markets() {
  // 市场列表数据
  const [items, setItems] = useState([]);
  // 错误信息
  const [error, setError] = useState(null);

  // 组件挂载时加载市场列表（最多 50 条）
  useEffect(() => {
    listMarkets({ limit: 50 })
      .then((res) => setItems(res.items || []))
      .catch((e) => setError(e.message));
  }, []);

  return (
    <div>
      {/* 页面标题 */}
      <h1>预测市场</h1>
      {/* 显示错误信息（如有） */}
      {error && <p className="status-error">{error}</p>}
      {/* 无市场时的提示 */}
      {items.length === 0 && <p>暂无市场。部署合约后运行 seed-markets，并确保 Indexer 运行。</p>}
      {/* 遍历渲染每个市场卡片 */}
      {items.map((mk) => (
        // 市场卡片链接到市场详情页
        <Link key={mk.id} to={`/markets/${mk.id}`} className="card link-card">
          {/* 市场问题描述 */}
          <p>{mk.question}</p>
          {/* 关联赛事信息（如有） */}
          {mk.match && (
            <p>
              {mk.match.home_team} vs {mk.match.away_team}
            </p>
          )}
          {/* 显示 Yes/No 池金额和市场状态 */}
          <p>
            Yes 池 {formatUsdc(mk.yes_pool)} / No 池 {formatUsdc(mk.no_pool)} · {mk.status}
          </p>
        </Link>
      ))}
    </div>
  );
}
