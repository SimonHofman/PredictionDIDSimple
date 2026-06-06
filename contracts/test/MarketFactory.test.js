const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time } = require("@nomicfoundation/hardhat-network-helpers");

describe("MarketFactory", function () {
  it("creates market and emits event", async function () {
    const [owner, oracle] = await ethers.getSigners();
    const USDC = await ethers.getContractFactory("MockUSDC");
    const collateral = await USDC.deploy();
    const Adapter = await ethers.getContractFactory("OracleAdapter");
    const adapter = await Adapter.deploy(owner.address, 0);
    await adapter.grantOracle(oracle.address);
    const Factory = await ethers.getContractFactory("MarketFactory");
    const factory = await Factory.deploy(await collateral.getAddress(), await adapter.getAddress());
    const endTime = (await time.latest()) + 86400;

    await expect(
      factory.createMarket(ethers.id("match-1"), "Home wins?", endTime)
    ).to.emit(factory, "MarketCreated");

    expect(await factory.marketCount()).to.equal(1);
    const marketAddr = await factory.markets(1);
    expect(marketAddr).to.properAddress;
    const market = await ethers.getContractAt("PredictionMarket", marketAddr);
    expect(await market.question()).to.equal("Home wins?");
  });
});
