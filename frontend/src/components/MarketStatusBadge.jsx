// 市场状态标签映射表：将状态码映射为中文显示文本和样式类名
const LABELS = {
  // 开放下注状态
  OPEN: { text: '开放下注', className: 'badge-open' },
  // 已结算状态
  RESOLVED: { text: '已结算', className: 'badge-resolved' },
  // 已作废状态（用户可退款）
  VOID: { text: '已作废 (可退款)', className: 'badge-void' },
  // 预言机结算中状态
  ORACLE_PENDING: { text: '预言机结算中', className: 'badge-pending' },
};

/**
 * 市场状态徽章组件
 * 根据市场状态码显示对应的中文标签和颜色样式
 * @param {{ status: string }} props - 市场状态码
 */
export default function MarketStatusBadge({ status }) {
  // 查找对应状态的显示配置，若未找到则使用原始状态文本
  const m = LABELS[status] || { text: status, className: '' };
  // 渲染带样式的徽章标签
  return <span className={`badge ${m.className}`}>{m.text}</span>;
}
