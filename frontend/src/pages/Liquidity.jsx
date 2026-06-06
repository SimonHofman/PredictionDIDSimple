// 导入 React 钩子：副作用和状态管理
import { useEffect, useState } from 'react';
// 导入 wagmi 钱包账户钩子
import { useAccount } from 'wagmi';
// 导入 API 服务：市场列表和池数据
import { listMarkets, getMarketPool } from '../services/api';
// 导入合约交互方法
import {
  addLiquidityV3, // V3 市场添加流动性
  approveUsdc,    // USDC 代币授权
  formatUsdc,     // USDC 金额格式化
  parseUsdc,      // USDC 金额解析
} from '../services/contracts';

/**
 * 流动性管理页组件
 * 允许用户向 V3 CPMM 二元市场注入流动性
 */
export default function Liquidity() {
  // 获取钱包连接状态
  const { isConnected } = useAccount();
  // 可用的开放市场列表
  const [markets, setMarkets] = useState([]);
  // 当前选中的市场 ID
  const [selected, setSelected] = useState('');
  // 选中市场的池状态数据
  const [pool, setPool] = useState(null);
  // 注入金额输入值
  const [amount, setAmount] = useState('100');
  // 交易状态信息
  const [tx, setTx] = useState(null);

  // 组件挂载时加载所有开放状态的市场
  useEffect(() => {
    listMarkets({ status: 'OPEN' }).then((res) => {
      const rows = res.items || [];
      // 设置市场列表
      setMarkets(rows);
      // 默认选中第一个市场
      if (rows[0]) setSelected(String(rows[0].id));
    });
  }, []);

  // 切换选中市场时加载对应的池数据
  useEffect(() => {
    if (!selected) return;
    // 获取选中市场的池状态
    getMarketPool(selected).then(setPool).catch(console.error);
  }, [selected]);

  // 添加流动性操作处理方法
  const onAdd = async () => {
    // 查找选中的市场对象
    const m = markets.find((x) => String(x.id) === selected);
    // 前提条件检查：市场存在且钱包已连接
    if (!m || !isConnected) return;
    // 获取 USDC 代币合约地址
    const collateral = import.meta.env.VITE_MOCK_USDC_ADDRESS;
    if (!collateral) return;
    // 设置交易状态为等待中
    setTx('pending');
    try {
      // 将输入金额解析为链上整数
      const amt = parseUsdc(amount);
      // 先授权 USDC 代币给市场合约
      await approveUsdc(collateral, m.market_address, amt);
      // 调用添加流动性方法
      await addLiquidityV3(m.market_address, amt);
      // 交易成功
      setTx('ok');
      // 刷新池数据
      getMarketPool(selected).then(setPool);
    } catch (e) {
      // 交易失败，显示错误信息
      setTx(e.message);
    }
  };

  return (
    <div>
      {/* 页面标题 */}
      <h1>流动性 (CPMM)</h1>
      {/* 功能说明 */}
      <p>向 V3 二元市场注入流动性；需链上合约为 PredictionMarketV3。</p>
      {/* 市场选择下拉框 */}
      <label>
        市场{' '}
        <select value={selected} onChange={(e) => setSelected(e.target.value)}>
          {markets.map((m) => (
            <option key={m.id} value={m.id}>
              #{m.id} {m.question}
            </option>
          ))}
        </select>
      </label>
      {/* 显示当前池状态信息 */}
      {pool && (
        <p>
          类型 {pool.market_type} · Yes 储备 {formatUsdc(pool.reserve_yes)} · No 储备{' '}
          {formatUsdc(pool.reserve_no)} · 价格 Yes {pool.price_yes_bps || '—'} bps
        </p>
      )}
      {/* 注入金额输入框 */}
      <label style={{ display: 'block', marginTop: '0.5rem' }}>
        注入金额 (mUSDC){' '}
        <input type="number" value={amount} onChange={(e) => setAmount(e.target.value)} />
      </label>
      {/* 授权并添加流动性按钮 */}
      <button type="button" style={{ marginTop: '0.75rem' }} onClick={onAdd} disabled={!isConnected}>
        Approve + Add Liquidity
      </button>
      {/* 交易状态提示 */}
      {tx && <p>{tx}</p>}
    </div>
  );
}
