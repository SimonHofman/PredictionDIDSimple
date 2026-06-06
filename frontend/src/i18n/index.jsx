// 导入 React 钩子：上下文、useMemo 缓存、状态管理
import { createContext, useContext, useMemo, useState } from 'react';

// 国际化翻译消息定义（支持中文和英文）
const messages = {
  // 中文翻译
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
  // 英文翻译
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

// 创建国际化上下文对象
const I18nContext = createContext(null);

/**
 * 国际化提供者组件
 * 管理语言状态并向子组件提供翻译函数
 */
export function I18nProvider({ children }) {
  // 从本地存储读取语言偏好，默认为中文
  const [lang, setLang] = useState(() => localStorage.getItem('lang') || 'zh');
  // 使用 useMemo 缓存上下文值，避免不必要的重新渲染
  const value = useMemo(() => {
    // 翻译函数：根据当前语言和键名获取翻译文本
    const t = (key) => messages[lang][key] || key;
    // 切换语言方法：保存到本地存储并更新状态
    const setLanguage = (l) => {
      localStorage.setItem('lang', l);
      setLang(l);
    };
    // 返回语言状态、切换方法和翻译函数
    return { lang, setLang: setLanguage, t };
  }, [lang]);
  // 通过 Context Provider 向子组件提供国际化能力
  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

/**
 * 国际化消费钩子
 * 在组件中获取翻译函数和语言状态
 * @returns {{ lang: string, setLang: Function, t: Function }}
 */
export function useI18n() {
  // 获取上下文值
  const ctx = useContext(I18nContext);
  // 如果在 Provider 外部使用则抛出错误
  if (!ctx) throw new Error('useI18n outside provider');
  return ctx;
}
