const fs = require("fs");
const path = require("path");

const CONTRACTS = [
  "MarketFactory",
  "PredictionMarket",
  "MockUSDC",
  "OracleAdapter",
  "DIDRegistry",
  "MarketFactoryV3",
  "PredictionMarketV3",
  "MultiOutcomeMarket",
  "OracleAdapterV2",
];
const ARTIFACTS_DIR = path.join(__dirname, "..", "artifacts", "contracts");
const BACKEND_DIR = path.join(__dirname, "..", "..", "backend", "pkg", "contracts");
const FRONTEND_DIR = path.join(__dirname, "..", "..", "frontend", "src", "abis");

function loadArtifact(name) {
  const file = path.join(ARTIFACTS_DIR, `${name}.sol`, `${name}.json`);
  if (!fs.existsSync(file)) {
    throw new Error(
      `Artifact not found: ${file}\nRun: npm run compile`
    );
  }
  return JSON.parse(fs.readFileSync(file, "utf8"));
}

function ensureDir(dir) {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
}

function exportOne(name) {
  const artifact = loadArtifact(name);
  const backendPayload = {
    contractName: name,
    abi: artifact.abi,
    bytecode: artifact.bytecode,
  };
  const frontendPayload = {
    contractName: name,
    abi: artifact.abi,
  };

  ensureDir(BACKEND_DIR);
  ensureDir(FRONTEND_DIR);

  fs.writeFileSync(
    path.join(BACKEND_DIR, `${name}.json`),
    JSON.stringify(backendPayload, null, 2)
  );
  fs.writeFileSync(
    path.join(FRONTEND_DIR, `${name}.json`),
    JSON.stringify(frontendPayload, null, 2)
  );
  console.log(`Exported ${name} -> backend & frontend`);
}

function main() {
  for (const name of CONTRACTS) {
    exportOne(name);
  }
}

main();
