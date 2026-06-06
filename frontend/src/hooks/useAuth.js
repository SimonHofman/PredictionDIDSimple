import { useCallback, useState } from 'react';
import { useAccount, useSignMessage } from 'wagmi';
import { SiweMessage } from 'siwe';
import { config } from '../config';
import { bindDid, clearToken, getToken, setToken, siweAuth } from '../services/api';

export function useAuth() {
  const { address, chainId, isConnected } = useAccount();
  const { signMessageAsync } = useSignMessage();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const token = getToken();

  const login = useCallback(async () => {
    if (!address || !chainId) return;
    setLoading(true);
    setError(null);
    try {
      const message = new SiweMessage({
        domain: config.siweDomain,
        address,
        statement: 'Sign in to Prediction DID World Cup',
        uri: config.siweUri,
        version: '1',
        chainId,
        nonce: Math.random().toString(36).slice(2),
      });
      const prepared = message.prepareMessage();
      const signature = await signMessageAsync({ message: prepared });
      const res = await siweAuth(prepared, signature);
      setToken(res.token);
      const did = `did:pkh:eip155:${chainId}:${address.toLowerCase()}`;
      try {
        await bindDid(did, signature);
      } catch {
        // bind optional if already bound
      }
      window.location.reload();
    } catch (e) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, [address, chainId, signMessageAsync]);

  const logout = useCallback(() => {
    clearToken();
    window.location.reload();
  }, []);

  return {
    isConnected,
    address,
    token,
    isAuthenticated: Boolean(token),
    login,
    logout,
    loading,
    error,
  };
}
