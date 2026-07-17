// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./BaseSetup.sol";

/// @title PredictionMarketTest
/// @notice Phase2 二元彩池市场（parimutuel）完整测试套件
/// @dev 覆盖：buy / resolve / void / claim 全生命周期及权限与事件
contract PredictionMarketTest is BaseSetup {
    PredictionMarket internal market;
    address internal marketAddr;
    uint256 internal endTime;

    /// @notice 每个测试前独立部署市场；timelock=0 便于直接 resolve
    function setUp() public override {
        super.setUp();
        _deployPhase2Stack(0);
        endTime = _futureEndTime(ONE_WEEK);
        (market, marketAddr) = _createBinaryMarket(endTime);
        _mintAndApprove(alice, marketAddr, 1000 * USDC_UNIT);
        _mintAndApprove(bob, marketAddr, 1000 * USDC_UNIT);
    }

    // =========================================================================
    // 部署测试：immutable 字段与初始 Open 状态
    // =========================================================================

    function test_Deploy_immutableFields() public view {
        assertEq(address(market.collateral()), address(usdc));
        assertEq(market.oracle(), address(adapter));
        assertEq(market.factory(), address(factory));
        assertEq(market.matchRef(), MATCH_REF);
        assertEq(market.question(), QUESTION);
        assertEq(market.endTime(), endTime);
        assertEq(uint8(market.status()), uint8(PredictionMarket.Status.Open));
    }

    function test_Deploy_poolsStartAtZero() public view {
        assertEq(market.yesPool(), 0);
        assertEq(market.noPool(), 0);
    }

    // =========================================================================
    // 功能测试：下注、结算、领奖、作废退款
    // =========================================================================

    function test_Buy_yesUpdatesPool() public {
        uint256 amount = 100 * USDC_UNIT;
        vm.prank(alice);
        market.buy(0, amount); // outcome 0 = Yes
        assertEq(market.yesPool(), amount);
        assertEq(market.yesBalance(alice), amount);
        assertEq(market.noPool(), 0);
    }

    function test_Buy_noUpdatesPool() public {
        uint256 amount = 50 * USDC_UNIT;
        vm.prank(bob);
        market.buy(1, amount); // outcome 1 = No
        assertEq(market.noPool(), amount);
        assertEq(market.noBalance(bob), amount);
    }

    function test_ResolveAndClaim_winnerGetsFullPool() public {
        vm.prank(alice);
        market.buy(0, 100 * USDC_UNIT);
        vm.prank(bob);
        market.buy(1, 100 * USDC_UNIT);

        // 市场 oracle 指向 Adapter 合约地址，故用 adapter 身份 resolve
        vm.prank(address(adapter));
        market.resolve(0);

        uint256 before = usdc.balanceOf(alice);
        vm.prank(alice);
        market.claim();
        // 彩池分配：alice 押 Yes 赢，获得双方总池 200 USDC
        assertEq(usdc.balanceOf(alice), before + 200 * USDC_UNIT);
    }

    function test_VoidMarket_refundsStake() public {
        uint256 stake = 40 * USDC_UNIT;
        vm.prank(alice);
        market.buy(0, stake);

        vm.prank(address(adapter));
        market.voidMarket();

        uint256 before = usdc.balanceOf(alice);
        vm.prank(alice);
        market.claim();
        assertEq(usdc.balanceOf(alice), before + stake);
    }

    // =========================================================================
    // 边界测试：最小金额、双边下注、单侧池结算
    // =========================================================================

    function test_Buy_minimumAmount() public {
        vm.prank(alice);
        market.buy(0, 1);
        assertEq(market.yesPool(), 1);
    }

    function test_Buy_bothOutcomes() public {
        vm.prank(alice);
        market.buy(0, 100 * USDC_UNIT);
        vm.prank(alice);
        market.buy(1, 50 * USDC_UNIT);
        assertEq(market.yesBalance(alice), 100 * USDC_UNIT);
        assertEq(market.noBalance(alice), 50 * USDC_UNIT);
    }

    function test_Claim_singleSidePool() public {
        // 仅 Yes 侧有下注时，赢家仍可取走全部池子
        vm.prank(alice);
        market.buy(0, 100 * USDC_UNIT);
        vm.prank(address(adapter));
        market.resolve(0);
        uint256 before = usdc.balanceOf(alice);
        vm.prank(alice);
        market.claim();
        assertEq(usdc.balanceOf(alice), before + 100 * USDC_UNIT);
    }

    // =========================================================================
    // 错误测试：时间、参数、权限、重复领取
    // =========================================================================

    function test_RevertWhen_buyAfterEndTime() public {
        vm.warp(endTime); // 快进到截止时间
        vm.prank(alice);
        vm.expectRevert("ended");
        market.buy(0, 10 * USDC_UNIT);
    }

    function test_RevertWhen_buyInvalidOutcome() public {
        vm.prank(alice);
        vm.expectRevert("invalid outcome");
        market.buy(2, 10 * USDC_UNIT);
    }

    function test_RevertWhen_buyZeroAmount() public {
        vm.prank(alice);
        vm.expectRevert("zero amount");
        market.buy(0, 0);
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
        market.buy(1, 100 * USDC_UNIT);
        vm.prank(address(adapter));
        market.resolve(0);
        vm.prank(bob);
        vm.expectRevert("nothing to claim");
        market.claim();
    }

    function test_RevertWhen_doubleClaim() public {
        vm.prank(alice);
        market.buy(0, 50 * USDC_UNIT);
        vm.prank(address(adapter));
        market.resolve(0);
        vm.prank(alice);
        market.claim();
        vm.prank(alice);
        vm.expectRevert("already claimed");
        market.claim();
    }

    function test_RevertWhen_claimWhileOpen() public {
        vm.prank(alice);
        market.buy(0, 10 * USDC_UNIT);
        vm.prank(alice);
        vm.expectRevert("not claimable");
        market.claim();
    }

    // =========================================================================
    // 事件测试：vm.expectEmit 校验 Bought / Resolved / Claimed
    // =========================================================================

    function test_Event_bought() public {
        uint256 amount = 25 * USDC_UNIT;
        // (checkTopic1=user, checkData=amount)
        vm.expectEmit(true, false, false, true, marketAddr);
        emit PredictionMarket.Bought(alice, 0, amount);
        vm.prank(alice);
        market.buy(0, amount);
    }

    function test_Event_resolved() public {
        vm.prank(address(adapter));
        vm.expectEmit(false, false, false, true, marketAddr);
        emit PredictionMarket.Resolved(1);
        market.resolve(1);
    }

    function test_Event_claimed() public {
        vm.prank(alice);
        market.buy(0, 100 * USDC_UNIT);
        vm.prank(address(adapter));
        market.resolve(0);
        vm.expectEmit(true, false, false, true, marketAddr);
        emit PredictionMarket.Claimed(alice, 100 * USDC_UNIT);
        vm.prank(alice);
        market.claim();
    }

    // =========================================================================
    // Gas 测试：单次 buy / claim 的上限断言
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
        vm.prank(address(adapter));
        market.resolve(0);
        uint256 gasBefore = gasleft();
        vm.prank(alice);
        market.claim();
        uint256 gasUsed = gasBefore - gasleft();
        assertTrue(gasUsed > 0 && gasUsed < 150_000);
    }
}
