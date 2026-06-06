const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time } = require("@nomicfoundation/hardhat-network-helpers");

describe("Phase3", function () {
  it("CPMM buy shifts price", async function () {
    const [owner, oracle, alice] = await ethers.getSigners();
    const USDC = await ethers.getContractFactory("MockUSDC");
    const usdc = await USDC.deploy();
    const Adapter = await ethers.getContractFactory("OracleAdapterV2");
    const adapter = await Adapter.deploy(owner.address, 1);
    await adapter.grantOracle(oracle.address);
    const Factory = await ethers.getContractFactory("MarketFactoryV3");
    const factory = await Factory.deploy(await usdc.getAddress(), await adapter.getAddress(), 100);
    const liq = ethers.parseUnits("1000", 6);
    await usdc.mint(owner.address, liq * 2n);
    await usdc.connect(owner).approve(await factory.getAddress(), liq * 2n);
    const end = (await time.latest()) + 86400;
    await factory.connect(owner).createBinaryMarket(ethers.id("b1"), "Win?", end, liq);
    const marketAddr = await factory.markets(1);
    const market = await ethers.getContractAt("PredictionMarketV3", marketAddr);
    await usdc.mint(alice.address, ethers.parseUnits("500", 6));
    await usdc.connect(alice).approve(marketAddr, ethers.MaxUint256);
    await market.connect(alice).buy(0, ethers.parseUnits("100", 6));
    const [, , price] = await market.getPoolState();
    expect(price).to.be.gt(5000n);
  });

  it("multi outcome resolves winner", async function () {
    const [owner, oracle, alice] = await ethers.getSigners();
    const USDC = await ethers.getContractFactory("MockUSDC");
    const usdc = await USDC.deploy();
    const Adapter = await ethers.getContractFactory("OracleAdapterV2");
    const adapter = await Adapter.deploy(owner.address, 1);
    await adapter.grantOracle(oracle.address);
    const Factory = await ethers.getContractFactory("MarketFactoryV3");
    const factory = await Factory.deploy(await usdc.getAddress(), await adapter.getAddress(), 50);
    const end = (await time.latest()) + 86400;
    await factory.connect(owner).createMultiMarket(ethers.id("g1"), "Group winner?", end, 3);
    const marketAddr = await factory.markets(1);
    const market = await ethers.getContractAt("MultiOutcomeMarket", marketAddr);
    await usdc.mint(alice.address, ethers.parseUnits("300", 6));
    await usdc.connect(alice).approve(marketAddr, ethers.MaxUint256);
    await market.connect(alice).buy(0, ethers.parseUnits("100", 6));
    await market.connect(alice).buy(1, ethers.parseUnits("50", 6));
    await adapter.connect(oracle).proposeResolve(marketAddr, 0);
    const balBefore = await usdc.balanceOf(alice.address);
    await market.connect(alice).claim();
    expect(await usdc.balanceOf(alice.address)).to.be.gt(balBefore);
  });

  it("multisig needs threshold", async function () {
    const [o1, o2, o3, user] = await ethers.getSigners();
    const USDC = await ethers.getContractFactory("MockUSDC");
    const usdc = await USDC.deploy();
    const Adapter = await ethers.getContractFactory("OracleAdapterV2");
    const adapter = await Adapter.deploy(o1.address, 2);
    await adapter.grantOracle(o1.address);
    await adapter.grantOracle(o2.address);
    const Factory = await ethers.getContractFactory("MarketFactory");
    const factory = await Factory.deploy(await usdc.getAddress(), await adapter.getAddress());
    const end = (await time.latest()) + 86400;
    await factory.createMarket(ethers.id("m"), "Q", end);
    const marketAddr = await factory.markets(1);
    const tx = await adapter.connect(o1).proposeResolve(marketAddr, 0);
    await tx.wait();
    const p = await adapter.proposals(1);
    expect(p.executed).to.equal(false);
    await adapter.connect(o2).approveResolve(1);
    const p2 = await adapter.proposals(1);
    expect(p2.executed).to.equal(true);
  });
});
