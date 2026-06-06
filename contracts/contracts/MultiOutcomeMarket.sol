// SPDX-License-Identifier: MIT
// 许可证标识：MIT开源协议
pragma solidity ^0.8.24;
// 指定Solidity编译器版本为0.8.24及以上

import "@openzeppelin/contracts/token/ERC20/IERC20.sol"; // 导入ERC20代币接口
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol"; // 导入安全ERC20转账库
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol"; // 导入重入攻击防护合约

/// @title MultiOutcomeMarket — N-outcome parimutuel (2-8 outcomes)
/// @title MultiOutcomeMarket — N结果互池式押注市场（2-8个结果）
contract MultiOutcomeMarket is ReentrancyGuard {
    using SafeERC20 for IERC20; // 为IERC20启用安全转账方法

    // 市场状态枚举
    enum Status { Open, Resolved, Voided } // 开放、已结算、已作废

    IERC20 public immutable collateral; // 抵押品代币合约（不可变）
    address public immutable oracle; // 预言机地址（不可变）
    bytes32 public immutable matchRef; // 比赛引用哈希（不可变）
    string public question; // 预测问题
    uint256 public endTime; // 截止时间戳
    uint8 public outcomeCount; // 结果数量
    uint16 public feeBps; // 手续费基点数

    Status public marketStatus; // 当前市场状态
    uint8 public winningOutcome; // 获胜结果编号

    uint256[] public pool; // 各结果的资金池数组
    mapping(address => mapping(uint8 => uint256)) public stake; // 用户在各结果上的押注金额
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

    // 构造函数，初始化多结果市场
    constructor(
        address _collateral, // 抵押品代币地址
        address _oracle, // 预言机地址
        bytes32 _matchRef, // 比赛引用哈希
        string memory _question, // 预测问题
        uint256 _endTime, // 截止时间
        uint8 _outcomeCount, // 结果数量
        uint16 _feeBps // 手续费基点数
    ) {
        require(_outcomeCount >= 2 && _outcomeCount <= 8, "outcomes"); // 结果数量必须在2-8之间
        collateral = IERC20(_collateral); // 设置抵押品代币
        oracle = _oracle; // 设置预言机地址
        matchRef = _matchRef; // 设置比赛引用
        question = _question; // 设置预测问题
        endTime = _endTime; // 设置截止时间
        outcomeCount = _outcomeCount; // 设置结果数量
        feeBps = _feeBps; // 设置手续费率
        for (uint8 i = 0; i < _outcomeCount; i++) { // 初始化各结果的资金池为0
            pool.push(0);
        }
    }

    // 获取市场状态（返回数字编码）
    function status() external view returns (uint8) {
        return uint8(marketStatus); // 返回市场状态的数字表示
    }

    // 购买某个结果的份额，防重入攻击
    function buy(uint8 outcome, uint256 amount) external nonReentrant {
        require(marketStatus == Status.Open && block.timestamp < endTime, "closed"); // 市场必须开放且未过期
        require(outcome < outcomeCount && amount > 0, "invalid"); // 结果编号有效且金额大于0
        uint256 fee = (amount * feeBps) / 10_000; // 计算手续费
        uint256 net = amount - fee; // 扣除手续费后的净额
        collateral.safeTransferFrom(msg.sender, address(this), amount); // 从用户转入抵押品
        pool[outcome] += net; // 将净额加入对应结果的资金池
        stake[msg.sender][outcome] += net; // 记录用户在该结果上的押注
        emit Bought(msg.sender, outcome, amount); // 触发购买事件
    }

    // 结算市场，仅预言机可调用
    function resolve(uint8 _outcome) external onlyOracle {
        require(marketStatus == Status.Open && _outcome < outcomeCount, "invalid"); // 市场必须开放且结果有效
        marketStatus = Status.Resolved; // 设置市场状态为已结算
        winningOutcome = _outcome; // 记录获胜结果
        emit Resolved(_outcome); // 触发结算事件
    }

    // 作废市场，仅预言机可调用
    function voidMarket() external onlyOracle {
        require(marketStatus == Status.Open, "not open"); // 市场必须开放
        marketStatus = Status.Voided; // 设置市场状态为已作废
        emit MarketVoided(); // 触发作废事件
    }

    // 领取奖励，防重入攻击
    function claim() external nonReentrant {
        require(!claimed[msg.sender], "claimed"); // 不能重复领取
        uint256 payout; // 应付金额
        if (marketStatus == Status.Resolved) { // 如果市场已结算
            uint256 winPool = pool[winningOutcome]; // 获取获胜结果的资金池
            require(winPool > 0, "empty"); // 获胜池不能为空
            uint256 total; // 总资金池
            for (uint8 i = 0; i < outcomeCount; i++) { // 累加所有结果的资金池
                total += pool[i];
            }
            payout = (stake[msg.sender][winningOutcome] * total) / winPool; // 按比例计算奖励
        } else if (marketStatus == Status.Voided) { // 如果市场已作废
            for (uint8 i = 0; i < outcomeCount; i++) { // 累加用户所有押注
                payout += stake[msg.sender][i]; // 退回用户在各结果上的押注
            }
        } else { // 市场状态不允许领取
            revert("not claimable"); // 回退交易
        }
        require(payout > 0, "nothing"); // 应付金额必须大于0
        claimed[msg.sender] = true; // 标记为已领取
        collateral.safeTransfer(msg.sender, payout); // 转账奖励给用户
        emit Claimed(msg.sender, payout); // 触发领取事件
    }
}
