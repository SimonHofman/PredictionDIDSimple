// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./BaseSetup.sol";

/// @title MultiOutcomeMarketTest
/// @notice Phase3 多结果彩池市场测试（2–8 个 outcome）
contract MultiOutcomeMarketTest is BaseSetup {
    MultiOutcomeMarket internal market;
    address internal marketAddr;
    uint8 internal constant OUTCOMES = 3;
    uint256 internal endTime;

    function setUp() public override {
        super.setUp();
        _deployPhase3Stack(50, 1); // 0.5% 手续费
        endTime = _futureEndTime(ONE_DAY);
        marketAddr = _createMultiMarket();
        market = MultiOutcomeMarket(marketAddr);
        _mintAndApprove(alice, marketAddr, 1000 * USDC_UNIT);
    }

    function _createMultiMarket() internal returns (address) {
        vm.prank(owner);
        (address addr,) =
            factoryV3.createMultiMarket(keccak256("group-1"), "Group winner?", endTime, OUTCOMES);
        return addr;
    }

    // =========================================================================
    // 部署测试
    // =========================================================================

    function test_Deploy_outcomeCount() public view {
        assertEq(market.outcomeCount(), OUTCOMES);
    }

    function test_Deploy_poolsInitialized() public view {
        for (uint8 i = 0; i < OUTCOMES; i++) {
            assertEq(market.pool(i), 0);
        }
    }

    function test_Deploy_feeBps() public view {
        assertEq(market.feeBps(), 50);
    }

    // =========================================================================
    // 功能测试
    // =========================================================================

    function test_Buy_updatesPoolAndStake() public {
        uint256 amount = 100 * USDC_UNIT;
        vm.prank(alice);
        market.buy(0, amount);
        uint256 net = amount - (amount * 50) / 10_000; // 扣费后净额入池
        assertEq(market.pool(0), net);
        assertEq(market.stake(alice, 0), net);
    }

    function test_Buy_multipleOutcomes() public {
        vm.prank(alice);
        market.buy(0, 100 * USDC_UNIT);
        vm.prank(alice);
        market.buy(1, 50 * USDC_UNIT);
        assertGt(market.stake(alice, 0), 0);
        assertGt(market.stake(alice, 1), 0);
    }

    function test_ResolveAndClaim_winnerPaid() public {
        vm.prank(alice);
        market.buy(0, 100 * USDC_UNIT);
        vm.prank(alice);
        market.buy(1, 50 * USDC_UNIT);

        vm.prank(address(adapterV2));
        market.resolve(0);

        uint256 before = usdc.balanceOf(alice);
        vm.prank(alice);
        market.claim();
        assertGt(usdc.balanceOf(alice), before);
    }

    function test_VoidMarket_refundsAllStakes() public {
        vm.prank(alice);
        market.buy(0, 100 * USDC_UNIT);
        vm.prank(alice);
        market.buy(2, 30 * USDC_UNIT);

        vm.prank(address(adapterV2));
        market.voidMarket();

        uint256 before = usdc.balanceOf(alice);
        vm.prank(alice);
        market.claim();
        assertGt(usdc.balanceOf(alice), before);
    }

    // =========================================================================
    // 边界测试：首尾 outcome、最小/最大 outcomeCount
    // =========================================================================

    function test_Buy_firstOutcome() public {
        vm.prank(alice);
        market.buy(0, 1);
        assertGt(market.pool(0), 0);
    }

    function test_Buy_lastOutcome() public {
        vm.prank(alice);
        market.buy(OUTCOMES - 1, 10 * USDC_UNIT);
        assertGt(market.pool(OUTCOMES - 1), 0);
    }

    function test_Deploy_minOutcomes() public {
        vm.prank(owner);
        (address addr,) = factoryV3.createMultiMarket(keccak256("min"), "Min?", endTime, 2);
        assertEq(MultiOutcomeMarket(addr).outcomeCount(), 2);
    }

    function test_Deploy_maxOutcomes() public {
        vm.prank(owner);
        (address addr,) = factoryV3.createMultiMarket(keccak256("max"), "Max?", endTime, 8);
        assertEq(MultiOutcomeMarket(addr).outcomeCount(), 8);
    }

    // =========================================================================
    // 错误测试
    // =========================================================================

    function test_RevertWhen_buyInvalidOutcome() public {
        vm.prank(alice);
        vm.expectRevert("invalid");
        market.buy(OUTCOMES, 10 * USDC_UNIT);
    }

    function test_RevertWhen_buyZeroAmount() public {
        vm.prank(alice);
        vm.expectRevert("invalid");
        market.buy(0, 0);
    }

    function test_RevertWhen_buyAfterEnd() public {
        vm.warp(endTime);
        vm.prank(alice);
        vm.expectRevert("closed");
        market.buy(0, 10 * USDC_UNIT);
    }

    function test_RevertWhen_nonOracleResolves() public {
        vm.prank(alice);
        vm.expectRevert("not oracle");
        market.resolve(0);
    }

    function test_RevertWhen_loserClaims() public {
        vm.prank(alice);
        market.buy(0, 100 * USDC_UNIT);
        vm.prank(bob);
        _mintAndApprove(bob, marketAddr, 100 * USDC_UNIT);
        vm.prank(bob);
        market.buy(1, 100 * USDC_UNIT);

        vm.prank(address(adapterV2));
        market.resolve(0);

        vm.prank(bob);
        vm.expectRevert("nothing");
        market.claim();
    }

    function test_RevertWhen_tooFewOutcomes() public {
        vm.prank(owner);
        vm.expectRevert("outcomes");
        factoryV3.createMultiMarket(keccak256("bad"), "Bad?", endTime, 1);
    }

    function test_RevertWhen_tooManyOutcomes() public {
        vm.prank(owner);
        vm.expectRevert("outcomes");
        factoryV3.createMultiMarket(keccak256("bad2"), "Bad?", endTime, 9);
    }

    // =========================================================================
    // 事件测试
    // =========================================================================

    function test_Event_bought() public {
        uint256 amount = 25 * USDC_UNIT;
        vm.expectEmit(true, false, false, true, marketAddr);
        emit MultiOutcomeMarket.Bought(alice, 0, amount);
        vm.prank(alice);
        market.buy(0, amount);
    }

    function test_Event_resolved() public {
        vm.expectEmit(false, false, false, true, marketAddr);
        emit MultiOutcomeMarket.Resolved(1);
        vm.prank(address(adapterV2));
        market.resolve(1);
    }

    // =========================================================================
    // Gas 测试
    // =========================================================================

    function test_Gas_buy() public {
        uint256 gasBefore = gasleft();
        vm.prank(alice);
        market.buy(0, 100 * USDC_UNIT);
        uint256 gasUsed = gasBefore - gasleft();
        assertTrue(gasUsed > 0 && gasUsed < 200_000);
    }

    function test_Gas_claim() public {
        vm.prank(alice);
        market.buy(0, 100 * USDC_UNIT);
        vm.prank(address(adapterV2));
        market.resolve(0);
        uint256 gasBefore = gasleft();
        vm.prank(alice);
        market.claim();
        uint256 gasUsed = gasBefore - gasleft();
        assertTrue(gasUsed > 0 && gasUsed < 200_000);
    }
}
