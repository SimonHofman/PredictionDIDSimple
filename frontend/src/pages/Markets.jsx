import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { listMarkets } from '../services/api';
import { formatUsdc } from '../services/contracts';

export default function Markets() {
  const [items, setItems] = useState([]);
  const [error, setError] = useState(null);

  useEffect(() => {
    listMarkets({ limit: 50 })
      .then((res) => setItems(res.items || []))
      .catch((e) => setError(e.message));
  }, []);

  return (
    <div>
      <h1>预测市场</h1>
      {error && <p className="status-error">{error}</p>}
      {items.length === 0 && <p>暂无市场。部署合约后运行 seed-markets，并确保 Indexer 运行。</p>}
      {items.map((mk) => (
        <Link key={mk.id} to={`/markets/${mk.id}`} className="card link-card">
          <p>{mk.question}</p>
          {mk.match && (
            <p>
              {mk.match.home_team} vs {mk.match.away_team}
            </p>
          )}
          <p>
            Yes 池 {formatUsdc(mk.yes_pool)} / No 池 {formatUsdc(mk.no_pool)} · {mk.status}
          </p>
        </Link>
      ))}
    </div>
  );
}
