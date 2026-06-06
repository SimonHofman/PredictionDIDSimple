import { NavLink, Outlet } from 'react-router-dom';
import { config } from '../config';
import WalletBar from './WalletBar';
import { useI18n } from '../i18n/index.jsx';

export default function Layout() {
  const { t, lang, setLang } = useI18n();
  return (
    <div className="layout">
      <header className="header">
        <span className="brand">Prediction DID</span>
        <NavLink to="/" end className={({ isActive }) => (isActive ? 'active' : '')}>
          {t('home')}
        </NavLink>
        <NavLink to="/markets" className={({ isActive }) => (isActive ? 'active' : '')}>
          {t('markets')}
        </NavLink>
        <NavLink to="/stats" className={({ isActive }) => (isActive ? 'active' : '')}>
          {t('stats')}
        </NavLink>
        <NavLink to="/liquidity" className={({ isActive }) => (isActive ? 'active' : '')}>
          {t('liquidity')}
        </NavLink>
        <NavLink to="/me" className={({ isActive }) => (isActive ? 'active' : '')}>
          {t('me')}
        </NavLink>
        <NavLink to="/did" className={({ isActive }) => (isActive ? 'active' : '')}>
          {t('did')}
        </NavLink>
        <NavLink to="/admin/oracle" className={({ isActive }) => (isActive ? 'active' : '')}>
          {t('admin')}
        </NavLink>
        <button type="button" className="lang-toggle" onClick={() => setLang(lang === 'zh' ? 'en' : 'zh')}>
          {lang === 'zh' ? 'EN' : '中文'}
        </button>
        <WalletBar />
      </header>
      <main className="main">
        <Outlet />
      </main>
      <footer className="footer">
        {t('phase_footer')} · CHAIN_ID={config.chainId}
      </footer>
    </div>
  );
}
