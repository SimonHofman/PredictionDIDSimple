const { expect } = require("chai"); // 导入chai断言库
const { ethers } = require("hardhat"); // 导入Hardhat的ethers库
const { time } = require("@nomicfoundation/hardhat-network-helpers"); // 导入网络时间帮助工具

// MarketFactory合约测试套件
describe("MarketFactory", function () {
  // 测试：创建市场并触发事件
  it("creates market and emits event", async function () {
    const [owner, oracle] = await ethers.getSigners(); // 获取测试账户
    const USDC = await ethers.getContractFactory("MockUSDC"); // 获取MockUSDC工厂
    const collateral = await USDC.deploy(); // 部署MockUSDC
    const Adapter = await ethers.getContractFactory("OracleAdapter"); // 获取预言机适配器工厂
    const adapter = await Adapter.deploy(owner.address, 0); // 部署适配器（时间锁为0）
    await adapter.grantOracle(oracle.address); // 授予预言机角色
    const Factory = await ethers.getContractFactory("MarketFactory"); // 获取市场工厂合约工厂
    const factory = await Factory.deploy(await collateral.getAddress(), await adapter.getAddress()); // 部署市场工厂
    const endTime = (await time.latest()) + 86400; // 设置截止时间为24小时后

    // 验证创建市场时会触发MarketCreated事件
    await expect(
      factory.createMarket(ethers.id("match-1"), "Home wins?", endTime)
    ).to.emit(factory, "MarketCreated");

    expect(await factory.marketCount()).to.equal(1); // 验证市场计数为1
    const marketAddr = await factory.markets(1); // 获取市场地址
    expect(marketAddr).to.properAddress; // 验证是有效地址
    const market = await ethers.getContractAt("PredictionMarket", marketAddr); // 获取市场合约实例
    expect(await market.question()).to.equal("Home wins?"); // 验证市场问题正确
  });
});
