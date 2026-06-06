// SPDX-License-Identifier: MIT
// 许可证标识：MIT开源协议
pragma solidity ^0.8.24;
// 指定Solidity编译器版本为0.8.24及以上

import "@openzeppelin/contracts/token/ERC20/IERC20.sol"; // 导入ERC20代币接口
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol"; // 导入安全ERC20转账库

/// @title PredictionMarket — parimutuel Yes(0) / No(1) market
/// @title PredictionMarket — 互池式是(0)/否(1)二元预测市场
contract PredictionMarket {
    using SafeERC20 for IERC20; // 为IERC20启用安全转账方法

    // 市场状态枚举
    enum Status {
        Open, // 开放中
        Resolved, // 已结算
        Voided // 已作废
    }

    IERC20 public immutable collateral; // 抵押品代币合约（不可变）
    address public immutable oracle; // 预言机地址（不可变）
    address public immutable factory; // 工厂合约地址（不可变）
    bytes32 public immutable matchRef; // 比赛引用哈希（不可变）
    string public question; // 预测问题
    uint256 public endTime; // 截止时间戳

    Status public status; // 当前市场状态
    uint8 public winningOutcome; // 获胜结果编号

    uint256 public yesPool; // “是”方资金池
    uint256 public noPool; // “否”方资金池

    mapping(address => uint256) public yesBalance; // 用户在“是”方的押注余额
    mapping(address => uint256) public noBalance; // 用户在“否”方的押注余额
    mapping(address => bool) public claimed; // 用户是否已领取奖励

    event Bought(address indexed user, uint8 outcome, uint256 amount); // 购买事件
    event Resolved(uint8 winningOutcome); // 结算事件
    event Claimed(address indexed user, uint256 amount); // 领取奖励事件
    event MarketVoided(); // 市场作废事件

    // 仅预言机可调用的修饰符
    modifier onlyOracle() {
        require(msg.sender == oracle, "not oracle"); // 验证调用者是预言机
        _;
    }

    // 构造函数，初始化预测市场
    constructor(
        address _collateral, // 抵押品代币地址
        address _oracle, // 预言机地址
        address _factory, // 工厂合约地址
        bytes32 _matchRef, // 比赛引用哈希
        string memory _question, // 预测问题
        uint256 _endTime // 截止时间
    ) {
        require(_collateral != address(0), "collateral"); // 抵押品地址不能为零
        require(_oracle != address(0), "oracle"); // 预言机地址不能为零
        require(_endTime > block.timestamp, "end in past"); // 截止时间必须在未来
        collateral = IERC20(_collateral); // 设置抵押品代币
        oracle = _oracle; // 设置预言机地址
        factory = _factory; // 设置工厂合约地址
        matchRef = _matchRef; // 设置比赛引用
        question = _question; // 设置预测问题
        endTime = _endTime; // 设置截止时间
        status = Status.Open; // 初始状态为开放
    }

    // 购买预测份额（outcome: 0=是, 1=否）
    function buy(uint8 outcome, uint256 amount) external {
        require(status == Status.Open, "not open"); // 市场必须开放
        require(block.timestamp < endTime, "ended"); // 必须在截止时间之前
        require(outcome <= 1, "invalid outcome"); // 结果必须是0或1
        require(amount > 0, "zero amount"); // 金额必须大于0

        collateral.safeTransferFrom(msg.sender, address(this), amount); // 从用户转入抵押品

        if (outcome == 0) { // 如果押注“是”
            yesBalance[msg.sender] += amount; // 增加用户“是”方余额
            yesPool += amount; // 增加“是”方资金池
        } else { // 如果押注“否”
            noBalance[msg.sender] += amount; // 增加用户“否”方余额
            noPool += amount; // 增加“否”方资金池
        }

        emit Bought(msg.sender, outcome, amount); // 触发购买事件
    }

    // 结算市场，仅预言机可调用
    function resolve(uint8 _winningOutcome) external onlyOracle {
        require(status == Status.Open, "not open"); // 市场必须开放
        require(_winningOutcome <= 1, "invalid outcome"); // 结果必须是0或1
        status = Status.Resolved; // 设置市场状态为已结算
        winningOutcome = _winningOutcome; // 记录获胜结果
        emit Resolved(_winningOutcome); // 触发结算事件
    }

    // 作废市场，仅预言机可调用
    function voidMarket() external onlyOracle {
        require(status == Status.Open, "not open"); // 市场必须开放
        status = Status.Voided; // 设置市场状态为已作废
        emit MarketVoided(); // 触发作废事件
    }

    // 领取奖励
    function claim() external {
        require(!claimed[msg.sender], "already claimed"); // 不能重复领取
        uint256 payout; // 应付金额

        if (status == Status.Resolved) { // 如果市场已结算
            payout = _payoutResolved(msg.sender); // 计算结算后的奖励
        } else if (status == Status.Voided) { // 如果市场已作废
            payout = yesBalance[msg.sender] + noBalance[msg.sender]; // 退回用户全部押注
        } else { // 市场状态不允许领取
            revert("not claimable"); // 回退交易
        }

        require(payout > 0, "nothing to claim"); // 应付金额必须大于0
        claimed[msg.sender] = true; // 标记为已领取
        collateral.safeTransfer(msg.sender, payout); // 转账奖励给用户
        emit Claimed(msg.sender, payout); // 触发领取事件
    }

    // 内部函数：计算结算后用户的奖励金额
    function _payoutResolved(address user) internal view returns (uint256) {
        uint256 userStake; // 用户在获胜方的押注
        uint256 winSideTotal; // 获胜方总资金池
        uint256 totalPool = yesPool + noPool; // 两方总资金池

        if (winningOutcome == 0) { // 如果“是”方获胜
            userStake = yesBalance[user]; // 用户在“是”方的押注
            winSideTotal = yesPool; // “是”方总资金
        } else { // 如果“否”方获胜
            userStake = noBalance[user]; // 用户在“否”方的押注
            winSideTotal = noPool; // “否”方总资金
        }

        if (userStake == 0 || winSideTotal == 0) { // 如果用户未押注或获胜方无资金
            return 0; // 返回0
        }
        return (userStake * totalPool) / winSideTotal; // 按用户押注占比分配总资金池
    }
}
