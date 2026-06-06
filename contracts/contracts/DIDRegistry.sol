// SPDX-License-Identifier: MIT
// 许可证标识：MIT开源协议
pragma solidity ^0.8.24;
// 指定Solidity编译器版本为0.8.24及以上

import "@openzeppelin/contracts/access/Ownable.sol"; // 导入所有权管理合约
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol"; // 导入ECDSA签名验证库
import "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol"; // 导入消息哈希工具库

/// @title DIDRegistry — optional on-chain DID hash binding
/// @title DIDRegistry — 可选的链上DID哈希绑定合约
contract DIDRegistry is Ownable {
    using ECDSA for bytes32; // 为bytes32类型启用ECDSA库方法
    using MessageHashUtils for bytes32; // 为bytes32类型启用消息哈希工具方法

    mapping(address => bytes32) public didHashOf; // 地址到DID哈希的映射关系

    event DidBound(address indexed account, bytes32 indexed didHash); // DID绑定事件

    constructor() Ownable(msg.sender) {} // 构造函数，将部署者设为合约所有者

    // 绑定DID哈希到调用者地址，需要签名验证
    function bindDid(bytes32 didHash, bytes calldata signature) external {
        require(didHash != bytes32(0), "empty did"); // 要求DID哈希不能为空
        bytes32 digest = keccak256( // 构建签名消息摘要
            abi.encodePacked("BindDID:", msg.sender, didHash) // 将前缀、发送者地址和DID哈希编码打包
        ).toEthSignedMessageHash(); // 转换为以太坊签名消息哈希格式
        address signer = digest.recover(signature); // 从签名中恢复签名者地址
        require(signer == msg.sender, "invalid sig"); // 验证签名者必须是调用者本人
        didHashOf[msg.sender] = didHash; // 存储地址与DID哈希的绑定关系
        emit DidBound(msg.sender, didHash); // 触发DID绑定事件
    }

    // 解析某个地址绑定的DID哈希
    function resolveDid(address account) external view returns (bytes32) {
        return didHashOf[account]; // 返回该地址对应的DID哈希值
    }
}
