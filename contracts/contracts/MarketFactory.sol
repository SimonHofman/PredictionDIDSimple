// SPDX-License-Identifier: MIT
// 许可证标识：MIT开源协议
pragma solidity ^0.8.24;
// 指定Solidity编译器版本为0.8.24及以上

import "@openzeppelin/contracts/access/Ownable.sol"; // 导入所有权管理合约
import "@openzeppelin/contracts/token/ERC20/IERC20.sol"; // 导入ERC20代币接口
import "./PredictionMarket.sol"; // 导入预测市场合约

/// @title MarketFactory — deploys Yes/No prediction markets
/// @title MarketFactory — 部署是/否二元预测市场的工厂合约
contract MarketFactory is Ownable {
    IERC20 public immutable collateral; // 抵押品代币合约地址（不可变）
    address public oracle; // 预言机适配器地址

    uint256 public marketCount; // 已创建的市场总数

    mapping(uint256 => address) public markets; // 市场ID到市场合约地址的映射

    // 市场创建事件
    event MarketCreated(
        uint256 indexed marketId, // 市场ID
        address indexed market, // 市场合约地址
        bytes32 indexed matchRef, // 比赛引用哈希
        string question, // 预测问题
        uint256 endTime // 截止时间
    );

    // 构造函数，初始化抵押品代币和预言机地址
    constructor(address _collateral, address _oracle) Ownable(msg.sender) {
        require(_collateral != address(0), "collateral"); // 抵押品地址不能为零
        require(_oracle != address(0), "oracle"); // 预言机地址不能为零
        collateral = IERC20(_collateral); // 设置抵押品代币合约
        oracle = _oracle; // 设置预言机地址
    }

    // 设置新的预言机地址，仅合约所有者可调用
    function setOracle(address _oracle) external onlyOwner {
        require(_oracle != address(0), "oracle"); // 预言机地址不能为零
        oracle = _oracle; // 更新预言机地址
    }

    // 创建新的预测市场，仅合约所有者可调用
    function createMarket(
        bytes32 matchRef, // 比赛引用哈希
        string calldata question, // 预测问题
        uint256 endTime // 截止时间
    ) external onlyOwner returns (address market, uint256 marketId) {
        PredictionMarket m = new PredictionMarket( // 部署新的预测市场合约
            address(collateral), // 抵押品代币地址
            oracle, // 预言机地址
            address(this), // 工厂合约地址
            matchRef, // 比赛引用
            question, // 预测问题
            endTime // 截止时间
        );
        marketId = ++marketCount; // 市场计数器加一
        market = address(m); // 获取新市场合约地址
        markets[marketId] = market; // 存储市场ID与地址的映射
        emit MarketCreated(marketId, market, matchRef, question, endTime); // 触发市场创建事件
    }

    // 返回合约版本号
    function version() external pure returns (string memory) {
        return "2.0.0-phase2"; // 版本2.0.0-第二阶段
    }
}
