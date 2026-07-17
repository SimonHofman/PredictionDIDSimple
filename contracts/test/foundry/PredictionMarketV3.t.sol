// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./BaseSetup.sol";

/// @title PredictionMarketV3Test
/// @notice Phase3 CPMM 二元市场测试：做市、买卖、手续费、LP 与领奖
contract PredictionMarketV3Test is BaseSetup {
    PredictionMarketV3 internal market;
    address internal marketAddr;
    uint256 internal constant INITIAL_LIQ = 1000 * USDC_UNIT;
    uint256 internal endTime;

    function setUp() public override {
        super.setUp();
        _deployPhase3Stack(100, 1); // 1% 手续费，单签 oracle
        endTime = _futureEndTime(ONE_DAY);
        marketAddr = _createBinaryMarketWithLiquidity(INITIAL_LIQ);
        market = PredictionMarketV3(marketAddr);
        _mintAndApprove(alice, marketAddr, 5000 * USDC_UNIT);
    }

    /// @notice 工厂创建市场并注入初始流动性（每侧 perSide）
    function _createBinaryMarketWithLiquidity(uint256 perSide) internal returns (address) {
        _mintAndApprove(owner, address(factoryV3), perSide * 2);
        vm.prank(owner);
        (address addr,) = factoryV3.createBinaryMarket(MATCH_REF, QUESTION, endTime, perSide);
        return addr;
    }

    // =========================================================================
    // 部署测试
    // =========================================================================

    function test_Deploy_seededReserves() public view {
        assertEq(market.reserveYes(), INITIAL_LIQ);
        assertEq(market.reserveNo(), INITIAL_LIQ);
        assertEq(market.totalLPSupply(), INITIAL_LIQ * 2);
    }

    function test_Deploy_feeBps() public view {
        assertEq(market.feeBps(), 100);
    }

    function test_Deploy_getPoolState_initialPrice() public view {
        (uint256 yesR, uint256 noR, uint256 priceYesBps) = market.getPoolState();
        assertEq(yesR, INITIAL_LIQ);
        assertEq(noR, INITIAL_LIQ);
        assertEq(priceYesBps, 5000); // 50% 隐含概率（基点）
    }

    // =========================================================================
    // 功能测试
    // =========================================================================

    function test_Buy_yesShiftsPriceUp() public {
        vm.prank(alice);
        market.buy(0, 100 * USDC_UNIT);
        (,, uint256 priceYesBps) = market.getPoolState();
        assertGt(priceYesBps, 5000);
    }

    function test_Buy_noShiftsPriceDown() public {
        vm.prank(alice);
        market.buy(1, 100 * USDC_UNIT);
        (,, uint256 priceYesBps) = market.getPoolState();
        assertLt(priceYesBps, 5000);
    }

    function test_Buy_collectsFees() public {
        uint256 amount = 100 * USDC_UNIT;
        vm.prank(alice);
        market.buy(0, amount);
        assertEq(market.collectedFees(), (amount * 100) / 10_000);
    }

    function test_AddLiquidity_increasesReserves() public {
        _mintAndApprove(bob, marketAddr, 200 * USDC_UNIT);
        vm.prank(bob);
        market.addLiquidity(200 * USDC_UNIT);
        assertGt(market.reserveYes(), INITIAL_LIQ);
        assertGt(market.reserveNo(), INITIAL_LIQ);
        assertEq(market.lpBalance(bob), 200 * USDC_UNIT);
    }

    function test_RemoveLiquidity_returnsCollateral() public {
        uint256 lp = market.lpBalance(owner);
        assertGt(lp, 0);
        uint256 before = usdc.balanceOf(owner);
        vm.prank(owner);
        market.removeLiquidity(lp / 2);
        assertGt(usdc.balanceOf(owner), before);
    }

    function test_ResolveAndClaim_winnerPaid() public {
        vm.prank(alice);
        market.buy(0, 100 * USDC_UNIT);
        vm.prank(address(adapterV2));
        market.resolve(0);
        uint256 before = usdc.balanceOf(alice);
        vm.prank(alice);
        market.claim();
        assertGt(usdc.balanceOf(alice), before);
    }

    // =========================================================================
    // 边界测试
    // =========================================================================

    function test_Buy_smallAmount() public {
        // CPMM 大额储备下 1 wei 可能换出 0 shares，故用 10000 作为小额度边界
        uint256 amount = 10_000;
        vm.prank(alice);
        market.buy(0, amount);
        assertGt(market.yesBalance(alice), 0);
    }

    function test_Buy_atMaxBetLimit() public {
        uint256 maxBet = market.maxBetPerUser();
        _mintAndApprove(bob, marketAddr, maxBet);
        vm.prank(bob);
        market.buy(0, maxBet);
        assertEq(market.userBetTotal(bob), maxBet);
    }

    // =========================================================================
    // 错误测试
    // =========================================================================

    function test_RevertWhen_buyExceedsMaxBet() public {
        uint256 maxBet = market.maxBetPerUser();
        _mintAndApprove(bob, marketAddr, maxBet + 1);
        vm.prank(bob);
        vm.expectRevert("max bet");
        market.buy(0, maxBet + 1);
    }

    function test_RevertWhen_buyAfterEnd() public {
        vm.warp(endTime);
        vm.prank(alice);
        vm.expectRevert("ended");
        market.buy(0, 10 * USDC_UNIT);
    }

    function test_RevertWhen_removeMoreLpThanBalance() public {
        vm.prank(alice);
        vm.expectRevert(bytes("lp")); // 避免与 overload expectRevert 歧义
        market.removeLiquidity(1);
    }

    function test_RevertWhen_doubleClaim() public {
        vm.prank(alice);
        market.buy(0, 100 * USDC_UNIT);
        vm.prank(address(adapterV2));
        market.resolve(0);
        vm.prank(alice);
        market.claim();
        vm.prank(alice);
        vm.expectRevert("claimed");
        market.claim();
    }

    // =========================================================================
    // 事件测试
    // =========================================================================

    function test_Event_bought() public {
        uint256 amount = 50 * USDC_UNIT;
        // sharesOut 为动态计算值，仅校验 indexed user
        vm.expectEmit(true, false, false, false, marketAddr);
        emit PredictionMarketV3.Bought(alice, 0, amount, 0);
        vm.prank(alice);
        market.buy(0, amount);
    }

    function test_Event_liquidityAdded() public {
        _mintAndApprove(bob, marketAddr, 100 * USDC_UNIT);
        vm.expectEmit(true, false, false, true, marketAddr);
        emit PredictionMarketV3.LiquidityAdded(bob, 100 * USDC_UNIT);
        vm.prank(bob);
        market.addLiquidity(100 * USDC_UNIT);
    }

    // =========================================================================
    // Gas 测试
    // =========================================================================

    function test_Gas_buy() public {
        uint256 gasBefore = gasleft();
        vm.prank(alice);
        market.buy(0, 100 * USDC_UNIT);
        uint256 gasUsed = gasBefore - gasleft();
        assertTrue(gasUsed > 0 && gasUsed < 250_000);
    }

    function test_Gas_addLiquidity() public {
        _mintAndApprove(bob, marketAddr, 100 * USDC_UNIT);
        uint256 gasBefore = gasleft();
        vm.prank(bob);
        market.addLiquidity(100 * USDC_UNIT);
        uint256 gasUsed = gasBefore - gasleft();
        assertTrue(gasUsed > 0 && gasUsed < 200_000);
    }
}
