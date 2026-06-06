import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getHealth, listMatches } from '../services/api';

export default function Home() {
  const [apiStatus, setApiStatus] = useState('checking');
  const [matches, setMatches] = useState([]);
  const [error, setError] = useState(null);

  useEffect(() => {
    async function load() {
      try {
        await getHealth();
        setApiStatus('online');
        const res = await listMatches({ limit: 10 });
        setMatches(res.items || []);
      } catch (e) {
        setApiStatus('offline');
        setError(e.message);
      }
    }
    load();
  }, []);

  const upcoming = matches.filter((m) => ['SCHEDULED', 'LIVE'].includes(m.status));

  return (
    <div>
      <h1>世界杯预测市场</h1>
      <p className={apiStatus === 'online' ? 'status-ok' : 'status-error'}>
        API {apiStatus === 'online' ? '在线' : `离线${error ? `: ${error}` : ''}`}
      </p>

      <h2>即将开始 / 进行中</h2>
      {upcoming.length === 0 && <p>暂无赛事，请先运行后端 seed/sync。</p>}
      <div className="grid">
        {upcoming.map((m) => (
          <Link key={m.id} to={`/matches/${m.id}`} className="card link-card">
            <strong>
              {m.home_team} vs {m.away_team}
            </strong>
            <p>{new Date(m.kickoff_at).toLocaleString()}</p>
            <span className="badge">{m.status}</span>
            {m.home_score != null && (
              <p>
                比分 {m.home_score} - {m.away_score}
              </p>
            )}
          </Link>
        ))}
      </div>
    </div>
  );
}
