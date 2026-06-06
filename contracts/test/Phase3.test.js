const { expect } = require("chai"); // 导入chai断言库
const { ethers } = require("hardhat"); // 导入Hardhat的ethers库
const { time } = require("@nomicfoundation/hardhat-network-helpers"); // 导入网络时间帮助工具

// 第三阶段功能测试套件
describe("Phase3", function () {
  // 测试：CPMM购买会移动价格
  it("CPMM buy shifts price", async function () {
    const [owner, oracle, alice] = await ethers.getSigners(); // 获取测试账户
    const USDC = await ethers.getContractFactory("MockUSDC"); // 获取MockUSDC工厂
    const usdc = await USDC.deploy(); // 部署MockUSDC
    const Adapter = await ethers.getContractFactory("OracleAdapterV2"); // 获取预言机适配器V2工厂
    const adapter = await Adapter.deploy(owner.address, 1); // 部署适配器（阈值1）
    await adapter.grantOracle(oracle.address); // 授予预言机角色
    const Factory = await ethers.getContractFactory("MarketFactoryV3"); // 获取市场工厂V3合约工厂
    const factory = await Factory.deploy(await usdc.getAddress(), await adapter.getAddress(), 100); // 部署V3工厂（手续费100bps）
    const liq = ethers.parseUnits("1000", 6); // 初始流动性为1000 USDC
    await usdc.mint(owner.address, liq * 2n); // 为所有者铸造双倍流动性的代币
    await usdc.connect(owner).approve(await factory.getAddress(), liq * 2n); // 授权工厂合约
    const end = (await time.latest()) + 86400; // 截止时间为24小时后
    await factory.connect(owner).createBinaryMarket(ethers.id("b1"), "Win?", end, liq); // 创建二元市场并注入流动性
    const marketAddr = await factory.markets(1); // 获取市场地址
    const market = await ethers.getContractAt("PredictionMarketV3", marketAddr); // 获取V3市场合约实例
    await usdc.mint(alice.address, ethers.parseUnits("500", 6)); // 为Alice铸造500 USDC
    await usdc.connect(alice).approve(marketAddr, ethers.MaxUint256); // Alice授权市场合约
    await market.connect(alice).buy(0, ethers.parseUnits("100", 6)); // Alice押注100 USDC到“是”方
    const [, , price] = await market.getPoolState(); // 获取池子状态中的价格
    expect(price).to.be.gt(5000n); // 验证“是”方价格大于50%（购买后价格上升）
  });

  // 测试：多结果市场结算获胜者
  it("multi outcome resolves winner", async function () {
    const [owner, oracle, alice] = await ethers.getSigners(); // 获取测试账户
    const USDC = await ethers.getContractFactory("MockUSDC"); // 获取MockUSDC工厂
    const usdc = await USDC.deploy(); // 部署MockUSDC
    const Adapter = await ethers.getContractFactory("OracleAdapterV2"); // 获取适配器V2工厂
    const adapter = await Adapter.deploy(owner.address, 1); // 部署适配器（阈值1）
    await adapter.grantOracle(oracle.address); // 授予预言机角色
    const Factory = await ethers.getContractFactory("MarketFactoryV3"); // 获取V3工厂
    const factory = await Factory.deploy(await usdc.getAddress(), await adapter.getAddress(), 50); // 部署V3工厂（手续费50bps）
    const end = (await time.latest()) + 86400; // 截止时间为24小时后
    await factory.connect(owner).createMultiMarket(ethers.id("g1"), "Group winner?", end, 3); // 创建3结果市场
    const marketAddr = await factory.markets(1); // 获取市场地址
    const market = await ethers.getContractAt("MultiOutcomeMarket", marketAddr); // 获取多结果市场实例
    await usdc.mint(alice.address, ethers.parseUnits("300", 6)); // 为Alice铸造300 USDC
    await usdc.connect(alice).approve(marketAddr, ethers.MaxUint256); // Alice授权市场合约
    await market.connect(alice).buy(0, ethers.parseUnits("100", 6)); // Alice押注100 USDC到结果0
    await market.connect(alice).buy(1, ethers.parseUnits("50", 6)); // Alice押注50 USDC到结果1
    await adapter.connect(oracle).proposeResolve(marketAddr, 0); // 预言机提议结果0获胜
    const balBefore = await usdc.balanceOf(alice.address); // 记录领取前余额
    await market.connect(alice).claim(); // Alice领取奖励
    expect(await usdc.balanceOf(alice.address)).to.be.gt(balBefore); // 验证余额增加
  });

  // 测试：多签需要达到阈值
  it("multisig needs threshold", async function () {
    const [o1, o2, o3, user] = await ethers.getSigners(); // 获取测试账户
    const USDC = await ethers.getContractFactory("MockUSDC"); // 获取MockUSDC工厂
    const usdc = await USDC.deploy(); // 部署MockUSDC
    const Adapter = await ethers.getContractFactory("OracleAdapterV2"); // 获取适配器V2工厂
    const adapter = await Adapter.deploy(o1.address, 2); // 部署适配器（阈值2，需要2个批准）
    await adapter.grantOracle(o1.address); // 授予o1预言机角色
    await adapter.grantOracle(o2.address); // 授予o2预言机角色
    const Factory = await ethers.getContractFactory("MarketFactory"); // 获取市场工厂合约工厂
    const factory = await Factory.deploy(await usdc.getAddress(), await adapter.getAddress()); // 部署市场工厂
    const end = (await time.latest()) + 86400; // 截止时间为24小时后
    await factory.createMarket(ethers.id("m"), "Q", end); // 创建市场
    const marketAddr = await factory.markets(1); // 获取市场地址
    const tx = await adapter.connect(o1).proposeResolve(marketAddr, 0); // o1提议结算（自动批准一次）
    await tx.wait(); // 等待交易确认
    const p = await adapter.proposals(1); // 获取提案信息
    expect(p.executed).to.equal(false); // 验证提案未执行（只有1个批准，未达阈值2）
    await adapter.connect(o2).approveResolve(1); // o2批准提案（达到阈值）
    const p2 = await adapter.proposals(1); // 再次获取提案信息
    expect(p2.executed).to.equal(true); // 验证提案已执行
  });
});
