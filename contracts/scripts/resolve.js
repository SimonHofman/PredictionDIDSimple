const hre = require("hardhat"); // 导入Hardhat运行环境

// 主函数 - 手动结算预测市场的脚本
async function main() {
  const marketAddress = process.env.MARKET_ADDRESS; // 从环境变量获取市场地址
  const outcome = process.env.WINNING_OUTCOME; // 从环境变量获取获胜结果
  if (!marketAddress || outcome === undefined) { // 如果缺少必要参数
    throw new Error("Set MARKET_ADDRESS and WINNING_OUTCOME (0=yes, 1=no)"); // 抛出错误提示
  }
  const [oracle] = await hre.ethers.getSigners(); // 获取预言机账户（第一个签名者）
  const market = await hre.ethers.getContractAt("PredictionMarket", marketAddress); // 获取市场合约实例
  const tx = await market.connect(oracle).resolve(Number(outcome)); // 调用结算函数
  await tx.wait(); // 等待交易确认
  console.log("Resolved", marketAddress, "outcome", outcome); // 打印结算结果
}

// 执行主函数，捕获并处理错误
main().catch((e) => {
  console.error(e); // 打印错误信息
  process.exitCode = 1; // 设置退出码为1表示失败
});
