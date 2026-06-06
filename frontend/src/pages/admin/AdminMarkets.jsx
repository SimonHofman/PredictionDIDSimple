// 导入 React 状态管理钩子
import { useState } from 'react';
// 导入管理员 API 方法：注册市场和作废市场
import { adminRegisterMarket, adminVoidMarket } from '../../services/api';

/**
 * 管理后台市场配置页组件
 * 提供市场规则登记/更新和市场作废操作
 */
export default function AdminMarkets() {
  // 赛事 ID 输入值
  const [matchId, setMatchId] = useState('1');
  // 是否需要 VC 凭证
  const [requiresVC, setRequiresVC] = useState(true);
  // 要作废的市场 ID 输入值
  const [voidId, setVoidId] = useState('1');
  // 操作结果消息
  const [msg, setMsg] = useState('');

  // 注册/更新市场规则操作
  const onRegister = async () => {
    try {
      // 调用管理员接口注册市场元数据
      await adminRegisterMarket({
        match_id: Number(matchId),       // 关联的赛事 ID
        requires_vc: requiresVC,         // 是否要求 VC 凭证
        restricted_region: 'EU',         // 受限地区（欧盟）
        resolution_rule: 'HOME_WIN',     // 结算规则（主队获胜）
      });
      // 成功提示
      setMsg('市场元数据已更新');
    } catch (e) {
      // 失败显示错误信息
      setMsg(e.message);
    }
  };

  // 作废市场操作
  const onVoid = async () => {
    try {
      // 调用管理员接口作废指定市场
      await adminVoidMarket(Number(voidId));
      // 成功提示
      setMsg('市场已作废');
    } catch (e) {
      // 失败显示错误信息
      setMsg(e.message);
    }
  };

  return (
    <div>
      {/* 页面标题 */}
      <h2>市场配置</h2>
      {/* 登记/更新市场规则表单 */}
      <div className="card">
        <h3>登记 / 更新市场规则</h3>
        {/* 赛事 ID 输入框 */}
        <label>
          Match ID <input value={matchId} onChange={(e) => setMatchId(e.target.value)} />
        </label>
        {/* 是否需要 VC 凭证复选框 */}
        <label>
          <input type="checkbox" checked={requiresVC} onChange={(e) => setRequiresVC(e.target.checked)} />
          需要 VerifiedFan VC
        </label>
        {/* 保存按钮 */}
        <button type="button" onClick={onRegister}>
          保存
        </button>
      </div>
      {/* 作废市场表单 */}
      <div className="card">
        <h3>作废市场</h3>
        {/* 市场 ID 输入框 */}
        <label>
          Market ID <input value={voidId} onChange={(e) => setVoidId(e.target.value)} />
        </label>
        {/* 作废按钮 */}
        <button type="button" onClick={onVoid}>
          VOID
        </button>
      </div>
      {/* 操作结果消息 */}
      {msg && <p>{msg}</p>}
    </div>
  );
}
