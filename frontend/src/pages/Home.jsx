// 导入 React 钩子：副作用和状态管理
import { useEffect, useState } from 'react';
// 导入路由链接组件
import { Link } from 'react-router-dom';
// 导入 API 服务：健康检查和赛事列表
import { getHealth, listMatches } from '../services/api';

/**
 * 首页组件
 * 显示 API 连接状态和即将开始/进行中的赛事列表
 */
export default function Home() {
  // API 连接状态（checking/online/offline）
  const [apiStatus, setApiStatus] = useState('checking');
  // 赛事列表数据
  const [matches, setMatches] = useState([]);
  // 错误信息
  const [error, setError] = useState(null);

  // 组件挂载时加载数据
  useEffect(() => {
    async function load() {
      try {
        // 先检查后端健康状态
        await getHealth();
        // 健康检查通过，标记为在线
        setApiStatus('online');
        // 获取最近 10 场赛事
        const res = await listMatches({ limit: 10 });
        // 设置赛事列表数据
        setMatches(res.items || []);
      } catch (e) {
        // 请求失败，标记为离线
        setApiStatus('offline');
        // 保存错误信息
        setError(e.message);
      }
    }
    load();
  }, []);

  // 过滤出状态为"已排期"或"进行中"的赛事
  const upcoming = matches.filter((m) => ['SCHEDULED', 'LIVE'].includes(m.status));

  return (
    <div>
      {/* 页面标题 */}
      <h1>世界杯预测市场</h1>
      {/* API 连接状态指示器 */}
      <p className={apiStatus === 'online' ? 'status-ok' : 'status-error'}>
        API {apiStatus === 'online' ? '在线' : `离线${error ? `: ${error}` : ''}`}
      </p>

      {/* 即将开始和进行中的赛事标题 */}
      <h2>即将开始 / 进行中</h2>
      {/* 无赛事时的提示 */}
      {upcoming.length === 0 && <p>暂无赛事，请先运行后端 seed/sync。</p>}
      {/* 赛事卡片网格布局 */}
      <div className="grid">
        {upcoming.map((m) => (
          // 每张赛事卡片链接到赛事详情页
          <Link key={m.id} to={`/matches/${m.id}`} className="card link-card">
            {/* 主队 vs 客队 */}
            <strong>
              {m.home_team} vs {m.away_team}
            </strong>
            {/* 开球时间 */}
            <p>{new Date(m.kickoff_at).toLocaleString()}</p>
            {/* 赛事状态徽章 */}
            <span className="badge">{m.status}</span>
            {/* 如果有比分则显示 */}
            {m.home_score != null && (
              <p>
                比分 {m.home_score} - {m.away_score}
              </p>
            )}
          </Link>
        ))}
      </div>
    </div>
  );
}
