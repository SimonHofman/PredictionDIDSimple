const LABELS = {
  OPEN: { text: '开放下注', className: 'badge-open' },
  RESOLVED: { text: '已结算', className: 'badge-resolved' },
  VOID: { text: '已作废 (可退款)', className: 'badge-void' },
  ORACLE_PENDING: { text: '预言机结算中', className: 'badge-pending' },
};

export default function MarketStatusBadge({ status }) {
  const m = LABELS[status] || { text: status, className: '' };
  return <span className={`badge ${m.className}`}>{m.text}</span>;
}
