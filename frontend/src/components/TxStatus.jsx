export default function TxStatus({ status, error, hash }) {
  if (!status && !error) return null;
  return (
    <div className="card" style={{ marginTop: '0.75rem' }}>
      {status === 'pending' && <p>交易提交中…</p>}
      {status === 'success' && <p className="status-ok">交易成功</p>}
      {status === 'error' && <p className="status-error">失败：{error}</p>}
      {hash && (
        <p style={{ fontSize: '0.8rem', wordBreak: 'break-all' }}>
          tx: {hash}
        </p>
      )}
    </div>
  );
}
