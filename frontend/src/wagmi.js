import { http, createConfig } from 'wagmi';
import { injected } from 'wagmi/connectors';
import { defineChain } from 'viem';
import { config as appConfig } from './config';

export const hardhatLocal = defineChain({
  id: appConfig.chainId,
  name: 'Hardhat Local',
  nativeCurrency: { name: 'Ether', symbol: 'ETH', decimals: 18 },
  rpcUrls: {
    default: { http: ['http://127.0.0.1:8545'] },
  },
});

export const wagmiConfig = createConfig({
  chains: [hardhatLocal],
  connectors: [injected()],
  transports: {
    [hardhatLocal.id]: http(),
  },
});
