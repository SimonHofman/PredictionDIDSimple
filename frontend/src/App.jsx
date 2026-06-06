// 导入路由组件
import { Route, Routes } from 'react-router-dom';
// 导入页面布局组件
import Layout from './components/Layout';
// 导入首页组件
import Home from './pages/Home';
// 导入市场列表页组件
import Markets from './pages/Markets';
// 导入个人中心页组件
import Me from './pages/Me';
// 导入赛事详情页组件
import MatchDetail from './pages/MatchDetail';
// 导入市场详情页组件
import MarketDetail from './pages/MarketDetail';
// 导入 DID 身份档案页组件
import DIDProfile from './pages/DIDProfile';
// 导入管理后台布局组件
import AdminLayout from './pages/admin/AdminLayout';
// 导入预言机任务管理页组件
import OracleJobs from './pages/admin/OracleJobs';
// 导入管理后台市场配置页组件
import AdminMarkets from './pages/admin/AdminMarkets';
// 导入平台统计页组件
import Stats from './pages/Stats';
// 导入流动性管理页组件
import Liquidity from './pages/Liquidity';
// 导入合规检查包裹组件
import ComplianceWrapper from './components/ComplianceWrapper';

// 应用根组件，定义所有路由规则
export default function App() {
  return (
    // 路由配置容器
    <Routes>
      {/* 合规检查包裹层：用户必须通过合规确认才能访问 */}
      <Route element={<ComplianceWrapper />}>
      {/* 通用页面布局（含导航栏和页脚） */}
      <Route element={<Layout />}>
        {/* 首页路由 */}
        <Route index element={<Home />} />
        {/* 市场列表页路由 */}
        <Route path="markets" element={<Markets />} />
        {/* 市场详情页路由（带动态 ID 参数） */}
        <Route path="markets/:id" element={<MarketDetail />} />
        {/* 赛事详情页路由（带动态 ID 参数） */}
        <Route path="matches/:id" element={<MatchDetail />} />
        {/* 个人中心页路由 */}
        <Route path="me" element={<Me />} />
        {/* DID 身份档案页路由 */}
        <Route path="did" element={<DIDProfile />} />
        {/* 平台统计页路由 */}
        <Route path="stats" element={<Stats />} />
        {/* 流动性管理页路由 */}
        <Route path="liquidity" element={<Liquidity />} />
        {/* 管理后台嵌套路由 */}
        <Route path="admin" element={<AdminLayout />}>
          {/* 预言机任务管理页 */}
          <Route path="oracle" element={<OracleJobs />} />
          {/* 管理后台市场配置页 */}
          <Route path="markets" element={<AdminMarkets />} />
        </Route>
      </Route>
      </Route>
    </Routes>
  );
}
