// 导入 React 钩子：副作用和状态管理
import { useEffect, useState } from 'react';
// 导入路由链接组件
import { Link } from 'react-router-dom';
// 导入 wagmi 钱包账户钩子
import { useAccount } from 'wagmi';
// 导入获取用户持仓的 API 方法
import { myPositions } from '../services/api';
// 导入合约交互方法：领取奖金和金额格式化
import { claimMarket, formatUsdc } from '../services/contracts';
// 导入自定义认证钩子
import { useAuth } from '../hooks/useAuth';
// 导入交易状态组件
import TxStatus from '../components/TxStatus';
// 导入市场状态徽章组件
import MarketStatusBadge from '../components/MarketStatusBadge';

/**
 * 个人中心页组件
 * 展示用户的持仓列表和领取（Claim）操作
 */
export default function Me() {
  // 获取钱包连接状态
  const { isConnected } = useAccount();
  // 获取认证状态、登录方法和加载状态
  const { isAuthenticated, login, loading } = useAuth();
  // 持仓列表数据
  const [items, setItems] = useState([]);
  // 错误信息
  const [error, setError] = useState(null);
  // 交易状态
  const [tx, setTx] = useState({ status: null, error: null, hash: null });

  // 加载持仓数据的方法
  const load = () => {
    myPositions()
      .then((res) => setItems(res.items || []))
      .catch((e) => setError(e.message));
  };

  // 认证成功后自动加载持仓数据
  useEffect(() => {
    if (isAuthenticated) load();
  }, [isAuthenticated]);

  // 领取奖金操作处理方法
  const onClaim = async (marketAddress) => {
    // 设置交易状态为等待中
    setTx({ status: 'pending', error: null, hash: null });
    try {
      // 调用合约的 claim 方法领取奖金
      const receipt = await claimMarket(marketAddress);
      // 交易成功，更新状态
      setTx({ status: 'success', error: null, hash: receipt.transactionHash });
      // 刷新持仓数据
      load();
    } catch (e) {
      // 交易失败，显示错误信息
      setTx({ status: 'error', error: e.shortMessage || e.message, hash: null });
    }
  };

  return (
    <div>
      {/* 页面标题 */}
      <h1>我的持仓</h1>
      {/* 未连接钱包时提示 */}
      {!isConnected && <p>请先连接钱包</p>}
      {/* 已连接但未认证时显示登录按钮 */}
      {isConnected && !isAuthenticated && (
        <button type="button" onClick={login} disabled={loading}>
          SIWE 登录后查看持仓
        </button>
      )}
      {/* 显示错误信息（如有） */}
      {error && <p className="status-error">{error}</p>}
      {/* 已认证但无持仓时的提示 */}
      {isAuthenticated && items.length === 0 && <p>暂无持仓</p>}
      {/* 遍历渲染每个持仓卡片 */}
      {items.map((p) => (
        <div key={p.id} className="card">
          {/* 市场链接和问题描述 */}
          <Link to={`/markets/${p.market_id}`}>{p.market?.question || `Market #${p.market_id}`}</Link>
          {/* 显示 Yes/No 持仓金额 */}
          <p>
            Yes {formatUsdc(p.yes_amount)} / No {formatUsdc(p.no_amount)}
          </p>
          {/* 显示关联市场的当前状态 */}
          <p>市场状态：<MarketStatusBadge status={p.market?.status} /></p>
          {/* 市场已结算或已作废且未领取时显示 Claim 按钮 */}
          {(p.market?.status === 'RESOLVED' || p.market?.status === 'VOID') && !p.claimed && (
            <button type="button" onClick={() => onClaim(p.market.market_address)}>
              {p.market?.status === 'VOID' ? '退款 Claim' : '领取 Claim'}
            </button>
          )}
          {/* 已领取标记 */}
          {p.claimed && <span className="status-ok"> 已领取</span>}
        </div>
      ))}
      {/* 交易状态显示区域 */}
      <TxStatus {...tx} />
    </div>
  );
}
