// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./BaseSetup.sol";

/// @title MarketFactoryTest
/// @notice Phase2 市场工厂测试：创建市场、权限、事件与 Gas
contract MarketFactoryTest is BaseSetup {
    function setUp() public override {
        super.setUp();
        _deployPhase2Stack(0);
    }

    // =========================================================================
    // 部署测试
    // =========================================================================

    function test_Deploy_immutableCollateralAndOracle() public view {
        assertEq(address(factory.collateral()), address(usdc));
        assertEq(factory.oracle(), address(adapter));
        assertEq(factory.owner(), owner);
        assertEq(factory.marketCount(), 0);
    }

    function test_Deploy_version() public view {
        assertEq(factory.version(), "2.0.0-phase2");
    }

    // =========================================================================
    // 功能测试
    // =========================================================================

    function test_CreateMarket_incrementsCount() public {
        vm.prank(owner);
        factory.createMarket(MATCH_REF, QUESTION, _futureEndTime(ONE_DAY));
        assertEq(factory.marketCount(), 1);
    }

    function test_CreateMarket_storesMarketAddress() public {
        vm.prank(owner);
        (address marketAddr,) = factory.createMarket(MATCH_REF, QUESTION, _futureEndTime(ONE_DAY));
        assertEq(factory.markets(1), marketAddr);
        assertTrue(marketAddr != address(0));
    }

    function test_CreateMarket_wiresOracleToAdapter() public {
        vm.prank(owner);
        (address marketAddr,) = factory.createMarket(MATCH_REF, QUESTION, _futureEndTime(ONE_DAY));
        PredictionMarket market = PredictionMarket(marketAddr);
        assertEq(market.oracle(), address(adapter));
        assertEq(market.question(), QUESTION);
    }

    function test_SetOracle_updatesOracle() public {
        address newOracle = makeAddr("newOracle");
        vm.prank(owner);
        factory.setOracle(newOracle);
        assertEq(factory.oracle(), newOracle);
    }

    function test_CreateMultipleMarkets_uniqueIds() public {
        vm.startPrank(owner);
        (address m1,) = factory.createMarket(keccak256("m1"), "Q1?", _futureEndTime(ONE_DAY));
        (address m2,) = factory.createMarket(keccak256("m2"), "Q2?", _futureEndTime(ONE_DAY));
        vm.stopPrank();
        assertEq(factory.marketCount(), 2);
        assertTrue(m1 != m2);
        assertEq(factory.markets(1), m1);
        assertEq(factory.markets(2), m2);
    }

    // =========================================================================
    // 边界测试
    // =========================================================================

    function test_CreateMarket_endTimeJustAfterNow() public {
        uint256 end = block.timestamp + 1;
        vm.prank(owner);
        (address marketAddr,) = factory.createMarket(MATCH_REF, QUESTION, end);
        assertEq(PredictionMarket(marketAddr).endTime(), end);
    }

    // =========================================================================
    // 错误测试
    // =========================================================================

    function test_RevertWhen_createMarketNotOwner() public {
        vm.prank(alice);
        vm.expectRevert();
        factory.createMarket(MATCH_REF, QUESTION, _futureEndTime(ONE_DAY));
    }

    function test_RevertWhen_setOracleZeroAddress() public {
        vm.prank(owner);
        vm.expectRevert("oracle");
        factory.setOracle(address(0));
    }

    function test_RevertWhen_createMarketEndInPast() public {
        vm.prank(owner);
        vm.expectRevert("end in past");
        factory.createMarket(MATCH_REF, QUESTION, block.timestamp);
    }

    function test_RevertWhen_deployFactoryZeroCollateral() public {
        vm.expectRevert("collateral");
        new MarketFactory(address(0), address(adapter));
    }

    // =========================================================================
    // 事件测试：仅校验 marketId 与 matchRef 等 indexed 字段
    // =========================================================================

    function test_Event_marketCreated() public {
        uint256 end = _futureEndTime(ONE_DAY);
        // checkTopic1=id, checkTopic3=matchRef；不校验动态部署的 market 地址
        vm.expectEmit(true, false, true, false, address(factory));
        emit MarketFactory.MarketCreated(1, address(0), MATCH_REF, QUESTION, end);
        vm.prank(owner);
        factory.createMarket(MATCH_REF, QUESTION, end);
    }

    // =========================================================================
    // Gas 测试
    // =========================================================================

    function test_Gas_createMarket() public {
        uint256 gasBefore = gasleft();
        vm.prank(owner);
        factory.createMarket(MATCH_REF, QUESTION, _futureEndTime(ONE_DAY));
        uint256 gasUsed = gasBefore - gasleft();
        assertTrue(gasUsed > 0 && gasUsed < 3_000_000);
    }
}
