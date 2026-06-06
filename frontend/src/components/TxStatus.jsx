/**
 * 交易状态显示组件
 * 根据交易的不同阶段显示相应的状态信息
 * @param {{ status: string|null, error: string|null, hash: string|null }} props
 */
export default function TxStatus({ status, error, hash }) {
  // 如果没有状态也没有错误，不渲染任何内容
  if (!status && !error) return null;
  return (
    <div className="card" style={{ marginTop: '0.75rem' }}>
      {/* 交易提交中：显示等待提示 */}
      {status === 'pending' && <p>交易提交中…</p>}
      {/* 交易成功：显示成功提示 */}
      {status === 'success' && <p className="status-ok">交易成功</p>}
      {/* 交易失败：显示错误信息 */}
      {status === 'error' && <p className="status-error">失败：{error}</p>}
      {/* 如果有交易哈希，显示交易哈希值 */}
      {hash && (
        <p style={{ fontSize: '0.8rem', wordBreak: 'break-all' }}>
          tx: {hash}
        </p>
      )}
    </div>
  );
}
