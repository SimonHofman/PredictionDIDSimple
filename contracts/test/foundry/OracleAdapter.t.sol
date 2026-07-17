// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./BaseSetup.sol";

/// @title OracleAdapterTest
/// @notice Phase2 时间锁预言机适配器测试
/// @dev 覆盖 requestResolve → confirmResolve 两阶段流程及 voidMarket
contract OracleAdapterTest is BaseSetup {
    PredictionMarket internal market;
    address internal marketAddr;

    function setUp() public override {
        super.setUp();
        _deployPhase2Stack(120); // 120 秒时间锁
        marketAddr = _createMarket();
        market = PredictionMarket(marketAddr);
        _mintAndApprove(alice, marketAddr, 500 * USDC_UNIT);
    }

    function _createMarket() internal returns (address) {
        vm.prank(owner);
        (address addr,) = factory.createMarket(MATCH_REF, QUESTION, _futureEndTime(ONE_DAY));
        return addr;
    }

    // =========================================================================
    // 部署测试
    // =========================================================================

    function test_Deploy_timelockDelay() public view {
        assertEq(adapter.timelockDelay(), 120);
    }

    function test_Deploy_oracleHasRole() public view {
        assertTrue(adapter.hasRole(adapter.ORACLE_ROLE(), oracle));
        assertTrue(adapter.hasRole(adapter.DEFAULT_ADMIN_ROLE(), owner));
    }

    // =========================================================================
    // 功能测试
    // =========================================================================

    function test_RequestAndConfirmResolve_timelockedFlow() public {
        vm.prank(oracle);
        adapter.requestResolve(marketAddr, 0);

        (uint8 outcome, uint256 executeAfter, bool active) = adapter.pending(marketAddr);
        assertTrue(active);
        assertEq(outcome, 0);
        assertEq(executeAfter, block.timestamp + 120);

        vm.warp(block.timestamp + 121); // 满足时间锁
        vm.prank(oracle);
        adapter.confirmResolve(marketAddr);
        assertEq(uint8(market.status()), uint8(PredictionMarket.Status.Resolved));
    }

    function test_ResolveNow_zeroTimelock() public {
        // timelock=0 时走 resolveNow 快速路径
        vm.startPrank(owner);
        OracleAdapter fastAdapter = new OracleAdapter(owner, 0);
        fastAdapter.grantOracle(oracle);
        MarketFactory fastFactory = new MarketFactory(address(usdc), address(fastAdapter));
        (address addr,) = fastFactory.createMarket(MATCH_REF, QUESTION, _futureEndTime(ONE_DAY));
        vm.stopPrank();
        PredictionMarket m = PredictionMarket(addr);

        vm.prank(oracle);
        fastAdapter.resolveNow(addr, 1);
        assertEq(uint8(m.status()), uint8(PredictionMarket.Status.Resolved));
        assertEq(m.winningOutcome(), 1);
    }

    function test_VoidMarket_refundsUser() public {
        uint256 stake = 40 * USDC_UNIT;
        vm.prank(alice);
        market.buy(0, stake);

        vm.prank(oracle);
        adapter.voidMarket(marketAddr);

        assertEq(uint8(market.status()), uint8(PredictionMarket.Status.Voided));
        uint256 before = usdc.balanceOf(alice);
        vm.prank(alice);
        market.claim();
        assertEq(usdc.balanceOf(alice), before + stake);
    }

    function test_SetTimelockDelay() public {
        vm.prank(owner);
        adapter.setTimelockDelay(300);
        assertEq(adapter.timelockDelay(), 300);
    }

    function test_GrantOracle() public {
        address newOracle = makeAddr("newOracle");
        vm.prank(owner);
        adapter.grantOracle(newOracle);
        assertTrue(adapter.hasRole(adapter.ORACLE_ROLE(), newOracle));
    }

    // =========================================================================
    // 边界测试
    // =========================================================================

    function test_RequestResolve_outcomeOne() public {
        vm.prank(oracle);
        adapter.requestResolve(marketAddr, 1);
        (uint8 outcome,,) = adapter.pending(marketAddr);
        assertEq(outcome, 1);
    }

    // =========================================================================
    // 错误测试
    // =========================================================================

    function test_RevertWhen_confirmBeforeTimelock() public {
        vm.prank(oracle);
        adapter.requestResolve(marketAddr, 0);
        vm.prank(oracle);
        vm.expectRevert("timelock");
        adapter.confirmResolve(marketAddr);
    }

    function test_RevertWhen_confirmWithoutPending() public {
        vm.prank(oracle);
        vm.expectRevert("no pending");
        adapter.confirmResolve(marketAddr);
    }

    function test_RevertWhen_resolveNowWithTimelock() public {
        vm.prank(oracle);
        vm.expectRevert("use request+confirm");
        adapter.resolveNow(marketAddr, 0);
    }

    function test_RevertWhen_nonOracleRequests() public {
        vm.prank(alice);
        vm.expectRevert();
        adapter.requestResolve(marketAddr, 0);
    }

    function test_RevertWhen_invalidOutcome() public {
        vm.prank(oracle);
        vm.expectRevert("invalid outcome");
        adapter.requestResolve(marketAddr, 2);
    }

    function test_RevertWhen_voidClosedMarket() public {
        vm.prank(oracle);
        adapter.requestResolve(marketAddr, 0);
        vm.warp(block.timestamp + 121);
        vm.prank(oracle);
        adapter.confirmResolve(marketAddr);
        vm.prank(oracle);
        vm.expectRevert("not open");
        adapter.voidMarket(marketAddr);
    }

    // =========================================================================
    // 事件测试
    // =========================================================================

    function test_Event_oracleResolveRequested() public {
        uint256 executeAfter = block.timestamp + 120;
        vm.expectEmit(true, false, false, true, address(adapter));
        emit OracleAdapter.OracleResolveRequested(marketAddr, 0, executeAfter);
        vm.prank(oracle);
        adapter.requestResolve(marketAddr, 0);
    }

    function test_Event_oracleResolveConfirmed() public {
        vm.prank(oracle);
        adapter.requestResolve(marketAddr, 0);
        vm.warp(block.timestamp + 121);
        vm.expectEmit(true, false, false, true, address(adapter));
        emit OracleAdapter.OracleResolveConfirmed(marketAddr, 0);
        vm.prank(oracle);
        adapter.confirmResolve(marketAddr);
    }

    function test_Event_marketVoided() public {
        vm.expectEmit(true, false, false, false, address(adapter));
        emit OracleAdapter.MarketVoided(marketAddr);
        vm.prank(oracle);
        adapter.voidMarket(marketAddr);
    }

    // =========================================================================
    // Gas 测试
    // =========================================================================

    function test_Gas_requestResolve() public {
        uint256 gasBefore = gasleft();
        vm.prank(oracle);
        adapter.requestResolve(marketAddr, 0);
        uint256 gasUsed = gasBefore - gasleft();
        assertTrue(gasUsed > 0 && gasUsed < 150_000);
    }
}
