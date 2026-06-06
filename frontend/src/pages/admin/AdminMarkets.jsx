import { useState } from 'react';
import { adminRegisterMarket, adminVoidMarket } from '../../services/api';

export default function AdminMarkets() {
  const [matchId, setMatchId] = useState('1');
  const [requiresVC, setRequiresVC] = useState(true);
  const [voidId, setVoidId] = useState('1');
  const [msg, setMsg] = useState('');

  const onRegister = async () => {
    try {
      await adminRegisterMarket({
        match_id: Number(matchId),
        requires_vc: requiresVC,
        restricted_region: 'EU',
        resolution_rule: 'HOME_WIN',
      });
      setMsg('市场元数据已更新');
    } catch (e) {
      setMsg(e.message);
    }
  };

  const onVoid = async () => {
    try {
      await adminVoidMarket(Number(voidId));
      setMsg('市场已作废');
    } catch (e) {
      setMsg(e.message);
    }
  };

  return (
    <div>
      <h2>市场配置</h2>
      <div className="card">
        <h3>登记 / 更新市场规则</h3>
        <label>
          Match ID <input value={matchId} onChange={(e) => setMatchId(e.target.value)} />
        </label>
        <label>
          <input type="checkbox" checked={requiresVC} onChange={(e) => setRequiresVC(e.target.checked)} />
          需要 VerifiedFan VC
        </label>
        <button type="button" onClick={onRegister}>
          保存
        </button>
      </div>
      <div className="card">
        <h3>作废市场</h3>
        <label>
          Market ID <input value={voidId} onChange={(e) => setVoidId(e.target.value)} />
        </label>
        <button type="button" onClick={onVoid}>
          VOID
        </button>
      </div>
      {msg && <p>{msg}</p>}
    </div>
  );
}
