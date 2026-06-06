const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time } = require("@nomicfoundation/hardhat-network-helpers");

describe("OracleAdapter", function () {
  it("timelocked resolve flow", async function () {
    const [admin, oracle] = await ethers.getSigners();
    const USDC = await ethers.getContractFactory("MockUSDC");
    const usdc = await USDC.deploy();
    const Adapter = await ethers.getContractFactory("OracleAdapter");
    const adapter = await Adapter.deploy(admin.address, 120);
    await adapter.grantOracle(oracle.address);
    const Factory = await ethers.getContractFactory("MarketFactory");
    const factory = await Factory.deploy(await usdc.getAddress(), await adapter.getAddress());
    const endTime = (await time.latest()) + 86400;
    await factory.createMarket(ethers.id("m1"), "Home wins?", endTime);
    const marketAddr = await factory.markets(1);
    const market = await ethers.getContractAt("PredictionMarket", marketAddr);

    await adapter.connect(oracle).requestResolve(marketAddr, 0);
    await expect(adapter.connect(oracle).confirmResolve(marketAddr)).to.be.revertedWith("timelock");
    await time.increase(121);
    await adapter.connect(oracle).confirmResolve(marketAddr);
    expect(await market.status()).to.equal(1);
  });

  it("void market refunds stake", async function () {
    const [admin, oracle, alice] = await ethers.getSigners();
    const USDC = await ethers.getContractFactory("MockUSDC");
    const usdc = await USDC.deploy();
    const Adapter = await ethers.getContractFactory("OracleAdapter");
    const adapter = await Adapter.deploy(admin.address, 0);
    await adapter.grantOracle(oracle.address);
    const Factory = await ethers.getContractFactory("MarketFactory");
    const factory = await Factory.deploy(await usdc.getAddress(), await adapter.getAddress());
    const endTime = (await time.latest()) + 86400;
    await factory.createMarket(ethers.id("m2"), "Q?", endTime);
    const marketAddr = await factory.markets(1);
    const market = await ethers.getContractAt("PredictionMarket", marketAddr);
    await usdc.mint(alice.address, ethers.parseUnits("100", 6));
    await usdc.connect(alice).approve(marketAddr, ethers.MaxUint256);
    await market.connect(alice).buy(0, ethers.parseUnits("40", 6));
    await adapter.connect(oracle).voidMarket(marketAddr);
    const before = await usdc.balanceOf(alice.address);
    await market.connect(alice).claim();
    const after = await usdc.balanceOf(alice.address);
    expect(after - before).to.equal(ethers.parseUnits("40", 6));
  });
});
