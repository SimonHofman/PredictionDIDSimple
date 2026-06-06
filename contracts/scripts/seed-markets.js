const fs = require("fs");
const path = require("path");
const hre = require("hardhat");

async function main() {
  const deploymentsPath = path.join(__dirname, "..", "deployments.local.json");
  if (!fs.existsSync(deploymentsPath)) {
    throw new Error("Run deploy.js first");
  }
  const dep = JSON.parse(fs.readFileSync(deploymentsPath, "utf8"));
  const factory = await hre.ethers.getContractAt("MarketFactory", dep.marketFactory);
  const now = await hre.ethers.provider.getBlock("latest");
  const week = 7 * 24 * 60 * 60;

  const seeds = [
    { ref: "wc-2026-semi-001", q: "Argentina wins in 90 min?", days: 14 },
    { ref: "wc-2026-semi-002", q: "France wins in 90 min?", days: 14 },
    { ref: "wc-2026-final-001", q: "Over 2.5 goals in final?", days: 21 },
  ];

  for (const s of seeds) {
    const endTime = now.timestamp + s.days * 24 * 60 * 60;
    const tx = await factory.createMarket(hre.ethers.id(s.ref), s.q, endTime);
    await tx.wait();
    console.log("Created market:", s.ref, s.q);
  }
  console.log("marketCount:", (await factory.marketCount()).toString());
}

main().catch((e) => {
  console.error(e);
  process.exitCode = 1;
});
