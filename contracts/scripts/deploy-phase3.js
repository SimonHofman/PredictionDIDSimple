const fs = require("fs");
const path = require("path");
const hre = require("hardhat");

async function main() {
  const [deployer] = await hre.ethers.getSigners();
  const threshold = Number(process.env.ORACLE_MULTISIG_THRESHOLD || "2");
  const feeBps = Number(process.env.DEFAULT_FEE_BPS || "30");

  const USDC = await hre.ethers.getContractFactory("MockUSDC");
  const usdc = await USDC.deploy();
  await usdc.waitForDeployment();

  const AdapterV2 = await hre.ethers.getContractFactory("OracleAdapterV2");
  const adapter = await AdapterV2.deploy(deployer.address, threshold);
  await adapter.waitForDeployment();

  const FactoryV3 = await hre.ethers.getContractFactory("MarketFactoryV3");
  const factory = await FactoryV3.deploy(await usdc.getAddress(), await adapter.getAddress(), feeBps);
  await factory.waitForDeployment();

  await adapter.grantOracle(deployer.address);
  await usdc.mint(deployer.address, hre.ethers.parseUnits("500000", 6));

  const out = {
    chainId: (await hre.ethers.provider.getNetwork()).chainId.toString(),
    mockUSDC: await usdc.getAddress(),
    oracleAdapterV2: await adapter.getAddress(),
    marketFactoryV3: await factory.getAddress(),
    multisigThreshold: threshold,
    defaultFeeBps: feeBps,
    deployedAt: new Date().toISOString(),
  };
  const p = path.join(__dirname, "..", "deployments.phase3.json");
  fs.writeFileSync(p, JSON.stringify(out, null, 2));
  console.log(JSON.stringify(out, null, 2));
}

main().catch((e) => {
  console.error(e);
  process.exitCode = 1;
});
