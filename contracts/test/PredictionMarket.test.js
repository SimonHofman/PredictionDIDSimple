const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time } = require("@nomicfoundation/hardhat-network-helpers");

describe("PredictionMarket", function () {
  let collateral, factory, adapter, oracle, owner, alice, bob;
  const matchRef = ethers.id("wc-2026-001");
  const question = "Will Team A win?";
  let endTime;
  let market;
  let marketAddr;

  beforeEach(async function () {
    [owner, oracle, alice, bob] = await ethers.getSigners();
    endTime = (await time.latest()) + 86400 * 7;

    const USDC = await ethers.getContractFactory("MockUSDC");
    collateral = await USDC.deploy();
    const Adapter = await ethers.getContractFactory("OracleAdapter");
    adapter = await Adapter.deploy(owner.address, 0);
    await adapter.grantOracle(oracle.address);
    const Factory = await ethers.getContractFactory("MarketFactory");
    factory = await Factory.deploy(await collateral.getAddress(), await adapter.getAddress());
    await factory.createMarket(matchRef, question, endTime);
    marketAddr = await factory.markets(1);
    market = await ethers.getContractAt("PredictionMarket", marketAddr);

    await collateral.mint(alice.address, ethers.parseUnits("1000", 6));
    await collateral.mint(bob.address, ethers.parseUnits("1000", 6));
    await collateral.connect(alice).approve(marketAddr, ethers.MaxUint256);
    await collateral.connect(bob).approve(marketAddr, ethers.MaxUint256);
  });

  it("buys yes and no", async function () {
    await market.connect(alice).buy(0, ethers.parseUnits("100", 6));
    await market.connect(bob).buy(1, ethers.parseUnits("50", 6));
    expect(await market.yesPool()).to.equal(ethers.parseUnits("100", 6));
    expect(await market.noPool()).to.equal(ethers.parseUnits("50", 6));
  });

  it("resolves and winner claims", async function () {
    await market.connect(alice).buy(0, ethers.parseUnits("100", 6));
    await market.connect(bob).buy(1, ethers.parseUnits("100", 6));
    await adapter.connect(oracle).resolveNow(marketAddr, 0);

    const balBefore = await collateral.balanceOf(alice.address);
    await market.connect(alice).claim();
    const balAfter = await collateral.balanceOf(alice.address);
    expect(balAfter - balBefore).to.equal(ethers.parseUnits("200", 6));
  });

  it("loser cannot claim", async function () {
    await market.connect(alice).buy(0, ethers.parseUnits("100", 6));
    await market.connect(bob).buy(1, ethers.parseUnits("100", 6));
    await adapter.connect(oracle).resolveNow(marketAddr, 0);
    await expect(market.connect(bob).claim()).to.be.revertedWith("nothing to claim");
  });

  it("rejects double claim", async function () {
    await market.connect(alice).buy(0, ethers.parseUnits("50", 6));
    await adapter.connect(oracle).resolveNow(marketAddr, 0);
    await market.connect(alice).claim();
    await expect(market.connect(alice).claim()).to.be.revertedWith("already claimed");
  });

  it("non-oracle cannot resolve", async function () {
    await expect(market.connect(alice).resolve(0)).to.be.revertedWith("not oracle");
  });
});
