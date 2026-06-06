// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";

/// @title DIDRegistry — optional on-chain DID hash binding
contract DIDRegistry is Ownable {
    using ECDSA for bytes32;
    using MessageHashUtils for bytes32;

    mapping(address => bytes32) public didHashOf;

    event DidBound(address indexed account, bytes32 indexed didHash);

    constructor() Ownable(msg.sender) {}

    function bindDid(bytes32 didHash, bytes calldata signature) external {
        require(didHash != bytes32(0), "empty did");
        bytes32 digest = keccak256(
            abi.encodePacked("BindDID:", msg.sender, didHash)
        ).toEthSignedMessageHash();
        address signer = digest.recover(signature);
        require(signer == msg.sender, "invalid sig");
        didHashOf[msg.sender] = didHash;
        emit DidBound(msg.sender, didHash);
    }

    function resolveDid(address account) external view returns (bytes32) {
        return didHashOf[account];
    }
}
