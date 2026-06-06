const fs = require("fs"); // 导入文件系统模块
const path = require("path"); // 导入路径处理模块
const hre = require("hardhat"); // 导入Hardhat运行环境

// 主部署函数 - 第三阶段部署脚本
async function main() {
  const [deployer] = await hre.ethers.getSigners(); // 获取部署者账户
  const threshold = Number(process.env.ORACLE_MULTISIG_THRESHOLD || "2"); // 获取多签阈值，默认2
  const feeBps = Number(process.env.DEFAULT_FEE_BPS || "30"); // 获取默认手续费，默认30基点

  const USDC = await hre.ethers.getContractFactory("MockUSDC"); // 获取MockUSDC合约工厂
  const usdc = await USDC.deploy(); // 部署MockUSDC合约
  await usdc.waitForDeployment(); // 等待部署完成

  const AdapterV2 = await hre.ethers.getContractFactory("OracleAdapterV2"); // 获取预言机适配器V2工厂
  const adapter = await AdapterV2.deploy(deployer.address, threshold); // 部署预言机适配器V2
  await adapter.waitForDeployment(); // 等待部署完成

  const FactoryV3 = await hre.ethers.getContractFactory("MarketFactoryV3"); // 获取市场工厂V3合约工厂
  const factory = await FactoryV3.deploy(await usdc.getAddress(), await adapter.getAddress(), feeBps); // 部署市场工厂V3
  await factory.waitForDeployment(); // 等待部署完成

  await adapter.grantOracle(deployer.address); // 授予部署者预言机角色
  await usdc.mint(deployer.address, hre.ethers.parseUnits("500000", 6)); // 为部署者铸造500000 USDC

  // 构建部署信息对象
  const out = {
    chainId: (await hre.ethers.provider.getNetwork()).chainId.toString(), // 链 ID
    mockUSDC: await usdc.getAddress(), // MockUSDC合约地址
    oracleAdapterV2: await adapter.getAddress(), // 预言机适配器V2地址
    marketFactoryV3: await factory.getAddress(), // 市场工厂V3地址
    multisigThreshold: threshold, // 多签阈值
    defaultFeeBps: feeBps, // 默认手续费
    deployedAt: new Date().toISOString(), // 部署时间
  };
  const p = path.join(__dirname, "..", "deployments.phase3.json"); // 部署信息输出路径
  fs.writeFileSync(p, JSON.stringify(out, null, 2)); // 写入部署信息到文件
  console.log(JSON.stringify(out, null, 2)); // 打印部署信息
}

// 执行主函数，捕获并处理错误
main().catch((e) => {
  console.error(e); // 打印错误信息
  process.exitCode = 1; // 设置退出码为1表示失败
});
