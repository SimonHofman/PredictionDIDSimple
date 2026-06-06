import { parseUnits, formatUnits } from 'viem';
import { readContract, writeContract, waitForTransactionReceipt } from '@wagmi/core';
import { wagmiConfig } from '../wagmi';
import MockUSDC from '../abis/MockUSDC.json';
import PredictionMarket from '../abis/PredictionMarket.json';
import PredictionMarketV3 from '../abis/PredictionMarketV3.json';
import MultiOutcomeMarket from '../abis/MultiOutcomeMarket.json';

const USDC_DECIMALS = 6;

export function parseUsdc(amount) {
  return parseUnits(String(amount), USDC_DECIMALS);
}

export function formatUsdc(amount) {
  return formatUnits(BigInt(amount || 0), USDC_DECIMALS);
}

export async function readUsdcBalance(address, tokenAddress) {
  return readContract(wagmiConfig, {
    address: tokenAddress,
    abi: MockUSDC.abi,
    functionName: 'balanceOf',
    args: [address],
  });
}

export async function approveUsdc(tokenAddress, spender, amount) {
  const hash = await writeContract(wagmiConfig, {
    address: tokenAddress,
    abi: MockUSDC.abi,
    functionName: 'approve',
    args: [spender, amount],
  });
  return waitForTransactionReceipt(wagmiConfig, { hash });
}

export async function buyOutcome(marketAddress, outcome, amount) {
  const hash = await writeContract(wagmiConfig, {
    address: marketAddress,
    abi: PredictionMarket.abi,
    functionName: 'buy',
    args: [outcome, amount],
  });
  return waitForTransactionReceipt(wagmiConfig, { hash });
}

export async function claimMarket(marketAddress) {
  const hash = await writeContract(wagmiConfig, {
    address: marketAddress,
    abi: PredictionMarket.abi,
    functionName: 'claim',
  });
  return waitForTransactionReceipt(wagmiConfig, { hash });
}

export async function buyV3(marketAddress, outcome, amountIn) {
  const hash = await writeContract(wagmiConfig, {
    address: marketAddress,
    abi: PredictionMarketV3.abi,
    functionName: 'buy',
    args: [outcome, amountIn],
  });
  return waitForTransactionReceipt(wagmiConfig, { hash });
}

export async function buyMulti(marketAddress, outcome, amount) {
  const hash = await writeContract(wagmiConfig, {
    address: marketAddress,
    abi: MultiOutcomeMarket.abi,
    functionName: 'buy',
    args: [outcome, amount],
  });
  return waitForTransactionReceipt(wagmiConfig, { hash });
}

export async function addLiquidityV3(marketAddress, amount) {
  const hash = await writeContract(wagmiConfig, {
    address: marketAddress,
    abi: PredictionMarketV3.abi,
    functionName: 'addLiquidity',
    args: [amount],
  });
  return waitForTransactionReceipt(wagmiConfig, { hash });
}

export async function readPoolStateV3(marketAddress) {
  const [reserveYes, reserveNo, priceYesBps] = await readContract(wagmiConfig, {
    address: marketAddress,
    abi: PredictionMarketV3.abi,
    functionName: 'getPoolState',
  });
  return { reserveYes, reserveNo, priceYesBps };
}

export async function readMarketStatus(marketAddress) {
  const [status, winningOutcome, yesPool, noPool] = await Promise.all([
    readContract(wagmiConfig, {
      address: marketAddress,
      abi: PredictionMarket.abi,
      functionName: 'status',
    }),
    readContract(wagmiConfig, {
      address: marketAddress,
      abi: PredictionMarket.abi,
      functionName: 'winningOutcome',
    }),
    readContract(wagmiConfig, {
      address: marketAddress,
      abi: PredictionMarket.abi,
      functionName: 'yesPool',
    }),
    readContract(wagmiConfig, {
      address: marketAddress,
      abi: PredictionMarket.abi,
      functionName: 'noPool',
    }),
  ]);
  return { status, winningOutcome, yesPool, noPool };
}
