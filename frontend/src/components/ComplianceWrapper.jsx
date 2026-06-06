// 导入 React 钩子：副作用和状态管理
import { useEffect, useState } from 'react';
// 导入路由出口组件（渲染子路由）
import { Outlet } from 'react-router-dom';
// 导入合规检查 API 方法
import { getCompliance } from '../services/api';

// 合规确认标记在 localStorage 中的存储键名
const GATE_KEY = 'compliance_accepted_v1';

/**
 * 合规检查包裹组件
 * 在用户访问应用前进行地理围栏检查和风险确认
 * 如果用户在受限地区则阻止访问，否则要求用户确认风险后才能继续
 */
export default function ComplianceWrapper() {
  // 组件状态：加载中、是否受限地区、是否已接受合规确认
  const [state, setState] = useState({ loading: true, restricted: false, accepted: false });

  // 组件挂载时执行合规检查
  useEffect(() => {
    // 检查本地存储中是否已接受过合规确认
    const accepted = localStorage.getItem(GATE_KEY) === '1';
    // 调用后端合规检查接口
    getCompliance()
      // 成功时更新状态
      .then((c) => setState({ loading: false, restricted: c.restricted, accepted }))
      // 失败时默认为不受限（容错处理）
      .catch(() => setState({ loading: false, restricted: false, accepted }));
  }, []);

  // 加载中显示省略号
  if (state.loading) return <p>…</p>;
  // 如果在受限地区，显示不可用提示
  if (state.restricted) {
    return (
      <div className="card" style={{ maxWidth: 480, margin: '2rem auto' }}>
        {/* 标题：服务不可用 */}
        <h1>服务不可用</h1>
        {/* 提示信息：地理围栏限制 */}
        <p>您所在地区暂不支持访问预测市场（地理围栏）。</p>
      </div>
    );
  }
  // 如果用户尚未接受合规确认，显示确认界面
  if (!state.accepted) {
    return (
      <div className="card" style={{ maxWidth: 520, margin: '2rem auto' }}>
        {/* 标题：合规确认 */}
        <h1>合规确认</h1>
        {/* 风险提示文本 */}
        <p>参与预测存在资金损失风险。请确认您已阅读用户协议与风险披露，且不在受限司法辖区。</p>
        {/* 确认按钮：点击后保存合规确认标记 */}
        <button
          type="button"
          onClick={() => {
            // 将合规确认标记保存到本地存储
            localStorage.setItem(GATE_KEY, '1');
            // 更新状态为已接受
            setState((s) => ({ ...s, accepted: true }));
          }}
        >
          我已知晓并继续
        </button>
      </div>
    );
  }
  // 合规检查通过，渲染子路由内容
  return <Outlet />;
}
