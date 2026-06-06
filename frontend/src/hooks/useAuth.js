// 导入 React 钩子函数
import { useCallback, useState } from 'react';
// 导入 wagmi 钱包账户和签名相关钩子
import { useAccount, useSignMessage } from 'wagmi';
// 导入 SIWE（以太坊签名登录）消息构造类
import { SiweMessage } from 'siwe';
// 导入应用配置
import { config } from '../config';
// 导入 API 服务方法：绑定 DID、清除令牌、获取令牌、设置令牌、SIWE 认证
import { bindDid, clearToken, getToken, setToken, siweAuth } from '../services/api';

/**
 * 自定义认证钩子
 * 封装了钱包登录（SIWE）、登出和认证状态管理
 */
export function useAuth() {
  // 获取当前钱包地址、链 ID 和连接状态
  const { address, chainId, isConnected } = useAccount();
  // 获取异步签名消息方法
  const { signMessageAsync } = useSignMessage();
  // 登录加载状态
  const [loading, setLoading] = useState(false);
  // 错误信息状态
  const [error, setError] = useState(null);
  // 从本地存储获取 JWT 令牌
  const token = getToken();

  // 登录方法：使用 SIWE 协议进行以太坊签名登录
  const login = useCallback(async () => {
    // 如果钱包未连接或链 ID 缺失，则直接返回
    if (!address || !chainId) return;
    // 设置加载中状态
    setLoading(true);
    // 清除之前的错误
    setError(null);
    try {
      // 构造 SIWE 消息对象
      const message = new SiweMessage({
        // 签名登录的域名
        domain: config.siweDomain,
        // 钱包地址
        address,
        // 登录声明文本
        statement: 'Sign in to Prediction DID World Cup',
        // URI 地址
        uri: config.siweUri,
        // SIWE 版本号
        version: '1',
        // 区块链网络 ID
        chainId,
        // 随机数，防止重放攻击
        nonce: Math.random().toString(36).slice(2),
      });
      // 将消息对象转为签名所需的字符串格式
      const prepared = message.prepareMessage();
      // 请求用户钱包签名
      const signature = await signMessageAsync({ message: prepared });
      // 调用后端 SIWE 认证接口，获取 JWT 令牌
      const res = await siweAuth(prepared, signature);
      // 将 JWT 令牌保存到本地存储
      setToken(res.token);
      // 构造 DID 标识符（去中心化身份）
      const did = `did:pkh:eip155:${chainId}:${address.toLowerCase()}`;
      try {
        // 尝试绑定 DID 到用户账户
        await bindDid(did, signature);
      } catch {
        // bind optional if already bound
        // 绑定失败时忽略（可能已绑定过）
      }
      // 刷新页面以应用登录状态
      window.location.reload();
    } catch (e) {
      // 捕获错误并设置错误信息
      setError(e.message);
    } finally {
      // 无论成功或失败都结束加载状态
      setLoading(false);
    }
  }, [address, chainId, signMessageAsync]);

  // 登出方法：清除令牌并刷新页面
  const logout = useCallback(() => {
    // 清除本地存储的 JWT 令牌
    clearToken();
    // 刷新页面以重置应用状态
    window.location.reload();
  }, []);

  // 返回认证相关的状态和方法
  return {
    // 钱包是否已连接
    isConnected,
    // 当前钱包地址
    address,
    // JWT 令牌
    token,
    // 是否已通过认证（有有效令牌）
    isAuthenticated: Boolean(token),
    // 登录方法
    login,
    // 登出方法
    logout,
    // 加载状态
    loading,
    // 错误信息
    error,
  };
}
