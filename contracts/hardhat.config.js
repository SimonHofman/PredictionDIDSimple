// 导入Hardhat工具箱插件
require("@nomicfoundation/hardhat-toolbox");

// 导入代码覆盖率插件
require("solidity-coverage");

/** @type import('hardhat/config').HardhatUserConfig */
// Hardhat配置对象
module.exports = {
  solidity: { // Solidity编译器配置
    version: "0.8.24", // 编译器版本
    settings: {
      optimizer: { enabled: true, runs: 200 }, // 启用优化器，运行200次优化
      evmVersion: "cancun", // EVM版本为Cancun
    },
  },
  networks: { // 网络配置
    hardhat: { // Hardhat内置网络
      chainId: 31337, // 链 ID
    },
    localhost: { // 本地网络
      url: "http://127.0.0.1:8545", // RPC地址
      chainId: 31337, // 链 ID
    },
  },
  paths: { // 路径配置
    sources: "./contracts", // 合约源码目录
    tests: "./test", // 测试文件目录
    cache: "./cache", // 缓存目录
    artifacts: "./artifacts", // 编译产物目录
  },
};
