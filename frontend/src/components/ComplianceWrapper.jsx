import { useEffect, useState } from 'react';
import { Outlet } from 'react-router-dom';
import { getCompliance } from '../services/api';

const GATE_KEY = 'compliance_accepted_v1';

export default function ComplianceWrapper() {
  const [state, setState] = useState({ loading: true, restricted: false, accepted: false });

  useEffect(() => {
    const accepted = localStorage.getItem(GATE_KEY) === '1';
    getCompliance()
      .then((c) => setState({ loading: false, restricted: c.restricted, accepted }))
      .catch(() => setState({ loading: false, restricted: false, accepted }));
  }, []);

  if (state.loading) return <p>…</p>;
  if (state.restricted) {
    return (
      <div className="card" style={{ maxWidth: 480, margin: '2rem auto' }}>
        <h1>服务不可用</h1>
        <p>您所在地区暂不支持访问预测市场（地理围栏）。</p>
      </div>
    );
  }
  if (!state.accepted) {
    return (
      <div className="card" style={{ maxWidth: 520, margin: '2rem auto' }}>
        <h1>合规确认</h1>
        <p>参与预测存在资金损失风险。请确认您已阅读用户协议与风险披露，且不在受限司法辖区。</p>
        <button
          type="button"
          onClick={() => {
            localStorage.setItem(GATE_KEY, '1');
            setState((s) => ({ ...s, accepted: true }));
          }}
        >
          我已知晓并继续
        </button>
      </div>
    );
  }
  return <Outlet />;
}
