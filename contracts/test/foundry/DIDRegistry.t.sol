// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "./BaseSetup.sol";

/// @title DIDRegistryTest
/// @notice DID 哈希链上绑定测试：签名验证与查询
contract DIDRegistryTest is BaseSetup {
    uint256 internal aliceKey = 0xA11CE;
    address internal aliceAccount;

    function setUp() public override {
        super.setUp();
        aliceAccount = vm.addr(aliceKey); // 由私钥派生地址
        vm.prank(owner);
        didRegistry = new DIDRegistry();
    }

    // =========================================================================
    // 部署测试
    // =========================================================================

    function test_Deploy_ownerSet() public view {
        assertEq(didRegistry.owner(), owner);
    }

    function test_Deploy_noInitialBinding() public view {
        assertEq(didRegistry.didHashOf(aliceAccount), bytes32(0));
    }

    // =========================================================================
    // 功能测试
    // =========================================================================

    function test_BindDid_storesHash() public {
        bytes32 didHash = keccak256("did:example:alice");
        bytes memory sig = _signBindDid(aliceAccount, didHash, aliceKey);
        vm.prank(aliceAccount);
        didRegistry.bindDid(didHash, sig);
        assertEq(didRegistry.didHashOf(aliceAccount), didHash);
    }

    function test_ResolveDid_returnsHash() public {
        bytes32 didHash = keccak256("did:example:bob");
        uint256 bobKey = 0xB0B;
        address bobAccount = vm.addr(bobKey);
        bytes memory sig = _signBindDid(bobAccount, didHash, bobKey);
        vm.prank(bobAccount);
        didRegistry.bindDid(didHash, sig);
        assertEq(didRegistry.resolveDid(bobAccount), didHash);
    }

    function test_BindDid_overwritesPrevious() public {
        bytes32 first = keccak256("did:first");
        bytes32 second = keccak256("did:second");
        vm.prank(aliceAccount);
        didRegistry.bindDid(first, _signBindDid(aliceAccount, first, aliceKey));
        vm.prank(aliceAccount);
        didRegistry.bindDid(second, _signBindDid(aliceAccount, second, aliceKey));
        assertEq(didRegistry.didHashOf(aliceAccount), second);
    }

    // =========================================================================
    // 边界测试
    // =========================================================================

    function test_BindDid_maxHash() public {
        bytes32 didHash = bytes32(type(uint256).max);
        vm.prank(aliceAccount);
        didRegistry.bindDid(didHash, _signBindDid(aliceAccount, didHash, aliceKey));
        assertEq(didRegistry.didHashOf(aliceAccount), didHash);
    }

    // =========================================================================
    // 错误测试
    // =========================================================================

    function test_RevertWhen_emptyDidHash() public {
        vm.prank(aliceAccount);
        vm.expectRevert("empty did");
        didRegistry.bindDid(bytes32(0), "");
    }

    function test_RevertWhen_invalidSignature() public {
        bytes32 didHash = keccak256("did:bad");
        uint256 wrongKey = 0xDEAD;
        bytes memory sig = _signBindDid(vm.addr(wrongKey), didHash, wrongKey);
        vm.prank(aliceAccount);
        vm.expectRevert("invalid sig");
        didRegistry.bindDid(didHash, sig);
    }

    function test_RevertWhen_signatureFromOtherAccount() public {
        bytes32 didHash = keccak256("did:other");
        uint256 bobKey = 0xB0B;
        bytes memory sig = _signBindDid(vm.addr(bobKey), didHash, bobKey);
        vm.prank(aliceAccount);
        vm.expectRevert("invalid sig");
        didRegistry.bindDid(didHash, sig);
    }

    // =========================================================================
    // 事件测试
    // =========================================================================

    function test_Event_didBound() public {
        bytes32 didHash = keccak256("did:event");
        vm.expectEmit(true, true, false, false, address(didRegistry));
        emit DIDRegistry.DidBound(aliceAccount, didHash);
        vm.prank(aliceAccount);
        didRegistry.bindDid(didHash, _signBindDid(aliceAccount, didHash, aliceKey));
    }

    // =========================================================================
    // Gas 测试
    // =========================================================================

    function test_Gas_bindDid() public {
        bytes32 didHash = keccak256("did:gas");
        bytes memory sig = _signBindDid(aliceAccount, didHash, aliceKey);
        uint256 gasBefore = gasleft();
        vm.prank(aliceAccount);
        didRegistry.bindDid(didHash, sig);
        uint256 gasUsed = gasBefore - gasleft();
        assertTrue(gasUsed > 0 && gasUsed < 100_000);
    }
}
