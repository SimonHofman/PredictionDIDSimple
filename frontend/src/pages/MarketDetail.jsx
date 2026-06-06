// 导入 React 钩子：副作用和状态管理
import { useEffect, useState } from 'react';
// 导入路由参数钩子
import { useParams } from 'react-router-dom';
// 导入 wagmi 钱包账户钩子
import { useAccount } from 'wagmi';
// 导入获取市场详情的 API 方法
import { getMarket } from '../services/api';
// 导入合约交互方法
import {
  approveUsdc,      // USDC 授权
  buyOutcome,       // V1 市场下注
  buyV3,            // V3 CPMM 市场下注
  buyMulti,         // 多结果市场下注
  formatUsdc,       // USDC 金额格式化
  parseUsdc,        // USDC 金额解析
  readMarketStatus, // 读取市场链上状态
  readPoolStateV3,  // 读取 V3 池状态
} from '../services/contracts';
// 导入获取市场池数据的 API 方法
import { getMarketPool } from '../services/api';
// 导入交易状态组件
import TxStatus from '../components/TxStatus';
// 导入市场状态徽章组件
import MarketStatusBadge from '../components/MarketStatusBadge';

/**
 * 市场详情页组件
 * 展示单个预测市场的详细信息，包括下注操作界面
 */
export default function MarketDetail() {
  // 从 URL 参数获取市场 ID
  const { id } = useParams();
  // 获取钱包连接状态
  const { isConnected } = useAccount();
  // 市场数据
  const [data, setData] = useState(null);
  // 下注金额输入值
  const [amount, setAmount] = useState('10');
  // 选择的结果编号
  const [outcome, setOutcome] = useState(0);
  // 池状态数据
  const [pool, setPool] = useState(null);
  // 交易状态（status、error、hash）
  const [tx, setTx] = useState({ status: null, error: null, hash: null });

  // 从市场数据中提取抵押品地址
  const collateral = data?.collateral_address;
  // 提取市场信息
  const market = data?.market;
  // 提取访问控制信息（是否需要 VC）
  const access = data?.access;

  // 刷新市场数据的方法
  const refresh = async () => {
    // 从 API 获取市场详情
    const res = await getMarket(id);
    setData(res);
    // 如果市场有链上地址，读取链上状态
    if (res?.market?.market_address) {
      // 读取链上市场状态
      await readMarketStatus(res.market.market_address);
      // 如果是 V3 类型市场，读取 CPMM 池状态
      if (res.market.market_type === 'BINARY_V3') {
        setPool(await readPoolStateV3(res.market.market_address));
      }
      // 同时从 API 获取池快照数据
      getMarketPool(id).then(setPool).catch(() => {});
    }
  };

  // 组件挂载或 ID 变化时加载数据
  useEffect(() => {
    refresh().catch(console.error);
  }, [id]);

  // 下注操作处理方法
  const onBuy = async () => {
    // 检查前提条件：已连接钱包、市场存在、有抵押品地址
    if (!isConnected || !market || !collateral) return;
    // 如果市场需要 VC 且用户未获授权，阻止操作
    if (access && access.requires_vc && !access.allowed) return;
    // 设置交易状态为等待中
    setTx({ status: 'pending', error: null, hash: null });
    try {
      // 将输入金额解析为链上整数
      const amt = parseUsdc(amount);
      // 先授权 USDC 代币给市场合约
      await approveUsdc(collateral, market.market_address, amt);
      let receipt;
      // 根据市场类型调用不同的购买方法
      if (market.market_type === 'BINARY_V3') {
        // V3 CPMM 二元市场
        receipt = await buyV3(market.market_address, outcome, amt);
      } else if (market.market_type === 'MULTI') {
        // 多结果市场
        receipt = await buyMulti(market.market_address, outcome, amt);
      } else {
        // V1 二元市场
        receipt = await buyOutcome(market.market_address, outcome, amt);
      }
      // 交易成功，更新状态
      setTx({ status: 'success', error: null, hash: receipt.transactionHash });
      // 刷新市场数据
      await refresh();
    } catch (e) {
      // 交易失败，显示错误信息
      setTx({ status: 'error', error: e.shortMessage || e.message, hash: null });
    }
  };

  // 数据加载中显示提示
  if (!data) return <p>加载中…</p>;
  // 市场不存在时显示提示
  if (!market) return <p>市场不存在</p>;

  // 确定显示的市场状态（优先显示预言机结算中状态）
  const displayStatus = market.match?.status === 'ORACLE_PENDING' ? 'ORACLE_PENDING' : market.status;

  return (
    <div>
      {/* 市场问题作为标题 */}
      <h1>{market.question}</h1>
      {/* 关联赛事信息和市场状态徽章 */}
      {market.match && (
        <p>
          {market.match.home_team} vs {market.match.away_team}{' '}
          <MarketStatusBadge status={displayStatus} />
        </p>
      )}
      {/* 市场截止时间 */}
      <p>截止：{new Date(market.end_time).toLocaleString()}</p>
      {/* Yes/No 池金额和结果数量 */}
      <p>
        Yes 池 {formatUsdc(market.yes_pool)} · No 池 {formatUsdc(market.no_pool)}
        {market.outcome_count > 2 && ` · ${market.outcome_count} outcomes`}
      </p>
      {/* 显示 CPMM Yes 价格（来自链上） */}
      {pool?.priceYesBps != null && (
        <p>CPMM Yes 价格约 {(Number(pool.priceYesBps) / 100).toFixed(2)}%</p>
      )}
      {/* 显示 API 池快照 Yes 价格 */}
      {pool?.price_yes_bps && (
        <p>API 池快照 Yes {(Number(pool.price_yes_bps) / 100).toFixed(2)}%</p>
      )}
      {/* 受限市场提示：需要 VC 凭证才能参与 */}
      {access?.requires_vc && !access.allowed && (
        <p className="status-error">
          受限市场：需要 {access.credential_type} 凭证，请前往 DID 页登录并联系管理员签发。
        </p>
      )}

      {/* 下注表单：仅在市场开放且用户有权限时显示 */}
      {market.status === 'OPEN' && (!access?.requires_vc || access.allowed) && (
        <div className="card">
          <h2>下注</h2>
          {/* 结果选择下拉框 */}
          <label>
            结果{' '}
            <select value={outcome} onChange={(e) => setOutcome(Number(e.target.value))}>
              {Array.from({ length: market.outcome_count || 2 }, (_, i) => (
                <option key={i} value={i}>
                  Outcome {i}
                </option>
              ))}
            </select>
          </label>
          {/* 金额输入框 */}
          <label style={{ display: 'block', marginTop: '0.5rem' }}>
            金额 (mUSDC){' '}
            <input value={amount} onChange={(e) => setAmount(e.target.value)} type="number" min="1" />
          </label>
          {/* 授权并购买按钮 */}
          <button type="button" style={{ marginTop: '0.75rem' }} onClick={onBuy} disabled={!isConnected || !collateral}>
            Approve + Buy
          </button>
        </div>
      )}

      {/* 交易状态显示区域 */}
      <TxStatus {...tx} />
    </div>
  );
}
