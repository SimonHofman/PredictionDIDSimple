const fs = require("fs"); // 导入文件系统模块
const path = require("path"); // 导入路径处理模块
const hre = require("hardhat"); // 导入Hardhat运行环境

// 主部署函数 - 第二阶段部署脚本
async function main() {
  const [deployer] = await hre.ethers.getSigners(); // 获取部署者账户
  const timelock = Number(process.env.ORACLE_TIMELOCK_SECONDS || "120"); // 获取时间锁延迟，默认120秒
  console.log("Deployer:", deployer.address); // 打印部署者地址

  const USDC = await hre.ethers.getContractFactory("MockUSDC"); // 获取MockUSDC合约工厂
  const usdc = await USDC.deploy(); // 部署MockUSDC合约
  await usdc.waitForDeployment(); // 等待部署完成
  const usdcAddr = await usdc.getAddress(); // 获取USDC合约地址

  const Adapter = await hre.ethers.getContractFactory("OracleAdapter"); // 获取预言机适配器工厂
  const adapter = await Adapter.deploy(deployer.address, timelock); // 部署预言机适配器
  await adapter.waitForDeployment(); // 等待部署完成
  const adapterAddr = await adapter.getAddress(); // 获取适配器合约地址

  const Registry = await hre.ethers.getContractFactory("DIDRegistry"); // 获取DID注册表工厂
  const registry = await Registry.deploy(); // 部署DID注册表合约
  await registry.waitForDeployment(); // 等待部署完成
  const registryAddr = await registry.getAddress(); // 获取注册表合约地址

  const Factory = await hre.ethers.getContractFactory("MarketFactory"); // 获取市场工厂合约工厂
  const factory = await Factory.deploy(usdcAddr, adapterAddr); // 部署市场工厂合约
  await factory.waitForDeployment(); // 等待部署完成
  const factoryAddr = await factory.getAddress(); // 获取工厂合约地址

  await adapter.setFactory(factoryAddr); // 在适配器中设置工厂地址
  await adapter.grantOracle(deployer.address); // 授予部署者预言机角色

  const mintAmount = hre.ethers.parseUnits("100000", 6); // 铸造金额：100000 USDC
  await usdc.mint(deployer.address, mintAmount); // 为部署者铸造测试代币

  // 构建部署信息对象
  const out = {
    chainId: (await hre.ethers.provider.getNetwork()).chainId.toString(), // 链 ID
    mockUSDC: usdcAddr, // MockUSDC合约地址
    oracleAdapter: adapterAddr, // 预言机适配器地址
    didRegistry: registryAddr, // DID注册表地址
    marketFactory: factoryAddr, // 市场工厂地址
    oracle: deployer.address, // 预言机地址（部署者）
    timelockSeconds: timelock, // 时间锁延迟秒数
    deployedAt: new Date().toISOString(), // 部署时间
  };

  const outPath = path.join(__dirname, "..", "deployments.local.json"); // 部署信息输出路径
  fs.writeFileSync(outPath, JSON.stringify(out, null, 2)); // 写入部署信息到文件
  console.log("Wrote", outPath); // 打印输出路径
  console.log(JSON.stringify(out, null, 2)); // 打印部署信息
}

// 执行主函数，捕获并处理错误
main().catch((e) => {
  console.error(e); // 打印错误信息
  process.exitCode = 1; // 设置退出码为1表示失败
});
