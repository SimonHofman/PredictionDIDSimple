// 导入 React 核心库
import React from 'react';
// 导入 ReactDOM 客户端渲染 API
import ReactDOM from 'react-dom/client';
// 导入浏览器路由组件
import { BrowserRouter } from 'react-router-dom';
// 导入 Web3 提供者（钱包与区块链连接）
import Web3Provider from './providers/Web3Provider';
// 导入根应用组件
import App from './App';
// 导入国际化语言提供者
import { I18nProvider } from './i18n/index.jsx';
// 导入全局样式
import './index.css';

// 创建 React 根节点并渲染应用
ReactDOM.createRoot(document.getElementById('root')).render(
  // 严格模式：帮助检测潜在问题
  <React.StrictMode>
    {/* 国际化语言提供者，包裹整个应用 */}
    <I18nProvider>
      {/* Web3 提供者，提供钱包连接和区块链交互能力 */}
      <Web3Provider>
        {/* 浏览器路由，启用客户端路由 */}
        <BrowserRouter>
          {/* 应用根组件 */}
          <App />
        </BrowserRouter>
      </Web3Provider>
    </I18nProvider>
  </React.StrictMode>
);
