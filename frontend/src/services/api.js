import { config } from '../config';

const TOKEN_KEY = 'prediction_jwt';

export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

async function request(path, options = {}) {
  const headers = {
    Accept: 'application/json',
    ...(options.body ? { 'Content-Type': 'application/json' } : {}),
    ...options.headers,
  };
  const token = getToken();
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  const res = await fetch(`${config.apiUrl}${path}`, { ...options, headers });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || `HTTP ${res.status}`);
  }
  return data;
}

export async function getHealth() {
  return request('/health');
}

export async function getReady() {
  const res = await fetch(`${config.apiUrl}/ready`);
  const data = await res.json();
  return { ok: res.ok, data };
}

export async function listMatches(params = {}) {
  const q = new URLSearchParams(params).toString();
  return request(`/matches?${q}`);
}

export async function getMatch(id) {
  return request(`/matches/${id}`);
}

export async function listMarkets(params = {}) {
  const q = new URLSearchParams(params).toString();
  return request(`/markets?${q}`);
}

export async function getMarket(id) {
  return request(`/markets/${id}`);
}

export async function siweAuth(message, signature) {
  return request('/auth/siwe', {
    method: 'POST',
    body: JSON.stringify({ message, signature }),
  });
}

export async function bindDid(did, signature) {
  return request('/users/bind-did', {
    method: 'POST',
    body: JSON.stringify({ did, signature }),
  });
}

export async function myPositions() {
  return request('/me/positions');
}

export async function myCredentials() {
  return request('/users/me/credentials');
}

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

function adminHeaders() {
  const key = sessionStorage.getItem('admin_key') || '';
  return { 'X-Admin-Key': key };
}

export async function adminListOracleJobs(status = '') {
  const q = status ? `?status=${status}` : '';
  return request(`/admin/oracle-jobs${q}`, { headers: adminHeaders() });
}

export async function adminRetryOracleJob(id) {
  return request(`/admin/oracle-jobs/${id}/retry`, {
    method: 'POST',
    headers: adminHeaders(),
  });
}

export async function adminRegisterMarket(body) {
  return request('/admin/markets', {
    method: 'POST',
    headers: adminHeaders(),
    body: JSON.stringify(body),
  });
}

export async function adminVoidMarket(id) {
  return request(`/admin/markets/${id}/void`, {
    method: 'POST',
    headers: adminHeaders(),
  });
}

export async function getCompliance() {
  return request('/compliance/restricted');
}

export async function getPlatformStats() {
  return request('/stats/platform');
}

export async function getMarketPool(id) {
  return request(`/markets/${id}/pool`);
}

export async function getMarketOrderbook(id) {
  return request(`/markets/${id}/orderbook`);
}
