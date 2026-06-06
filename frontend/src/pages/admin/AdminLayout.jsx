// 导入路由导航链接和子路由出口组件
import { NavLink, Outlet } from 'react-router-dom';

/**
 * 管理后台布局组件
 * 包含管理员认证提示和子路由导航
 */
export default function AdminLayout() {
  // 从 sessionStorage 获取管理员 API Key
  const key = sessionStorage.getItem('admin_key') || '';
  return (
    <div>
      {/* 页面标题 */}
      <h1>管理后台</h1>
      {/* 如果未设置管理员 Key，显示设置提示 */}
      {!key && (
        <p className="status-error">
          请先在控制台执行：sessionStorage.setItem(&apos;admin_key&apos;, &apos;你的ADMIN_API_KEY&apos;)
        </p>
      )}
      {/* 管理后台子路由导航 */}
      <nav style={{ display: 'flex', gap: '1rem', marginBottom: '1rem' }}>
        {/* Oracle 任务队列导航链接 */}
        <NavLink to="/admin/oracle">Oracle 队列</NavLink>
        {/* 市场配置导航链接 */}
        <NavLink to="/admin/markets">市场配置</NavLink>
      </nav>
      {/* 渲染管理后台子路由内容 */}
      <Outlet />
    </div>
  );
}
