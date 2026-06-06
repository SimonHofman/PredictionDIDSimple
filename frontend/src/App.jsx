import { Route, Routes } from 'react-router-dom';
import Layout from './components/Layout';
import Home from './pages/Home';
import Markets from './pages/Markets';
import Me from './pages/Me';
import MatchDetail from './pages/MatchDetail';
import MarketDetail from './pages/MarketDetail';
import DIDProfile from './pages/DIDProfile';
import AdminLayout from './pages/admin/AdminLayout';
import OracleJobs from './pages/admin/OracleJobs';
import AdminMarkets from './pages/admin/AdminMarkets';
import Stats from './pages/Stats';
import Liquidity from './pages/Liquidity';
import ComplianceWrapper from './components/ComplianceWrapper';

export default function App() {
  return (
    <Routes>
      <Route element={<ComplianceWrapper />}>
      <Route element={<Layout />}>
        <Route index element={<Home />} />
        <Route path="markets" element={<Markets />} />
        <Route path="markets/:id" element={<MarketDetail />} />
        <Route path="matches/:id" element={<MatchDetail />} />
        <Route path="me" element={<Me />} />
        <Route path="did" element={<DIDProfile />} />
        <Route path="stats" element={<Stats />} />
        <Route path="liquidity" element={<Liquidity />} />
        <Route path="admin" element={<AdminLayout />}>
          <Route path="oracle" element={<OracleJobs />} />
          <Route path="markets" element={<AdminMarkets />} />
        </Route>
      </Route>
      </Route>
    </Routes>
  );
}
