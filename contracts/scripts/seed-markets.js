const fs = require("fs"); // 导入文件系统模块
const path = require("path"); // 导入路径处理模块
const hre = require("hardhat"); // 导入Hardhat运行环境

// 主函数 - 创建种子预测市场的脚本
async function main() {
  const deploymentsPath = path.join(__dirname, "..", "deployments.local.json"); // 部署信息文件路径
  if (!fs.existsSync(deploymentsPath)) { // 如果部署信息文件不存在
    throw new Error("Run deploy.js first"); // 抛出错误，提示先执行部署
  }
  const dep = JSON.parse(fs.readFileSync(deploymentsPath, "utf8")); // 读取部署信息
  const factory = await hre.ethers.getContractAt("MarketFactory", dep.marketFactory); // 获取工厂合约实例
  const now = await hre.ethers.provider.getBlock("latest"); // 获取最新区块信息
  const week = 7 * 24 * 60 * 60; // 一周的秒数

  // 种子市场配置数据
  const seeds = [
    { ref: "wc-2026-semi-001", q: "Argentina wins in 90 min?", days: 14 }, // 阿根廷90分钟内获胜？
    { ref: "wc-2026-semi-002", q: "France wins in 90 min?", days: 14 }, // 法国90分钟内获胜？
    { ref: "wc-2026-final-001", q: "Over 2.5 goals in final?", days: 21 }, // 决赛进球超过2.5个？
  ];

  for (const s of seeds) { // 遍历种子市场配置
    const endTime = now.timestamp + s.days * 24 * 60 * 60; // 计算截止时间
    const tx = await factory.createMarket(hre.ethers.id(s.ref), s.q, endTime); // 创建市场
    await tx.wait(); // 等待交易确认
    console.log("Created market:", s.ref, s.q); // 打印创建成功信息
  }
  console.log("marketCount:", (await factory.marketCount()).toString()); // 打印市场总数
}

// 执行主函数，捕获并处理错误
main().catch((e) => {
  console.error(e); // 打印错误信息
  process.exitCode = 1; // 设置退出码为1表示失败
});
