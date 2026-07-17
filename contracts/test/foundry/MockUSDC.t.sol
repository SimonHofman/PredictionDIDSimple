// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./BaseSetup.sol";

/// @title MockUSDCTest
/// @notice MockUSDC 抵押代币测试：验证 ERC20 基础属性与 mint 行为
contract MockUSDCTest is BaseSetup {
    MockUSDC internal token;

    function setUp() public override {
        super.setUp();
        token = new MockUSDC();
    }

    // =========================================================================
    // 部署测试（Deployment）：验证构造函数与初始状态
    // =========================================================================

    function test_Deploy_nameAndSymbol() public view {
        assertEq(token.name(), "Mock USDC");
        assertEq(token.symbol(), "mUSDC");
    }

    function test_Deploy_decimalsIsSix() public view {
        // 与真实 USDC 一致，便于金额换算（1 USDC = 1e6）
        assertEq(token.decimals(), 6);
    }

    function test_Deploy_initialSupplyZero() public view {
        assertEq(token.totalSupply(), 0);
    }

    // =========================================================================
    // 功能测试（Functional）：正常 mint / transfer 流程
    // =========================================================================

    function test_Mint_increasesBalance() public {
        uint256 amount = 100 * USDC_UNIT;
        token.mint(alice, amount);
        assertEq(token.balanceOf(alice), amount);
        assertEq(token.totalSupply(), amount);
    }

    function test_Transfer_movesBalance() public {
        uint256 amount = 50 * USDC_UNIT;
        token.mint(alice, amount);
        vm.prank(alice); // 模拟 alice 发起转账
        token.transfer(bob, amount);
        assertEq(token.balanceOf(alice), 0);
        assertEq(token.balanceOf(bob), amount);
    }

    // =========================================================================
    // 边界测试（Boundary）：零值与大额
    // =========================================================================

    function test_Mint_zeroAmount() public {
        token.mint(alice, 0);
        assertEq(token.balanceOf(alice), 0);
    }

    function test_Mint_largeAmount() public {
        uint256 large = type(uint128).max;
        token.mint(alice, large);
        assertEq(token.balanceOf(alice), large);
    }

    // =========================================================================
    // 错误测试（Error）：余额不足应 revert
    // =========================================================================

    function test_RevertWhen_transferInsufficientBalance() public {
        vm.prank(alice);
        vm.expectRevert();
        token.transfer(bob, 1);
    }

    // =========================================================================
    // Gas 测试：记录 mint 消耗并设合理上限
    // =========================================================================

    function test_Gas_mint() public {
        uint256 gasBefore = gasleft();
        token.mint(alice, 100 * USDC_UNIT);
        uint256 gasUsed = gasBefore - gasleft();
        assertTrue(gasUsed > 0 && gasUsed < 100_000);
    }
}
