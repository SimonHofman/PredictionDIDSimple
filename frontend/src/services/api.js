// 导入应用配置（API 基础地址）
import { config } from '../config';

// JWT 令牌在 localStorage 中的存储键名
const TOKEN_KEY = 'prediction_jwt';

// 从本地存储获取 JWT 令牌
export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

// 将 JWT 令牌保存到本地存储
export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token);
}

// 从本地存储中删除 JWT 令牌（登出时使用）
export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

/**
 * 通用 HTTP 请求封装函数
 * 自动处理请求头（Content-Type、Authorization）和错误响应
 * @param {string} path - API 路径
 * @param {object} options - fetch 请求选项
 * @returns {Promise<object>} 响应 JSON 数据
 */
async function request(path, options = {}) {
  // 构造请求头
  const headers = {
    // 接受 JSON 格式响应
    Accept: 'application/json',
    // 如果有请求体，设置 Content-Type 为 JSON
    ...(options.body ? { 'Content-Type': 'application/json' } : {}),
    // 合并自定义请求头
    ...options.headers,
  };
  // 获取本地存储的 JWT 令牌
  const token = getToken();
  // 如果有令牌，添加 Authorization 请求头
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  // 发送 HTTP 请求
  const res = await fetch(`${config.apiUrl}${path}`, { ...options, headers });
  // 解析响应 JSON，失败时返回空对象
  const data = await res.json().catch(() => ({}));
  // 如果响应状态码不是 2xx，抛出错误
  if (!res.ok) {
    throw new Error(data.error || `HTTP ${res.status}`);
  }
  // 返回解析后的数据
  return data;
}

// 健康检查接口：检测后端服务是否正常运行
export async function getHealth() {
  return request('/health');
}

// 就绪检查接口：检测后端服务及依赖是否准备就绪
export async function getReady() {
  const res = await fetch(`${config.apiUrl}/ready`);
  const data = await res.json();
  return { ok: res.ok, data };
}

// 获取赛事列表（支持分页等参数）
export async function listMatches(params = {}) {
  // 将参数对象转换为 URL 查询字符串
  const q = new URLSearchParams(params).toString();
  return request(`/matches?${q}`);
}

// 根据 ID 获取单个赛事详情
export async function getMatch(id) {
  return request(`/matches/${id}`);
}

// 获取预测市场列表（支持分页、状态过滤等参数）
export async function listMarkets(params = {}) {
  // 将参数对象转换为 URL 查询字符串
  const q = new URLSearchParams(params).toString();
  return request(`/markets?${q}`);
}

// 根据 ID 获取单个市场详情
export async function getMarket(id) {
  return request(`/markets/${id}`);
}

// SIWE 认证接口：发送签名消息和签名到后端进行验证，返回 JWT 令牌
export async function siweAuth(message, signature) {
  return request('/auth/siwe', {
    method: 'POST',
    body: JSON.stringify({ message, signature }),
  });
}

// 绑定 DID 接口：将去中心化身份绑定到用户账户
export async function bindDid(did, signature) {
  return request('/users/bind-did', {
    method: 'POST',
    body: JSON.stringify({ did, signature }),
  });
}

// 获取当前用户的持仓列表
export async function myPositions() {
  return request('/me/positions');
}

// 获取当前用户的可验证凭证（VC）列表
export async function myCredentials() {
  return request('/users/me/credentials');
}

// 验证可验证凭证（VC）：提交 VC JSON、凭证类型和地区进行验证
export async function verifyVC(vcJson, credentialType, region) {
  return request('/auth/verify-vc', {
    method: 'POST',
    body: JSON.stringify({
      vc_json: vcJson,
      credential_type: credentialType,
      region,
    }),
  });
}

// 获取管理员 API Key 请求头（从 sessionStorage 中读取）
function adminHeaders() {
  const key = sessionStorage.getItem('admin_key') || '';
  return { 'X-Admin-Key': key };
}

// 管理员接口：获取预言机任务列表（可按状态过滤）
export async function adminListOracleJobs(status = '') {
  const q = status ? `?status=${status}` : '';
  return request(`/admin/oracle-jobs${q}`, { headers: adminHeaders() });
}

// 管理员接口：重试失败的预言机任务
export async function adminRetryOracleJob(id) {
  return request(`/admin/oracle-jobs/${id}/retry`, {
    method: 'POST',
    headers: adminHeaders(),
  });
}

// 管理员接口：注册/更新市场元数据
export async function adminRegisterMarket(body) {
  return request('/admin/markets', {
    method: 'POST',
    headers: adminHeaders(),
    body: JSON.stringify(body),
  });
}

// 管理员接口：作废市场（用户可退款）
export async function adminVoidMarket(id) {
  return request(`/admin/markets/${id}/void`, {
    method: 'POST',
    headers: adminHeaders(),
  });
}

// 获取合规限制信息（检查用户是否在受限地区）
export async function getCompliance() {
  return request('/compliance/restricted');
}

// 获取平台统计数据（成交量、用户数等）
export async function getPlatformStats() {
  return request('/stats/platform');
}

// 获取市场流动性池状态
export async function getMarketPool(id) {
  return request(`/markets/${id}/pool`);
}

// 获取市场订单簿数据
export async function getMarketOrderbook(id) {
  return request(`/markets/${id}/orderbook`);
}
