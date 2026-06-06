// 导入 viem 的金额解析和格式化工具
import { parseUnits, formatUnits } from 'viem';
// 导入 wagmi 核心的合约读写和交易回执等待方法
import { readContract, writeContract, waitForTransactionReceipt } from '@wagmi/core';
// 导入 wagmi 配置实例
import { wagmiConfig } from '../wagmi';
// 导入 MockUSDC 合约 ABI
import MockUSDC from '../abis/MockUSDC.json';
// 导入二元预测市场合约 ABI（V1 版本）
import PredictionMarket from '../abis/PredictionMarket.json';
// 导入二元预测市场合约 ABI（V3 CPMM 版本）
import PredictionMarketV3 from '../abis/PredictionMarketV3.json';
// 导入多结果预测市场合约 ABI
import MultiOutcomeMarket from '../abis/MultiOutcomeMarket.json';

// USDC 代币的小数位数（6 位）
const USDC_DECIMALS = 6;

/**
 * 将人类可读的 USDC 金额解析为链上整数（含 6 位小数）
 * @param {string|number} amount - 人类可读的金额
 * @returns {bigint} 链上整数金额
 */
export function parseUsdc(amount) {
  return parseUnits(String(amount), USDC_DECIMALS);
}

/**
 * 将链上的 USDC 整数金额格式化为人类可读的字符串
 * @param {bigint|string|number} amount - 链上整数金额
 * @returns {string} 格式化后的金额字符串
 */
export function formatUsdc(amount) {
  return formatUnits(BigInt(amount || 0), USDC_DECIMALS);
}

/**
 * 读取指定地址的 USDC 余额
 * @param {string} address - 要查询余额的钱包地址
 * @param {string} tokenAddress - USDC 代币合约地址
 * @returns {Promise<bigint>} 余额数值
 */
export async function readUsdcBalance(address, tokenAddress) {
  return readContract(wagmiConfig, {
    address: tokenAddress,
    abi: MockUSDC.abi,
    functionName: 'balanceOf',
    args: [address],
  });
}

/**
 * 授权 USDC 代币给指定的合约地址（用于下注前的授权）
 * @param {string} tokenAddress - USDC 代币合约地址
 * @param {string} spender - 被授权的合约地址（市场合约）
 * @param {bigint} amount - 授权金额
 * @returns {Promise<object>} 交易回执
 */
export async function approveUsdc(tokenAddress, spender, amount) {
  // 调用 ERC20 approve 方法
  const hash = await writeContract(wagmiConfig, {
    address: tokenAddress,
    abi: MockUSDC.abi,
    functionName: 'approve',
    args: [spender, amount],
  });
  // 等待交易确认并返回回执
  return waitForTransactionReceipt(wagmiConfig, { hash });
}

/**
 * 在 V1 二元市场中购买指定结果
 * @param {string} marketAddress - 市场合约地址
 * @param {number} outcome - 结果编号（0=Yes, 1=No）
 * @param {bigint} amount - 购买金额
 * @returns {Promise<object>} 交易回执
 */
export async function buyOutcome(marketAddress, outcome, amount) {
  // 调用市场合约的 buy 方法
  const hash = await writeContract(wagmiConfig, {
    address: marketAddress,
    abi: PredictionMarket.abi,
    functionName: 'buy',
    args: [outcome, amount],
  });
  // 等待交易确认并返回回执
  return waitForTransactionReceipt(wagmiConfig, { hash });
}

/**
 * 在已结算市场中领取奖金
 * @param {string} marketAddress - 市场合约地址
 * @returns {Promise<object>} 交易回执
 */
export async function claimMarket(marketAddress) {
  // 调用市场合约的 claim 方法
  const hash = await writeContract(wagmiConfig, {
    address: marketAddress,
    abi: PredictionMarket.abi,
    functionName: 'claim',
  });
  // 等待交易确认并返回回执
  return waitForTransactionReceipt(wagmiConfig, { hash });
}

