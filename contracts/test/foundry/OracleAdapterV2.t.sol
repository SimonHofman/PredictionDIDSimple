// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./BaseSetup.sol";

/// @title OracleAdapterV2Test
/// @notice Phase3 多签预言机适配器测试（m-of-n propose + approve）
contract OracleAdapterV2Test is BaseSetup {
    address internal marketAddr;

    function setUp() public override {
        super.setUp();
        _deployPhase2Stack(0); // 复用 usdc 部署
        _deployMultisigAdapter(2); // 2-of-2 多签
        marketAddr = _createMarketForMultisig();
    }

    function _deployMultisigAdapter(uint256 threshold) internal {
        vm.startPrank(owner);
        adapterV2 = new OracleAdapterV2(owner, threshold);
        adapterV2.grantOracle(oracle);
        adapterV2.grantOracle(oracle2);
        vm.stopPrank();
    }

    function _createMarketForMultisig() internal returns (address) {
        vm.prank(owner);
        MarketFactory multisigFactory = new MarketFactory(address(usdc), address(adapterV2));
        vm.prank(owner);
        (address addr,) = multisigFactory.createMarket(MATCH_REF, QUESTION, _futureEndTime(ONE_DAY));
        return addr;
    }

    // =========================================================================
    // 部署测试
    // =========================================================================

    function test_Deploy_threshold() public view {
        assertEq(adapterV2.threshold(), 2);
        assertEq(adapterV2.proposalCount(), 0);
    }

    // =========================================================================
    // 功能测试
    // =========================================================================

    function test_ProposeResolve_autoApprovesProposer() public {
        vm.prank(oracle);
        uint256 id = adapterV2.proposeResolve(marketAddr, 0);
        assertEq(id, 1);
        (address market,, uint256 approvals, bool executed) = adapterV2.proposals(1);
        assertEq(approvals, 1); // 提议者自动计入一票
        assertFalse(executed);
        market; // silence unused warning
    }

    function test_ApproveResolve_reachesThreshold() public {
        vm.prank(oracle);
        adapterV2.proposeResolve(marketAddr, 0);

        vm.prank(oracle2);
        adapterV2.approveResolve(1); // 第二票触发执行

        (,,, bool executed) = adapterV2.proposals(1);
        assertTrue(executed);
        assertEq(uint8(PredictionMarket(marketAddr).status()), uint8(PredictionMarket.Status.Resolved));
    }

    function test_SingleOracleThreshold_executesImmediately() public {
        vm.startPrank(owner);
        OracleAdapterV2 single = new OracleAdapterV2(owner, 1);
        single.grantOracle(oracle);
        MarketFactory f = new MarketFactory(address(usdc), address(single));
        (address addr,) = f.createMarket(MATCH_REF, QUESTION, _futureEndTime(ONE_DAY));
        vm.stopPrank();

        vm.prank(oracle);
        single.proposeResolve(addr, 1); // 1-of-1 提议即执行
        assertEq(uint8(PredictionMarket(addr).winningOutcome()), 1);
    }

    function test_VoidMarket() public {
        vm.prank(oracle);
        adapterV2.voidMarket(marketAddr);
        assertEq(uint8(PredictionMarket(marketAddr).status()), uint8(PredictionMarket.Status.Voided));
    }

    function test_SetThreshold() public {
        vm.prank(owner);
        adapterV2.setThreshold(1);
        assertEq(adapterV2.threshold(), 1);
    }

    // =========================================================================
    // 边界测试：最大 outcome 索引（多结果市场 outcome 7）
    // =========================================================================

    function test_ProposeResolve_maxOutcome() public {
        vm.startPrank(owner);
        MarketFactoryV3 f3 = new MarketFactoryV3(address(usdc), address(adapterV2), 0);
        (address multiAddr,) = f3.createMultiMarket(keccak256("multi"), "Q?", _futureEndTime(ONE_DAY), 8);
        vm.stopPrank();

        vm.prank(oracle);
        adapterV2.proposeResolve(multiAddr, 7);
        vm.prank(oracle2);
        adapterV2.approveResolve(1);
        assertEq(MultiOutcomeMarket(multiAddr).winningOutcome(), 7);
    }

    // =========================================================================
    // 错误测试
    // =========================================================================

    function test_RevertWhen_doubleApprove() public {
        vm.prank(oracle);
        uint256 id = adapterV2.proposeResolve(marketAddr, 0);
        vm.prank(oracle);
        vm.expectRevert("approved");
        adapterV2.approveResolve(id);
    }

    function test_RevertWhen_approveAfterExecuted() public {
        vm.prank(oracle);
        adapterV2.proposeResolve(marketAddr, 0);
        vm.prank(oracle2);
        adapterV2.approveResolve(1);
        vm.prank(oracle2);
        vm.expectRevert(bytes("executed"));
        adapterV2.approveResolve(1);
    }

    function test_RevertWhen_nonOracleProposes() public {
        vm.prank(alice);
        vm.expectRevert();
        adapterV2.proposeResolve(marketAddr, 0);
    }

    function test_RevertWhen_setThresholdZero() public {
        vm.prank(owner);
        vm.expectRevert("threshold");
        adapterV2.setThreshold(0);
    }

    function test_RevertWhen_proposeClosedMarket() public {
        vm.prank(oracle);
        adapterV2.proposeResolve(marketAddr, 0);
        vm.prank(oracle2);
        adapterV2.approveResolve(1);
        vm.prank(oracle);
        vm.expectRevert("not open");
        adapterV2.proposeResolve(marketAddr, 0);
    }

    // =========================================================================
    // 事件测试
    // =========================================================================

    function test_Event_proposalCreated() public {
        vm.expectEmit(true, true, false, true, address(adapterV2));
        emit OracleAdapterV2.ProposalCreated(1, marketAddr, 0);
        vm.prank(oracle);
        adapterV2.proposeResolve(marketAddr, 0);
    }

    function test_Event_proposalApproved() public {
        vm.prank(oracle);
        adapterV2.proposeResolve(marketAddr, 0);
        vm.expectEmit(true, true, false, false, address(adapterV2));
        emit OracleAdapterV2.ProposalApproved(1, oracle2);
        vm.prank(oracle2);
        adapterV2.approveResolve(1);
    }

    function test_Event_proposalExecuted() public {
        vm.prank(oracle);
        adapterV2.proposeResolve(marketAddr, 0);
        vm.expectEmit(true, false, false, true, address(adapterV2));
        emit OracleAdapterV2.ProposalExecuted(1, 0);
        vm.prank(oracle2);
        adapterV2.approveResolve(1);
    }

    // =========================================================================
    // Gas 测试
    // =========================================================================

    function test_Gas_proposeResolve() public {
        uint256 gasBefore = gasleft();
        vm.prank(oracle);
        adapterV2.proposeResolve(marketAddr, 0);
        uint256 gasUsed = gasBefore - gasleft();
        assertTrue(gasUsed > 0 && gasUsed < 300_000);
    }
}
