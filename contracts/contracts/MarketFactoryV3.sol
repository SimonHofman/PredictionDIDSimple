// SPDX-License-Identifier: MIT
// 许可证标识：MIT开源协议
pragma solidity ^0.8.24;
// 指定Solidity编译器版本为0.8.24及以上

import "@openzeppelin/contracts/access/Ownable.sol"; // 导入所有权管理合约
import "@openzeppelin/contracts/token/ERC20/IERC20.sol"; // 导入ERC20代币接口
import "@openzeppelin/contracts/utils/Pausable.sol"; // 导入可暂停合约
import "./PredictionMarketV3.sol"; // 导入V3版预测市场合约
import "./MultiOutcomeMarket.sol"; // 导入多结果市场合约

/// @title MarketFactoryV3 — binary CPMM + multi-outcome markets
/// @title MarketFactoryV3 — 二元CPMM + 多结果市场工厂合约
contract MarketFactoryV3 is Ownable, Pausable {
    IERC20 public immutable collateral; // 抵押品代币合约地址（不可变）
    address public oracle; // 预言机适配器地址
    uint16 public defaultFeeBps; // 默认手续费基点数（1bps=0.01%）
    uint256 public defaultMaxBet; // 默认最大押注金额

    uint256 public marketCount; // 已创建的市场总数
    mapping(uint256 => address) public markets; // 市场ID到市场合约地址的映射
    mapping(uint256 => uint8) public marketTypes; // 0=binary v3, 1=multi
    // 市场类型映射（0=二元V3市场，1=多结果市场）

    // 二元市场创建事件
    event BinaryMarketCreated(uint256 indexed id, address market, bytes32 matchRef, string question);
    // 多结果市场创建事件
    event MultiMarketCreated(uint256 indexed id, address market, uint8 outcomes, string question);

    // 构造函数，初始化抵押品、预言机和默认手续费
    constructor(address _collateral, address _oracle, uint16 _feeBps) Ownable(msg.sender) {
        collateral = IERC20(_collateral); // 设置抵押品代币
        oracle = _oracle; // 设置预言机地址
        defaultFeeBps = _feeBps; // 设置默认手续费率
        defaultMaxBet = 10_000 * 1e6; // 默认最大押注为10000 USDC
    }

    function pause() external onlyOwner { _pause(); } // 暂停合约，仅所有者可调用
    function unpause() external onlyOwner { _unpause(); } // 恢复合约，仅所有者可调用

    function setOracle(address _oracle) external onlyOwner { oracle = _oracle; } // 设置预言机地址
    function setDefaultFeeBps(uint16 bps) external onlyOwner { defaultFeeBps = bps; } // 设置默认手续费率

    // 创建二元预测市场（带初始流动性），仅所有者可调用，合约未暂停时
    function createBinaryMarket(
        bytes32 matchRef, // 比赛引用哈希
        string calldata question, // 预测问题
        uint256 endTime, // 截止时间
        uint256 initialLiquidity // 初始流动性金额
    ) external onlyOwner whenNotPaused returns (address market, uint256 id) {
        if (initialLiquidity > 0) { // 如果提供了初始流动性
            collateral.transferFrom(msg.sender, address(this), initialLiquidity * 2); // 从调用者转入双倍流动性的抵押品
        }
        PredictionMarketV3 m = new PredictionMarketV3( // 部署新的V3预测市场合约
            address(collateral), // 抵押品地址
            oracle, // 预言机地址
            address(this), // 工厂合约地址
            matchRef, // 比赛引用
            question, // 预测问题
            endTime, // 截止时间
            defaultFeeBps, // 手续费率
            0, // 初始流动性设为0（稍后通过seedReserves注入）
            defaultMaxBet // 最大押注金额
        );
        if (initialLiquidity > 0) { // 如果有初始流动性
            collateral.transfer(address(m), initialLiquidity * 2); // 将抵押品转入市场合约
            m.seedReserves(initialLiquidity, msg.sender); // 初始化市场储备金
        }
        id = ++marketCount; // 市场计数器加一
        market = address(m); // 获取市场合约地址
        markets[id] = market; // 存储市场ID与地址的映射
        marketTypes[id] = 0; // 标记为二元市场类型
        emit BinaryMarketCreated(id, market, matchRef, question); // 触发二元市场创建事件
    }

    // 创建多结果预测市场，仅所有者可调用，合约未暂停时
    function createMultiMarket(
        bytes32 matchRef, // 比赛引用哈希
        string calldata question, // 预测问题
        uint256 endTime, // 截止时间
        uint8 outcomeCount // 结果数量（2-8个）
    ) external onlyOwner whenNotPaused returns (address market, uint256 id) {
        MultiOutcomeMarket m = new MultiOutcomeMarket( // 部署新的多结果市场合约
            address(collateral), // 抵押品地址
            oracle, // 预言机地址
            matchRef, // 比赛引用
            question, // 预测问题
            endTime, // 截止时间
            outcomeCount, // 结果数量
            defaultFeeBps // 手续费率
        );
        id = ++marketCount; // 市场计数器加一
        market = address(m); // 获取市场合约地址
        markets[id] = market; // 存储市场ID与地址的映射
        marketTypes[id] = 1; // 标记为多结果市场类型
        emit MultiMarketCreated(id, market, outcomeCount, question); // 触发多结果市场创建事件
    }

    // 返回合约版本号
    function version() external pure returns (string memory) {
        return "3.0.0-phase3"; // 版本3.0.0-第三阶段
    }
}