/**
 * 在 V3 CPMM 二元市场中购买指定结果
 * @param {string} marketAddress - 市场合约地址
 * @param {number} outcome - 结果编号
 * @param {bigint} amountIn - 投入金额
 * @returns {Promise<object>} 交易回执
 */
export async function buyV3(marketAddress, outcome, amountIn) {
  // 调用 V3 市场合约的 buy 方法
  const hash = await writeContract(wagmiConfig, {
    address: marketAddress,
    abi: PredictionMarketV3.abi,
    functionName: 'buy',
    args: [outcome, amountIn],
  });
  // 等待交易确认并返回回执
  return waitForTransactionReceipt(wagmiConfig, { hash });
}

/**
 * 在多结果市场中购买指定结果
 * @param {string} marketAddress - 市场合约地址
 * @param {number} outcome - 结果编号
 * @param {bigint} amount - 购买金额
 * @returns {Promise<object>} 交易回执
 */
export async function buyMulti(marketAddress, outcome, amount) {
  // 调用多结果市场合约的 buy 方法
  const hash = await writeContract(wagmiConfig, {
    address: marketAddress,
    abi: MultiOutcomeMarket.abi,
    functionName: 'buy',
    args: [outcome, amount],
  });
  // 等待交易确认并返回回执
  return waitForTransactionReceipt(wagmiConfig, { hash });
}

/**
 * 向 V3 CPMM 市场添加流动性
 * @param {string} marketAddress - 市场合约地址
 * @param {bigint} amount - 注入的流动性金额
 * @returns {Promise<object>} 交易回执
 */
export async function addLiquidityV3(marketAddress, amount) {
  // 调用 V3 市场合约的 addLiquidity 方法
  const hash = await writeContract(wagmiConfig, {
    address: marketAddress,
    abi: PredictionMarketV3.abi,
    functionName: 'addLiquidity',
    args: [amount],
  });
  // 等待交易确认并返回回执
  return waitForTransactionReceipt(wagmiConfig, { hash });
}

/**
 * 读取 V3 市场的 CPMM 池状态（储备量和价格）
 * @param {string} marketAddress - 市场合约地址
 * @returns {Promise<object>} 包含 reserveYes、reserveNo、priceYesBps 的对象
 */
export async function readPoolStateV3(marketAddress) {
  // 调用合约获取池状态
  const [reserveYes, reserveNo, priceYesBps] = await readContract(wagmiConfig, {
    address: marketAddress,
    abi: PredictionMarketV3.abi,
    functionName: 'getPoolState',
  });
  // 返回解构后的池状态
  return { reserveYes, reserveNo, priceYesBps };
}

/**
 * 读取 V1 市场的链上状态（状态码、获胜结果、各池金额）
 * @param {string} marketAddress - 市场合约地址
 * @returns {Promise<object>} 包含 status、winningOutcome、yesPool、noPool 的对象
 */
export async function readMarketStatus(marketAddress) {
  // 并行调用多个合约方法以提高性能
  const [status, winningOutcome, yesPool, noPool] = await Promise.all([
    // 获取市场状态
    readContract(wagmiConfig, {
      address: marketAddress,
      abi: PredictionMarket.abi,
      functionName: 'status',
    }),
    // 获取获胜结果
    readContract(wagmiConfig, {
      address: marketAddress,
      abi: PredictionMarket.abi,
      functionName: 'winningOutcome',
    }),
    // 获取 Yes 池金额
    readContract(wagmiConfig, {
      address: marketAddress,
      abi: PredictionMarket.abi,
      functionName: 'yesPool',
    }),
    // 获取 No 池金额
    readContract(wagmiConfig, {
      address: marketAddress,
      abi: PredictionMarket.abi,
      functionName: 'noPool',
    }),
  ]);
  // 返回市场状态对象
  return { status, winningOutcome, yesPool, noPool };
}
