// SPDX-License-Identifier: MIT
// 许可证标识：MIT开源协议
pragma solidity ^0.8.24;
// 指定Solidity编译器版本为0.8.24及以上

import "@openzeppelin/contracts/access/AccessControl.sol"; // 导入角色访问控制合约
import "./interfaces/IPredictionMarket.sol"; // 导入预测市场接口

/// @title OracleAdapter — timelocked resolve + void for prediction markets
/// @title OracleAdapter — 带时间锁的预测市场结算和作废适配器
contract OracleAdapter is AccessControl {
    bytes32 public constant ORACLE_ROLE = keccak256("ORACLE_ROLE"); // 预言机角色标识符

    uint256 public timelockDelay; // 时间锁延迟秒数
    address public factory; // 工厂合约地址

    // 待处理的结算请求结构体
    struct PendingResolve {
        uint8 outcome; // 结果编号
        uint256 executeAfter; // 可执行时间戳
        bool active; // 是否活跃
    }

    mapping(address => PendingResolve) public pending; // 市场地址到待处理结算的映射

    // 结算请求事件
    event OracleResolveRequested(
        address indexed market, // 市场地址
        uint8 outcome, // 结果编号
        uint256 executeAfter // 可执行时间
    );
    event OracleResolveConfirmed(address indexed market, uint8 outcome); // 结算确认事件
    event MarketVoided(address indexed market); // 市场作废事件

    // 构造函数，初始化管理员和时间锁延迟
    constructor(address admin, uint256 _timelockDelay) {
        _grantRole(DEFAULT_ADMIN_ROLE, admin); // 授予管理员角色
        _grantRole(ORACLE_ROLE, admin); // 授予预言机角色
        timelockDelay = _timelockDelay; // 设置时间锁延迟
    }

    // 设置时间锁延迟，仅管理员可调用
    function setTimelockDelay(uint256 delay) external onlyRole(DEFAULT_ADMIN_ROLE) {
        timelockDelay = delay; // 更新时间锁延迟值
    }

    // 设置工厂合约地址，仅管理员可调用
    function setFactory(address _factory) external onlyRole(DEFAULT_ADMIN_ROLE) {
        factory = _factory; // 更新工厂合约地址
    }

    // 授予预言机角色，仅管理员可调用
    function grantOracle(address account) external onlyRole(DEFAULT_ADMIN_ROLE) {
        _grantRole(ORACLE_ROLE, account); // 授予指定地址预言机角色
    }

    // 请求结算市场（启动时间锁），仅预言机角色可调用
    function requestResolve(address market, uint8 outcome) external onlyRole(ORACLE_ROLE) {
        require(outcome <= 1, "invalid outcome"); // 结果必须是0或1
        require(IPredictionMarket(market).status() == 0, "not open"); // 市场必须开放
        uint256 executeAfter = block.timestamp + timelockDelay; // 计算可执行时间
        pending[market] = PendingResolve({ // 存储待处理结算请求
            outcome: outcome, // 结果编号
            executeAfter: executeAfter, // 可执行时间
            active: true // 标记为活跃
        });
        emit OracleResolveRequested(market, outcome, executeAfter); // 触发结算请求事件
    }

    // 确认结算市场（时间锁过后执行），仅预言机角色可调用
    function confirmResolve(address market) external onlyRole(ORACLE_ROLE) {
        PendingResolve storage p = pending[market]; // 获取待处理结算信息
        require(p.active, "no pending"); // 必须存在待处理的请求
        require(block.timestamp >= p.executeAfter, "timelock"); // 时间锁必须已过期
        p.active = false; // 标记为已处理
        IPredictionMarket(market).resolve(p.outcome); // 调用市场合约结算
        emit OracleResolveConfirmed(market, p.outcome); // 触发结算确认事件
    }

    /// @notice Fast path when timelock is 0
    /// @notice 时间锁为0时的快速结算路径
    function resolveNow(address market, uint8 outcome) external onlyRole(ORACLE_ROLE) {
        require(timelockDelay == 0, "use request+confirm"); // 时间锁必须为0
        require(outcome <= 1, "invalid outcome"); // 结果必须是0或1
        IPredictionMarket(market).resolve(outcome); // 直接结算市场
        emit OracleResolveConfirmed(market, outcome); // 触发结算确认事件
    }

    // 作废市场，仅预言机角色可调用
    function voidMarket(address market) external onlyRole(ORACLE_ROLE) {
        require(IPredictionMarket(market).status() == 0, "not open"); // 市场必须开放
        IPredictionMarket(market).voidMarket(); // 调用市场合约作废
        emit MarketVoided(market); // 触发市场作废事件
    }
}
