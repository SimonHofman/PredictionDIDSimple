// SPDX-License-Identifier: MIT
// 许可证标识：MIT开源协议
pragma solidity ^0.8.24;
// 指定Solidity编译器版本为0.8.24及以上

import "@openzeppelin/contracts/access/AccessControl.sol"; // 导入角色访问控制合约
import "./interfaces/IPredictionMarket.sol"; // 导入预测市场接口

/// @title OracleAdapterV2 — m-of-n multisig resolve + void
/// @title OracleAdapterV2 — m-of-n多签结算和作废适配器
contract OracleAdapterV2 is AccessControl {
    bytes32 public constant ORACLE_ROLE = keccak256("ORACLE_ROLE"); // 预言机角色标识符

    uint256 public threshold; // 多签批准阈值（需要多少个预言机批准）
    uint256 public proposalCount; // 提案总数计数器

    // 提案结构体
    struct Proposal {
        address market; // 目标市场地址
        uint8 outcome; // 提议的结果编号
        uint256 approvals; // 已获得的批准数
        bool executed; // 是否已执行
    }

    mapping(uint256 => Proposal) public proposals; // 提案ID到提案的映射
    mapping(uint256 => mapping(address => bool)) public approved; // 提案ID和地址到是否已批准的映射

    event ProposalCreated(uint256 indexed id, address indexed market, uint8 outcome); // 提案创建事件
    event ProposalApproved(uint256 indexed id, address indexed oracle); // 提案批准事件
    event ProposalExecuted(uint256 indexed id, uint8 outcome); // 提案执行事件
    event MarketVoided(address indexed market); // 市场作废事件

    // 构造函数，初始化管理员和多签阈值
    constructor(address admin, uint256 _threshold) {
        _grantRole(DEFAULT_ADMIN_ROLE, admin); // 授予管理员角色
        _grantRole(ORACLE_ROLE, admin); // 授予预言机角色
        threshold = _threshold; // 设置多签阈值
    }

    // 设置多签阈值，仅管理员可调用
    function setThreshold(uint256 t) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(t > 0, "threshold"); // 阈值必须大于0
        threshold = t; // 更新阈值
    }

    // 授予预言机角色，仅管理员可调用
    function grantOracle(address account) external onlyRole(DEFAULT_ADMIN_ROLE) {
        _grantRole(ORACLE_ROLE, account); // 授予指定地址预言机角色
    }

    // 提议结算市场，仅预言机角色可调用
    function proposeResolve(address market, uint8 outcome) external onlyRole(ORACLE_ROLE) returns (uint256 id) {
        require(outcome <= 7, "outcome"); // 结果编号必须在0-7之间
        require(IPredictionMarket(market).status() == 0, "not open"); // 市场必须开放
        id = ++proposalCount; // 提案计数器加一
        proposals[id] = Proposal({market: market, outcome: outcome, approvals: 0, executed: false}); // 创建新提案
        _approve(id); // 提案者自动批准
        emit ProposalCreated(id, market, outcome); // 触发提案创建事件
    }

    // 批准提案，仅预言机角色可调用
    function approveResolve(uint256 id) external onlyRole(ORACLE_ROLE) {
        _approve(id); // 调用内部批准方法
    }

    // 内部批准逻辑
    function _approve(uint256 id) internal {
        Proposal storage p = proposals[id]; // 获取提案引用
        require(!p.executed, "executed"); // 提案不能已执行
        require(!approved[id][msg.sender], "approved"); // 不能重复批准
        approved[id][msg.sender] = true; // 记录批准
        p.approvals++; // 批准数加一
        emit ProposalApproved(id, msg.sender); // 触发批准事件
        if (p.approvals >= threshold) { // 如果批准数达到阈值
            _execute(id); // 执行提案
        }
    }

    // 内部执行提案逻辑
    function _execute(uint256 id) internal {
        Proposal storage p = proposals[id]; // 获取提案引用
        require(!p.executed, "executed"); // 提案不能已执行
        p.executed = true; // 标记为已执行
        IPredictionMarket(p.market).resolve(p.outcome); // 调用市场合约结算
        emit ProposalExecuted(id, p.outcome); // 触发提案执行事件
    }

    // 作废市场，仅预言机角色可调用
    function voidMarket(address market) external onlyRole(ORACLE_ROLE) {
        require(IPredictionMarket(market).status() == 0, "not open"); // 市场必须开放
        IPredictionMarket(market).voidMarket(); // 调用市场合约作废
        emit MarketVoided(market); // 触发市场作废事件
    }
}
