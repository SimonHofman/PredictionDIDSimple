const fs = require("fs");
const path = require("path");
const hre = require("hardhat");

async function main() {
  const [deployer] = await hre.ethers.getSigners();
  const timelock = Number(process.env.ORACLE_TIMELOCK_SECONDS || "120");
  console.log("Deployer:", deployer.address);

  const USDC = await hre.ethers.getContractFactory("MockUSDC");
  const usdc = await USDC.deploy();
  await usdc.waitForDeployment();
  const usdcAddr = await usdc.getAddress();

  const Adapter = await hre.ethers.getContractFactory("OracleAdapter");
  const adapter = await Adapter.deploy(deployer.address, timelock);
  await adapter.waitForDeployment();
  const adapterAddr = await adapter.getAddress();

  const Registry = await hre.ethers.getContractFactory("DIDRegistry");
  const registry = await Registry.deploy();
  await registry.waitForDeployment();
  const registryAddr = await registry.getAddress();

  const Factory = await hre.ethers.getContractFactory("MarketFactory");
  const factory = await Factory.deploy(usdcAddr, adapterAddr);
  await factory.waitForDeployment();
  const factoryAddr = await factory.getAddress();

  await adapter.setFactory(factoryAddr);
  await adapter.grantOracle(deployer.address);

  const mintAmount = hre.ethers.parseUnits("100000", 6);
  await usdc.mint(deployer.address, mintAmount);

  const out = {
    chainId: (await hre.ethers.provider.getNetwork()).chainId.toString(),
    mockUSDC: usdcAddr,
    oracleAdapter: adapterAddr,
    didRegistry: registryAddr,
    marketFactory: factoryAddr,
    oracle: deployer.address,
    timelockSeconds: timelock,
    deployedAt: new Date().toISOString(),
  };

  const outPath = path.join(__dirname, "..", "deployments.local.json");
  fs.writeFileSync(outPath, JSON.stringify(out, null, 2));
  console.log("Wrote", outPath);
  console.log(JSON.stringify(out, null, 2));
}

main().catch((e) => {
  console.error(e);
  process.exitCode = 1;
});
