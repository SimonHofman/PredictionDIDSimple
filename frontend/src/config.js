const apiUrl = import.meta.env.VITE_API_URL;
const chainIdRaw = import.meta.env.VITE_CHAIN_ID;

export const config = {
  apiUrl: apiUrl || 'http://localhost:8080',
  chainId: chainIdRaw ? Number(chainIdRaw) : 31337,
  apiUrlConfigured: Boolean(apiUrl),
  mockUsdc: import.meta.env.VITE_MOCK_USDC_ADDRESS || '',
  marketFactory: import.meta.env.VITE_MARKET_FACTORY_ADDRESS || '',
  siweDomain: import.meta.env.VITE_SIWE_DOMAIN || 'localhost',
  siweUri: import.meta.env.VITE_SIWE_URI || 'http://localhost:5173',
};
