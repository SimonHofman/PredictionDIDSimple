// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./BaseSetup.sol";

/// @title MarketFactoryV3Test
/// @notice Phase3 工厂测试：CPMM 二元市场、多结果市场、暂停与配置
contract MarketFactoryV3Test is BaseSetup {
    uint256 internal endTime;

    function setUp() public override {
        super.setUp();
        _deployPhase3Stack(100, 1);
        endTime = _futureEndTime(ONE_DAY);
    }

    // =========================================================================
    // 部署测试
    // =========================================================================

    function test_Deploy_defaults() public view {
        assertEq(address(factoryV3.collateral()), address(usdc));
        assertEq(factoryV3.oracle(), address(adapterV2));
        assertEq(factoryV3.defaultFeeBps(), 100);
        assertEq(factoryV3.defaultMaxBet(), 10_000 * USDC_UNIT);
        assertEq(factoryV3.marketCount(), 0);
    }

    function test_Deploy_version() public view {
        assertEq(factoryV3.version(), "3.0.0-phase3");
    }

    // =========================================================================
    // 功能测试
    // =========================================================================

    function test_CreateBinaryMarket_withLiquidity() public {
        uint256 liq = 500 * USDC_UNIT;
        _mintAndApprove(owner, address(factoryV3), liq * 2);
        vm.prank(owner);
        (address marketAddr, uint256 id) = factoryV3.createBinaryMarket(MATCH_REF, QUESTION, endTime, liq);
        assertEq(id, 1);
        assertEq(factoryV3.markets(1), marketAddr);
        assertEq(factoryV3.marketTypes(1), 0); // 0 = binary v3
        PredictionMarketV3 market = PredictionMarketV3(marketAddr);
        assertEq(market.reserveYes(), liq);
        assertEq(market.reserveNo(), liq);
    }

    function test_CreateBinaryMarket_withoutLiquidity() public {
        vm.prank(owner);
        (address marketAddr,) = factoryV3.createBinaryMarket(MATCH_REF, QUESTION, endTime, 0);
        PredictionMarketV3 market = PredictionMarketV3(marketAddr);
        assertEq(market.totalLPSupply(), 0);
    }

    function test_CreateMultiMarket() public {
        vm.prank(owner);
        (address marketAddr, uint256 id) =
            factoryV3.createMultiMarket(keccak256("g1"), "Winner?", endTime, 4);
        assertEq(id, 1);
        assertEq(factoryV3.marketTypes(1), 1); // 1 = multi
        assertEq(MultiOutcomeMarket(marketAddr).outcomeCount(), 4);
    }

    function test_PauseBlocksCreation() public {
        vm.prank(owner);
        factoryV3.pause();
        vm.prank(owner);
        vm.expectRevert();
        factoryV3.createBinaryMarket(MATCH_REF, QUESTION, endTime, 0);
    }

    function test_UnpauseAllowsCreation() public {
        vm.startPrank(owner);
        factoryV3.pause();
        factoryV3.unpause();
        (address addr,) = factoryV3.createBinaryMarket(MATCH_REF, QUESTION, endTime, 0);
        vm.stopPrank();
        assertTrue(addr != address(0));
    }

    function test_SetOracle() public {
        address newOracle = makeAddr("newOracle");
        vm.prank(owner);
        factoryV3.setOracle(newOracle);
        assertEq(factoryV3.oracle(), newOracle);
    }

    function test_SetDefaultFeeBps() public {
        vm.prank(owner);
        factoryV3.setDefaultFeeBps(200);
        assertEq(factoryV3.defaultFeeBps(), 200);
    }

    // =========================================================================
    // 边界测试
    // =========================================================================

    function test_CreateMultipleMarkets_incrementingIds() public {
        vm.startPrank(owner);
        factoryV3.createBinaryMarket(keccak256("b1"), "B1?", endTime, 0);
        factoryV3.createMultiMarket(keccak256("m1"), "M1?", endTime, 3);
        vm.stopPrank();
        assertEq(factoryV3.marketCount(), 2);
        assertEq(factoryV3.marketTypes(1), 0);
        assertEq(factoryV3.marketTypes(2), 1);
    }

    // =========================================================================
    // 错误测试
    // =========================================================================

    function test_RevertWhen_createNotOwner() public {
        vm.prank(alice);
        vm.expectRevert();
        factoryV3.createBinaryMarket(MATCH_REF, QUESTION, endTime, 0);
    }

    function test_RevertWhen_pauseNotOwner() public {
        vm.prank(alice);
        vm.expectRevert();
        factoryV3.pause();
    }

    // =========================================================================
    // 事件测试
    // =========================================================================

    function test_Event_binaryMarketCreated() public {
        vm.expectEmit(true, false, false, false, address(factoryV3));
        emit MarketFactoryV3.BinaryMarketCreated(1, address(0), MATCH_REF, QUESTION);
        vm.prank(owner);
        factoryV3.createBinaryMarket(MATCH_REF, QUESTION, endTime, 0);
    }

    function test_Event_multiMarketCreated() public {
        vm.expectEmit(true, false, false, false, address(factoryV3));
        emit MarketFactoryV3.MultiMarketCreated(1, address(0), 3, "Winner?");
        vm.prank(owner);
        factoryV3.createMultiMarket(keccak256("g1"), "Winner?", endTime, 3);
    }

    // =========================================================================
    // Gas 测试
    // =========================================================================

    function test_Gas_createBinaryMarket() public {
        uint256 gasBefore = gasleft();
        vm.prank(owner);
        factoryV3.createBinaryMarket(MATCH_REF, QUESTION, endTime, 0);
        uint256 gasUsed = gasBefore - gasleft();
        assertTrue(gasUsed > 0 && gasUsed < 4_000_000);
    }

    function test_Gas_createMultiMarket() public {
        uint256 gasBefore = gasleft();
        vm.prank(owner);
        factoryV3.createMultiMarket(keccak256("g1"), "Winner?", endTime, 3);
        uint256 gasUsed = gasBefore - gasleft();
        assertTrue(gasUsed > 0 && gasUsed < 3_000_000);
    }
}
