const { expect } = require("chai"); // 导入chai断言库
const { ethers } = require("hardhat"); // 导入Hardhat的ethers库
const { time } = require("@nomicfoundation/hardhat-network-helpers"); // 导入网络时间帮助工具

// PredictionMarket合约测试套件
describe("PredictionMarket", function () {
  let collateral, factory, adapter, oracle, owner, alice, bob; // 声明测试变量
  const matchRef = ethers.id("wc-2026-001"); // 比赛引用哈希
  const question = "Will Team A win?"; // 预测问题
  let endTime; // 截止时间
  let market; // 市场合约实例
  let marketAddr; // 市场合约地址

  // 每个测试前的初始化设置
  beforeEach(async function () {
    [owner, oracle, alice, bob] = await ethers.getSigners(); // 获取测试账户
    endTime = (await time.latest()) + 86400 * 7; // 截止时间为7天后

    const USDC = await ethers.getContractFactory("MockUSDC"); // 获取MockUSDC工厂
    collateral = await USDC.deploy(); // 部署MockUSDC
    const Adapter = await ethers.getContractFactory("OracleAdapter"); // 获取适配器工厂
    adapter = await Adapter.deploy(owner.address, 0); // 部署适配器（时间锁为0）
    await adapter.grantOracle(oracle.address); // 授予预言机角色
    const Factory = await ethers.getContractFactory("MarketFactory"); // 获取工厂合约工厂
    factory = await Factory.deploy(await collateral.getAddress(), await adapter.getAddress()); // 部署市场工厂
    await factory.createMarket(matchRef, question, endTime); // 创建预测市场
    marketAddr = await factory.markets(1); // 获取市场地址
    market = await ethers.getContractAt("PredictionMarket", marketAddr); // 获取市场合约实例

    await collateral.mint(alice.address, ethers.parseUnits("1000", 6)); // 为Alice铸造1000 USDC
    await collateral.mint(bob.address, ethers.parseUnits("1000", 6)); // 为Bob铸造1000 USDC
    await collateral.connect(alice).approve(marketAddr, ethers.MaxUint256); // Alice授权市场合约
    await collateral.connect(bob).approve(marketAddr, ethers.MaxUint256); // Bob授权市场合约
  });

  // 测试：购买“是”和“否”
  it("buys yes and no", async function () {
    await market.connect(alice).buy(0, ethers.parseUnits("100", 6)); // Alice押注100 USDC到“是”方
    await market.connect(bob).buy(1, ethers.parseUnits("50", 6)); // Bob押注50 USDC到“否”方
    expect(await market.yesPool()).to.equal(ethers.parseUnits("100", 6)); // 验证“是”方资金池为100
    expect(await market.noPool()).to.equal(ethers.parseUnits("50", 6)); // 验证“否”方资金池为50
  });

  // 测试：结算后获胜者领取奖励
  it("resolves and winner claims", async function () {
    await market.connect(alice).buy(0, ethers.parseUnits("100", 6)); // Alice押注100 USDC到“是”方
    await market.connect(bob).buy(1, ethers.parseUnits("100", 6)); // Bob押注100 USDC到“否”方
    await adapter.connect(oracle).resolveNow(marketAddr, 0); // 结算市场，“是”方获胜

    const balBefore = await collateral.balanceOf(alice.address); // 记录Alice领取前余额
    await market.connect(alice).claim(); // Alice领取奖励
    const balAfter = await collateral.balanceOf(alice.address); // 记录Alice领取后余额
    expect(balAfter - balBefore).to.equal(ethers.parseUnits("200", 6)); // 验证获得200 USDC（总池）
  });

  // 测试：失败者无法领取
  it("loser cannot claim", async function () {
    await market.connect(alice).buy(0, ethers.parseUnits("100", 6)); // Alice押注“是”方
    await market.connect(bob).buy(1, ethers.parseUnits("100", 6)); // Bob押注“否”方
    await adapter.connect(oracle).resolveNow(marketAddr, 0); // “是”方获胜
    await expect(market.connect(bob).claim()).to.be.revertedWith("nothing to claim"); // 验证Bob无法领取
  });

  // 测试：拒绝重复领取
  it("rejects double claim", async function () {
    await market.connect(alice).buy(0, ethers.parseUnits("50", 6)); // Alice押注50 USDC
    await adapter.connect(oracle).resolveNow(marketAddr, 0); // 结算市场
    await market.connect(alice).claim(); // Alice第一次领取
    await expect(market.connect(alice).claim()).to.be.revertedWith("already claimed"); // 验证第二次领取被拒绝
  });

  // 测试：非预言机无法结算
  it("non-oracle cannot resolve", async function () {
    await expect(market.connect(alice).resolve(0)).to.be.revertedWith("not oracle"); // 验证普通用户无法结算
  });
});
