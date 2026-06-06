// 导入 React 钩子：副作用和状态管理
import { useEffect, useState } from 'react';
// 导入路由链接和参数钩子
import { Link, useParams } from 'react-router-dom';
// 导入获取赛事详情的 API 方法
import { getMatch } from '../services/api';

/**
 * 赛事详情页组件
 * 展示单场赛事的详细信息和关联的预测市场列表
 */
export default function MatchDetail() {
  // 从 URL 参数获取赛事 ID
  const { id } = useParams();
  // 赛事数据
  const [data, setData] = useState(null);
  // 错误信息
  const [error, setError] = useState(null);

  // 组件挂载或 ID 变化时加载赛事数据
  useEffect(() => {
    getMatch(id)
      .then(setData)
      .catch((e) => setError(e.message));
  }, [id]);

  // 加载出错时显示错误信息
  if (error) return <p className="status-error">{error}</p>;
  // 数据加载中显示提示
  if (!data) return <p>加载中…</p>;

  // 提取赛事对象
  const m = data.match;
  return (
    <div>
      {/* 赛事标题：主队 vs 客队 */}
      <h1>
        {m.home_team} vs {m.away_team}
      </h1>
      {/* 开球时间 */}
      <p>开球：{new Date(m.kickoff_at).toLocaleString()}</p>
      {/* 赛事状态 */}
      <p>状态：{m.status}</p>
      {/* 如果有比分则显示 */}
      {m.home_score != null && (
        <p>
          比分 {m.home_score} - {m.away_score}
        </p>
      )}

      {/* 关联的预测市场列表 */}
      <h2>关联市场</h2>
      {/* 无关联市场时的提示 */}
      {(data.markets || []).length === 0 && <p>暂无市场</p>}
      {/* 遍历渲染每个关联市场卡片 */}
      {(data.markets || []).map((mk) => (
        <Link key={mk.id} to={`/markets/${mk.id}`} className="card link-card">
          {mk.question} · {mk.status}
        </Link>
      ))}
    </div>
  );
}
