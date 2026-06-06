import { NavLink, Outlet } from 'react-router-dom';

export default function AdminLayout() {
  const key = sessionStorage.getItem('admin_key') || '';
  return (
    <div>
      <h1>管理后台</h1>
      {!key && (
        <p className="status-error">
          请先在控制台执行：sessionStorage.setItem(&apos;admin_key&apos;, &apos;你的ADMIN_API_KEY&apos;)
        </p>
      )}
      <nav style={{ display: 'flex', gap: '1rem', marginBottom: '1rem' }}>
        <NavLink to="/admin/oracle">Oracle 队列</NavLink>
        <NavLink to="/admin/markets">市场配置</NavLink>
      </nav>
      <Outlet />
    </div>
  );
}
