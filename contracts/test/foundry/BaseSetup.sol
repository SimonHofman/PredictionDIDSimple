// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "forge-std/Test.sol";
import "contracts/MockUSDC.sol";
import "contracts/PredictionMarket.sol";
import "contracts/MarketFactory.sol";
import "contracts/OracleAdapter.sol";
import "contracts/PredictionMarketV3.sol";
import "contracts/MarketFactoryV3.sol";
import "contracts/MultiOutcomeMarket.sol";
import "contracts/OracleAdapterV2.sol";
import "contracts/DIDRegistry.sol";

/// @title BaseSetup
/// @notice Foundry 测试基类：提供统一账户、常量与部署辅助函数
/// @dev 所有测试合约继承此类，在各自 setUp() 中调用 _deployPhase2Stack 或 _deployPhase3Stack
///      设计原则：每个测试用例通过独立 setUp 获得干净链上状态，保证隔离性与可重复性
abstract contract BaseSetup is Test {
    // --- 常量：与 MockUSDC 6 位小数及业务时间窗口对齐 ---
    uint256 internal constant USDC_UNIT = 1e6;   // 1 USDC（6 decimals）
    uint256 internal constant ONE_DAY = 1 days;
    uint256 internal constant ONE_WEEK = 7 days;

    // --- 合约实例：由子类 setUp 按需部署 ---
    MockUSDC internal usdc;
    OracleAdapter internal adapter;       // Phase2 时间锁预言机
    MarketFactory internal factory;       // Phase2 工厂
    OracleAdapterV2 internal adapterV2;   // Phase3 多签预言机
    MarketFactoryV3 internal factoryV3;   // Phase3 工厂（CPMM + 多结果）
    DIDRegistry internal didRegistry;

    // --- 测试账户：使用 makeAddr 生成确定性地址，便于调试 ---
    address internal owner;    // 工厂 Owner / Adapter Admin
    address internal oracle;   // 主预言机
    address internal oracle2;  // 第二预言机（多签场景）
    address internal alice;    // 普通用户 A
    address internal bob;      // 普通用户 B

    bytes32 internal constant MATCH_REF = keccak256("match-ref-001");
    string internal constant QUESTION = "Will Team A win?";

    /// @notice 初始化测试账户；子类必须调用 super.setUp()
    function setUp() public virtual {
        owner = makeAddr("owner");
        oracle = makeAddr("oracle");
        oracle2 = makeAddr("oracle2");
        alice = makeAddr("alice");
        bob = makeAddr("bob");
    }

    /// @notice 部署 Phase2 技术栈：MockUSDC + OracleAdapter + MarketFactory
    /// @param timelockDelay 结算时间锁秒数；0 表示可使用 resolveNow 快速路径
    function _deployPhase2Stack(uint256 timelockDelay) internal {
        vm.startPrank(owner);
        usdc = new MockUSDC();
        adapter = new OracleAdapter(owner, timelockDelay);
        adapter.grantOracle(oracle);
        factory = new MarketFactory(address(usdc), address(adapter));
        adapter.setFactory(address(factory));
        vm.stopPrank();
    }

    /// @notice 部署 Phase3 技术栈：MockUSDC + OracleAdapterV2 + MarketFactoryV3
    /// @param feeBps 默认手续费（基点，100 = 1%）
    /// @param threshold 多签阈值；>1 时自动授予 oracle2 角色
    function _deployPhase3Stack(uint16 feeBps, uint256 threshold) internal {
        vm.startPrank(owner);
        usdc = new MockUSDC();
        adapterV2 = new OracleAdapterV2(owner, threshold);
        adapterV2.grantOracle(oracle);
        if (threshold > 1) {
            adapterV2.grantOracle(oracle2);
        }
        factoryV3 = new MarketFactoryV3(address(usdc), address(adapterV2), feeBps);
        vm.stopPrank();
    }

    /// @notice 计算未来截止时间，避免硬编码 block.timestamp
    function _futureEndTime(uint256 offset) internal view returns (uint256) {
        return block.timestamp + offset;
    }

    /// @notice 通过 Phase2 工厂创建二元市场并返回类型化实例
    function _createBinaryMarket(uint256 endTime) internal returns (PredictionMarket market, address marketAddr) {
        vm.prank(owner);
        (marketAddr,) = factory.createMarket(MATCH_REF, QUESTION, endTime);
        market = PredictionMarket(marketAddr);
    }

    /// @notice 为用户铸造 USDC 并授权 spender（市场或工厂）拉取代币
    function _mintAndApprove(address user, address spender, uint256 amount) internal {
        usdc.mint(user, amount);
        vm.prank(user);
        usdc.approve(spender, amount);
    }

    /// @notice 仅铸造 USDC，不授权
    function _mintUsdc(address user, uint256 amount) internal {
        usdc.mint(user, amount);
    }

    /// @notice 为 DIDRegistry.bindDid 生成 EIP-191 个人签名
    /// @dev 消息格式与合约一致：keccak256("BindDID:" + account + didHash)
    function _signBindDid(address account, bytes32 didHash, uint256 privateKey)
        internal
        pure
        returns (bytes memory signature)
    {
        bytes32 hash = keccak256(abi.encodePacked("BindDID:", account, didHash));
        bytes32 digest = keccak256(abi.encodePacked("\x19Ethereum Signed Message:\n32", hash));
        (uint8 v, bytes32 r, bytes32 s) = vm.sign(privateKey, digest);
        signature = abi.encodePacked(r, s, v);
    }
}
