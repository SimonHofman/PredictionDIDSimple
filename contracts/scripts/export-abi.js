const fs = require("fs"); // 导入文件系统模块
const path = require("path"); // 导入路径处理模块

// 需要导出的合约名称列表
const CONTRACTS = [
  "MarketFactory", // 市场工厂
  "PredictionMarket", // 预测市场
  "MockUSDC", // 测试用USDC
  "OracleAdapter", // 预言机适配器
  "DIDRegistry", // DID注册表
  "MarketFactoryV3", // 市场工厂V3
  "PredictionMarketV3", // 预测市场V3
  "MultiOutcomeMarket", // 多结果市场
  "OracleAdapterV2", // 预言机适配器V2
];
const ARTIFACTS_DIR = path.join(__dirname, "..", "artifacts", "contracts"); // 编译产物目录
const BACKEND_DIR = path.join(__dirname, "..", "..", "backend", "pkg", "contracts"); // 后端合约目录
const FRONTEND_DIR = path.join(__dirname, "..", "..", "frontend", "src", "abis"); // 前端ABI目录

// 加载合约编译产物
function loadArtifact(name) {
  const file = path.join(ARTIFACTS_DIR, `${name}.sol`, `${name}.json`); // 构建产物文件路径
  if (!fs.existsSync(file)) { // 如果文件不存在
    throw new Error( // 抛出错误
      `Artifact not found: ${file}\nRun: npm run compile`
    );
  }
  return JSON.parse(fs.readFileSync(file, "utf8")); // 读取并解析JSON文件
}

// 确保目录存在，不存在则创建
function ensureDir(dir) {
  if (!fs.existsSync(dir)) { // 如果目录不存在
    fs.mkdirSync(dir, { recursive: true }); // 递归创建目录
  }
}

// 导出单个合约的ABI到前后端
function exportOne(name) {
  const artifact = loadArtifact(name); // 加载合约产物
  // 后端载荷（包含ABI和字节码）
  const backendPayload = {
    contractName: name, // 合约名称
    abi: artifact.abi, // 合约ABI
    bytecode: artifact.bytecode, // 合约字节码
  };
  // 前端载荷（仅包含ABI）
  const frontendPayload = {
    contractName: name, // 合约名称
    abi: artifact.abi, // 合约ABI
  };

  ensureDir(BACKEND_DIR); // 确保后端目录存在
  ensureDir(FRONTEND_DIR); // 确保前端目录存在

  // 写入后端JSON文件
  fs.writeFileSync(
    path.join(BACKEND_DIR, `${name}.json`),
    JSON.stringify(backendPayload, null, 2)
  );
  // 写入前端JSON文件
  fs.writeFileSync(
    path.join(FRONTEND_DIR, `${name}.json`),
    JSON.stringify(frontendPayload, null, 2)
  );
  console.log(`Exported ${name} -> backend & frontend`); // 打印导出成功信息
}

// 主函数：遍历所有合约并导出
function main() {
  for (const name of CONTRACTS) { // 遍历合约列表
    exportOne(name); // 导出每个合约
  }
}

main(); // 执行主函数
