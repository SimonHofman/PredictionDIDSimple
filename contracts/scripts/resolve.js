const hre = require("hardhat");

async function main() {
  const marketAddress = process.env.MARKET_ADDRESS;
  const outcome = process.env.WINNING_OUTCOME;
  if (!marketAddress || outcome === undefined) {
    throw new Error("Set MARKET_ADDRESS and WINNING_OUTCOME (0=yes, 1=no)");
  }
  const [oracle] = await hre.ethers.getSigners();
  const market = await hre.ethers.getContractAt("PredictionMarket", marketAddress);
  const tx = await market.connect(oracle).resolve(Number(outcome));
  await tx.wait();
  console.log("Resolved", marketAddress, "outcome", outcome);
}

main().catch((e) => {
  console.error(e);
  process.exitCode = 1;
});
