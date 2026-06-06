import { useEffect, useState } from 'react';
import { adminListOracleJobs, adminRetryOracleJob } from '../../services/api';

export default function OracleJobs() {
  const [items, setItems] = useState([]);
  const [error, setError] = useState(null);

  const load = () => {
    adminListOracleJobs()
      .then((r) => setItems(r.items || []))
      .catch((e) => setError(e.message));
  };

  useEffect(() => {
    load();
    const id = setInterval(load, 10000);
    return () => clearInterval(id);
  }, []);

  return (
    <div>
      <h2>Oracle 任务队列</h2>
      {error && <p className="status-error">{error}</p>}
      {items.map((j) => (
        <div key={j.id} className="card">
          <p>
            #{j.id} · {j.status} · {j.question}
          </p>
          <p style={{ fontSize: '0.85rem' }}>{j.market_address}</p>
          {j.primary_home != null && (
            <p>
              主源 {j.primary_home}-{j.primary_away} / 备源 {j.secondary_home}-{j.secondary_away}
            </p>
          )}
          {j.error_message && <p className="status-error">{j.error_message}</p>}
          {j.status === 'manual_review' && (
            <button type="button" onClick={() => adminRetryOracleJob(j.id).then(load)}>
              重试
            </button>
          )}
        </div>
      ))}
    </div>
  );
}
