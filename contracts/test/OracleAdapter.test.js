const { expect } = require("chai"); // 导入chai断言库
const { ethers } = require("hardhat"); // 导入Hardhat的ethers库
const { time } = require("@nomicfoundation/hardhat-network-helpers"); // 导入网络时间帮助工具

// OracleAdapter合约测试套件
describe("OracleAdapter", function () {
  // 测试：时间锁结算流程
  it("timelocked resolve flow", async function () {
    const [admin, oracle] = await ethers.getSigners(); // 获取测试账户
    const USDC = await ethers.getContractFactory("MockUSDC"); // 获取MockUSDC工厂
    const usdc = await USDC.deploy(); // 部署MockUSDC
    const Adapter = await ethers.getContractFactory("OracleAdapter"); // 获取预言机适配器工厂
    const adapter = await Adapter.deploy(admin.address, 120); // 部署适配器（时间锁120秒）
    await adapter.grantOracle(oracle.address); // 授予预言机角色
    const Factory = await ethers.getContractFactory("MarketFactory"); // 获取市场工厂合约工厂
    const factory = await Factory.deploy(await usdc.getAddress(), await adapter.getAddress()); // 部署市场工厂
    const endTime = (await time.latest()) + 86400; // 截止时间为24小时后
    await factory.createMarket(ethers.id("m1"), "Home wins?", endTime); // 创建市场
    const marketAddr = await factory.markets(1); // 获取市场地址
    const market = await ethers.getContractAt("PredictionMarket", marketAddr); // 获取市场合约实例

    await adapter.connect(oracle).requestResolve(marketAddr, 0); // 请求结算（启动时间锁）
    await expect(adapter.connect(oracle).confirmResolve(marketAddr)).to.be.revertedWith("timelock"); // 确认时间锁未过时会失败
    await time.increase(121); // 快进时间121秒
    await adapter.connect(oracle).confirmResolve(marketAddr); // 时间锁过后确认结算
    expect(await market.status()).to.equal(1); // 验证市场状态为已结算(1)
  });

  // 测试：作废市场退回押注
  it("void market refunds stake", async function () {
    const [admin, oracle, alice] = await ethers.getSigners(); // 获取测试账户
    const USDC = await ethers.getContractFactory("MockUSDC"); // 获取MockUSDC工厂
    const usdc = await USDC.deploy(); // 部署MockUSDC
    const Adapter = await ethers.getContractFactory("OracleAdapter"); // 获取适配器工厂
    const adapter = await Adapter.deploy(admin.address, 0); // 部署适配器（时间锁为0）
    await adapter.grantOracle(oracle.address); // 授予预言机角色
    const Factory = await ethers.getContractFactory("MarketFactory"); // 获取市场工厂合约工厂
    const factory = await Factory.deploy(await usdc.getAddress(), await adapter.getAddress()); // 部署市场工厂
    const endTime = (await time.latest()) + 86400; // 截止时间为24小时后
    await factory.createMarket(ethers.id("m2"), "Q?", endTime); // 创建市场
    const marketAddr = await factory.markets(1); // 获取市场地址
    const market = await ethers.getContractAt("PredictionMarket", marketAddr); // 获取市场合约实例
    await usdc.mint(alice.address, ethers.parseUnits("100", 6)); // 为Alice铸造100 USDC
    await usdc.connect(alice).approve(marketAddr, ethers.MaxUint256); // Alice授权市场合约无限额度
    await market.connect(alice).buy(0, ethers.parseUnits("40", 6)); // Alice押注40 USDC到“是”方
    await adapter.connect(oracle).voidMarket(marketAddr); // 预言机作废市场
    const before = await usdc.balanceOf(alice.address); // 记录Alice领取前余额
    await market.connect(alice).claim(); // Alice领取奖励
    const after = await usdc.balanceOf(alice.address); // 记录Alice领取后余额
    expect(after - before).to.equal(ethers.parseUnits("40", 6)); // 验证退回了40 USDC
  });
});
