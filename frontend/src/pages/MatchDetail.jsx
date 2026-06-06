import { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { getMatch } from '../services/api';

export default function MatchDetail() {
  const { id } = useParams();
  const [data, setData] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    getMatch(id)
      .then(setData)
      .catch((e) => setError(e.message));
  }, [id]);

  if (error) return <p className="status-error">{error}</p>;
  if (!data) return <p>加载中…</p>;

  const m = data.match;
  return (
    <div>
      <h1>
        {m.home_team} vs {m.away_team}
      </h1>
      <p>开球：{new Date(m.kickoff_at).toLocaleString()}</p>
      <p>状态：{m.status}</p>
      {m.home_score != null && (
        <p>
          比分 {m.home_score} - {m.away_score}
        </p>
      )}

      <h2>关联市场</h2>
      {(data.markets || []).length === 0 && <p>暂无市场</p>}
      {(data.markets || []).map((mk) => (
        <Link key={mk.id} to={`/markets/${mk.id}`} className="card link-card">
          {mk.question} · {mk.status}
        </Link>
      ))}
    </div>
  );
}
