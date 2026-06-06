// 导入 React 钩子：副作用和状态管理
import { useEffect, useState } from 'react';
// 导入管理员 API 方法：获取预言机任务列表和重试任务
import { adminListOracleJobs, adminRetryOracleJob } from '../../services/api';

/**
 * 预言机任务队列页组件
 * 展示所有预言机结算任务的状态，并支持手动重试
 */
export default function OracleJobs() {
  // 任务列表数据
  const [items, setItems] = useState([]);
  // 错误信息
  const [error, setError] = useState(null);

  // 加载任务列表的方法
  const load = () => {
    adminListOracleJobs()
      .then((r) => setItems(r.items || []))
      .catch((e) => setError(e.message));
  };

  // 组件挂载时加载数据，并设置每 10 秒自动刷新
  useEffect(() => {
    // 初次加载
    load();
    // 设置定时器：每 10 秒自动刷新任务列表
    const id = setInterval(load, 10000);
    // 组件卸载时清除定时器
    return () => clearInterval(id);
  }, []);

  return (
    <div>
      {/* 页面标题 */}
      <h2>Oracle 任务队列</h2>
      {/* 显示错误信息（如有） */}
      {error && <p className="status-error">{error}</p>}
      {/* 遍历渲染每个任务卡片 */}
      {items.map((j) => (
        <div key={j.id} className="card">
          {/* 任务 ID、状态和相关问题 */}
          <p>
            #{j.id} · {j.status} · {j.question}
          </p>
          {/* 市场合约地址 */}
          <p style={{ fontSize: '0.85rem' }}>{j.market_address}</p>
          {/* 如果有数据源结果，显示主源和备源比分 */}
          {j.primary_home != null && (
            <p>
              主源 {j.primary_home}-{j.primary_away} / 备源 {j.secondary_home}-{j.secondary_away}
            </p>
          )}
          {/* 显示错误信息（如果任务执行出错） */}
          {j.error_message && <p className="status-error">{j.error_message}</p>}
          {/* 需要人工审核的任务显示重试按钮 */}
          {j.status === 'manual_review' && (
            <button type="button" onClick={() => adminRetryOracleJob(j.id).then(load)}>
              重试
            </button>
          )}
        </div>
      ))}
    </div>
  );
}
