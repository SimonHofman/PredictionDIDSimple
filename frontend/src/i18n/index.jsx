import { createContext, useContext, useMemo, useState } from 'react';

const messages = {
  zh: {
    home: '首页',
    markets: '市场',
    me: '我的',
    did: 'DID',
    admin: '管理',
    stats: '统计',
    liquidity: '流动性',
    compliance: '合规',
    loading: '加载中…',
    restricted_title: '服务不可用',
    restricted_body: '您所在地区暂不支持本服务。',
    accept_risk: '我已阅读风险披露并确认非受限地区用户',
    continue: '继续',
    phase_footer: 'Phase 3 · CPMM / 多 outcome / 合规',
  },
  en: {
    home: 'Home',
    markets: 'Markets',
    me: 'Me',
    did: 'DID',
    admin: 'Admin',
    stats: 'Stats',
    liquidity: 'Liquidity',
    compliance: 'Compliance',
    loading: 'Loading…',
    restricted_title: 'Service unavailable',
    restricted_body: 'This service is not available in your region.',
    accept_risk: 'I accept the risk disclosure and confirm I am not in a restricted region',
    continue: 'Continue',
    phase_footer: 'Phase 3 · CPMM / multi-outcome / compliance',
  },
};

const I18nContext = createContext(null);

export function I18nProvider({ children }) {
  const [lang, setLang] = useState(() => localStorage.getItem('lang') || 'zh');
  const value = useMemo(() => {
    const t = (key) => messages[lang][key] || key;
    const setLanguage = (l) => {
      localStorage.setItem('lang', l);
      setLang(l);
    };
    return { lang, setLang: setLanguage, t };
  }, [lang]);
  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n() {
  const ctx = useContext(I18nContext);
  if (!ctx) throw new Error('useI18n outside provider');
  return ctx;
}
