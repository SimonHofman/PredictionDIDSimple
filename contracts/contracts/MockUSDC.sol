// SPDX-License-Identifier: MIT
// 许可证标识：MIT开源协议
pragma solidity ^0.8.24;
// 指定Solidity编译器版本为0.8.24及以上

import "@openzeppelin/contracts/token/ERC20/ERC20.sol"; // 导入ERC20代币标准合约

/// @title MockUSDC — test collateral (6 decimals like USDC)
/// @title MockUSDC — 测试用抵押品代币（6位小数，模拟USDC）
contract MockUSDC is ERC20 {
    // 构造函数，初始化代币名称为"Mock USDC"，符号为"mUSDC"
    constructor() ERC20("Mock USDC", "mUSDC") {}

    // 覆写精度函数，返回6位小数（与USDC一致）
    function decimals() public pure override returns (uint8) {
        return 6; // USDC使用6位小数
    }

    // 铸造代币函数，任何人都可以调用（仅用于测试）
    function mint(address to, uint256 amount) external {
        _mint(to, amount); // 调用内部铸造方法，向目标地址发行代币
    }
}
