// 导入路由导航链接和子路由出口组件
import { NavLink, Outlet } from 'react-router-dom';
// 导入应用配置（获取链 ID 等信息）
import { config } from '../config';
// 导入钱包栏组件
import WalletBar from './WalletBar';
// 导入国际化翻译钩子
import { useI18n } from '../i18n/index.jsx';

/**
 * 页面布局组件
 * 包含顶部导航栏、主内容区域和底部页脚
 */
export default function Layout() {
  // 获取翻译函数、当前语言和语言切换方法
  const { t, lang, setLang } = useI18n();
  return (
    <div className="layout">
      {/* 顶部导航栏 */}
      <header className="header">
        {/* 品牌名称 */}
        <span className="brand">Prediction DID</span>
        {/* 首页导航链接 */}
        <NavLink to="/" end className={({ isActive }) => (isActive ? 'active' : '')}>
          {t('home')}
        </NavLink>
        {/* 市场列表导航链接 */}
        <NavLink to="/markets" className={({ isActive }) => (isActive ? 'active' : '')}>
          {t('markets')}
        </NavLink>
        {/* 统计页导航链接 */}
        <NavLink to="/stats" className={({ isActive }) => (isActive ? 'active' : '')}>
          {t('stats')}
        </NavLink>
        {/* 流动性管理导航链接 */}
        <NavLink to="/liquidity" className={({ isActive }) => (isActive ? 'active' : '')}>
          {t('liquidity')}
        </NavLink>
        {/* 个人中心导航链接 */}
        <NavLink to="/me" className={({ isActive }) => (isActive ? 'active' : '')}>
          {t('me')}
        </NavLink>
        {/* DID 身份导航链接 */}
        <NavLink to="/did" className={({ isActive }) => (isActive ? 'active' : '')}>
          {t('did')}
        </NavLink>
        {/* 管理后台导航链接 */}
        <NavLink to="/admin/oracle" className={({ isActive }) => (isActive ? 'active' : '')}>
          {t('admin')}
        </NavLink>
        {/* 语言切换按钮：中英文互切 */}
        <button type="button" className="lang-toggle" onClick={() => setLang(lang === 'zh' ? 'en' : 'zh')}>
          {lang === 'zh' ? 'EN' : '中文'}
        </button>
        {/* 钱包连接和登录状态栏 */}
        <WalletBar />
      </header>
      {/* 主内容区域：渲染子路由页面 */}
      <main className="main">
        <Outlet />
      </main>
      {/* 底部页脚：显示版本信息和链 ID */}
      <footer className="footer">
        {t('phase_footer')} · CHAIN_ID={config.chainId}
      </footer>
    </div>
  );
}
