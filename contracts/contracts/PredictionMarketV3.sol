// SPDX-License-Identifier: MIT
// 许可证标识：MIT开源协议
pragma solidity ^0.8.24;
// 指定Solidity编译器版本为0.8.24及以上

import "@openzeppelin/contracts/token/ERC20/IERC20.sol"; // 导入ERC20代币接口
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol"; // 导入安全ERC20转账库
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol"; // 导入重入攻击防护合约

/// @title PredictionMarketV3 — CPMM binary market with LP + fees
/// @title PredictionMarketV3 — 带流动性池和手续费的CPMM二元预测市场
contract PredictionMarketV3 is ReentrancyGuard {
    using SafeERC20 for IERC20; // 为IERC20启用安全转账方法

    // 市场状态枚举
    enum Status { Open, Resolved, Voided } // 开放、已结算、已作废

    IERC20 public immutable collateral; // 抵押品代币合约（不可变）
    address public immutable oracle; // 预言机地址（不可变）
    address public immutable factory; // 工厂合约地址（不可变）
    bytes32 public immutable matchRef; // 比赛引用哈希（不可变）
    string public question; // 预测问题
    uint256 public endTime; // 截止时间戳
    uint16 public feeBps; // 手续费基点数（1bps=0.01%）
    uint256 public maxBetPerUser; // 每个用户最大押注金额

    Status public marketStatus; // 当前市场状态
    uint8 public winningOutcome; // 获胜结果编号

    uint256 public reserveYes; // “是”方储备金（流动性池）
    uint256 public reserveNo; // “否”方储备金（流动性池）
    uint256 public totalLPSupply; // LP份额总供应量
    uint256 public collectedFees; // 已收取的手续费总额

    mapping(address => uint256) public yesBalance; // 用户持有的“是”方份额
    mapping(address => uint256) public noBalance; // 用户持有的“否”方份额
    mapping(address => uint256) public lpBalance; // 用户持有的LP份额
    mapping(address => uint256) public userBetTotal; // 用户累计押注总额
    mapping(address => bool) public claimed; // 用户是否已领取奖励

    event Bought(address indexed user, uint8 outcome, uint256 amountIn, uint256 sharesOut); // 购买事件
    event LiquidityAdded(address indexed user, uint256 lpMinted); // 添加流动性事件
    event LiquidityRemoved(address indexed user, uint256 lpBurned); // 移除流动性事件
    event Resolved(uint8 winningOutcome); // 结算事件
    event Claimed(address indexed user, uint256 amount); // 领取奖励事件
    event MarketVoided(); // 市场作废事件

    // 仅预言机可调用的修饰符
    modifier onlyOracle() {
        require(msg.sender == oracle, "not oracle"); // 验证调用者是预言机
        _;
    }

    // 构造函数，初始化CPMM二元预测市场
    constructor(
        address _collateral, // 抵押品代币地址
        address _oracle, // 预言机地址
        address _factory, // 工厂合约地址
        bytes32 _matchRef, // 比赛引用哈希
        string memory _question, // 预测问题
        uint256 _endTime, // 截止时间
        uint16 _feeBps, // 手续费基点数
        uint256 _initialLiquidity, // 初始流动性
        uint256 _maxBetPerUser // 每用户最大押注
    ) {
        require(_collateral != address(0) && _oracle != address(0), "zero addr"); // 地址不能为零
        collateral = IERC20(_collateral); // 设置抵押品代币
        oracle = _oracle; // 设置预言机地址
        factory = _factory; // 设置工厂合约地址
        matchRef = _matchRef; // 设置比赛引用
        question = _question; // 设置预测问题
        endTime = _endTime; // 设置截止时间
        feeBps = _feeBps; // 设置手续费率
        maxBetPerUser = _maxBetPerUser; // 设置用户最大押注限额
        marketStatus = Status.Open; // 初始状态为开放
        if (_initialLiquidity > 0) { // 如果提供了初始流动性
            _seedReserves(_initialLiquidity, msg.sender); // 初始化储备金
        }
    }

    /// @dev Factory seeds pool after transferring collateral to this contract.
    /// @dev 工厂合约将抵押品转入后初始化流动性池
    function seedReserves(uint256 perSide, address lpRecipient) external {
        require(msg.sender == factory, "not factory"); // 仅工厂合约可调用
        require(totalLPSupply == 0, "already seeded"); // LP池必须未初始化
        _seedReserves(perSide, lpRecipient); // 调用内部初始化方法
    }

    // 内部函数：初始化流动性储备金
    function _seedReserves(uint256 perSide, address lpRecipient) internal {
        uint256 total = perSide * 2; // 总量为每方的两倍
        require(collateral.balanceOf(address(this)) >= total, "insufficient seed"); // 合约余额必须充足
        reserveYes = perSide; // 设置“是”方储备金
        reserveNo = perSide; // 设置“否”方储备金
        lpBalance[lpRecipient] = total; // 分配LP份额给提供者
        totalLPSupply = total; // 设置LP总供应量
        emit LiquidityAdded(lpRecipient, total); // 触发添加流动性事件
    }

    // 购买预测份额（CPMM模式），防重入攻击
    function buy(uint8 outcome, uint256 amountIn) external nonReentrant {
        require(marketStatus == Status.Open, "not open"); // 市场必须开放
        require(block.timestamp < endTime, "ended"); // 必须在截止时间之前
        require(outcome <= 1 && amountIn > 0, "invalid"); // 结果有效且金额大于0
        require(userBetTotal[msg.sender] + amountIn <= maxBetPerUser || maxBetPerUser == 0, "max bet"); // 检查最大押注限额

        uint256 fee = (amountIn * feeBps) / 10_000; // 计算手续费
        uint256 net = amountIn - fee; // 扣除手续费后的净额
        collectedFees += fee; // 累加已收取的手续费
        collateral.safeTransferFrom(msg.sender, address(this), amountIn); // 从用户转入抵押品

        uint256 sharesOut = _swap(outcome, net); // 执行CPMM交换，获取份额
        userBetTotal[msg.sender] += amountIn; // 累加用户押注总额
        if (outcome == 0) { // 如果押注“是”
            yesBalance[msg.sender] += sharesOut; // 增加用户“是”方份额
        } else { // 如果押注“否”
            noBalance[msg.sender] += sharesOut; // 增加用户“否”方份额
        }
        emit Bought(msg.sender, outcome, amountIn, sharesOut); // 触发购买事件
    }

    // 添加流动性，防重入攻击
    function addLiquidity(uint256 amount) external nonReentrant {
        require(marketStatus == Status.Open, "not open"); // 市场必须开放
        require(amount > 0, "zero"); // 金额必须大于0
        collateral.safeTransferFrom(msg.sender, address(this), amount); // 从用户转入抵押品
        uint256 half = amount / 2; // 平分为两半
        reserveYes += half; // 加入“是”方储备
        reserveNo += amount - half; // 加入“否”方储备（处理奇数）
        lpBalance[msg.sender] += amount; // 增加用户LP份额
        totalLPSupply += amount; // 增加LP总供应量
        emit LiquidityAdded(msg.sender, amount); // 触发添加流动性事件
    }

    // 移除流动性，防重入攻击
    function removeLiquidity(uint256 lpAmount) external nonReentrant {
        require(lpAmount > 0 && lpBalance[msg.sender] >= lpAmount, "lp"); // LP份额必须有效
        uint256 yesOut = (lpAmount * reserveYes) / totalLPSupply; // 计算可取回的“是”方储备
        uint256 noOut = (lpAmount * reserveNo) / totalLPSupply; // 计算可取回的“否”方储备
        reserveYes -= yesOut; // 减少“是”方储备
        reserveNo -= noOut; // 减少“否”方储备
        lpBalance[msg.sender] -= lpAmount; // 减少用户LP份额
        totalLPSupply -= lpAmount; // 减少LP总供应量
        collateral.safeTransfer(msg.sender, yesOut + noOut); // 转账抵押品给用户
        emit LiquidityRemoved(msg.sender, lpAmount); // 触发移除流动性事件
    }

    // 结算市场，仅预言机可调用
    function resolve(uint8 _outcome) external onlyOracle {
        require(marketStatus == Status.Open, "not open"); // 市场必须开放
        require(_outcome <= 1, "invalid"); // 结果必须是0或1
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
            payout = _claimResolved(msg.sender); // 计算结算后奖励
        } else if (marketStatus == Status.Voided) { // 如果市场已作废
            payout = yesBalance[msg.sender] + noBalance[msg.sender]; // 退回用户全部份额
        } else { // 市场状态不允许领取
            revert("not claimable"); // 回退交易
        }
        require(payout > 0, "nothing"); // 应付金额必须大于0
        claimed[msg.sender] = true; // 标记为已领取
        collateral.safeTransfer(msg.sender, payout); // 转账奖励给用户
        emit Claimed(msg.sender, payout); // 触发领取事件
    }

    // 获取市场状态（返回数字编码）
    function status() external view returns (uint8) {
        return uint8(marketStatus); // 返回市场状态的数字表示
    }

    // 获取池子状态：“是”方储备、“否”方储备、“是”方价格基点数
    function getPoolState() external view returns (uint256 yesR, uint256 noR, uint256 priceYesBps) {
        yesR = reserveYes; // “是”方储备金
        noR = reserveNo; // “否”方储备金
        uint256 total = reserveYes + reserveNo; // 总储备金
        priceYesBps = total == 0 ? 5000 : (reserveNo * 10_000) / total; // 计算“是”方价格（基点数表示）
    }

    // 内部函数：CPMM交换逻辑，根据恒定乘积公式计算份额
    function _swap(uint8 outcome, uint256 net) internal returns (uint256 sharesOut) {
        if (outcome == 0) { // 购买“是”方份额
            sharesOut = (net * reserveYes) / (reserveNo + net); // 根据CPMM公式计算份额
            reserveNo += net; // 增加“否”方储备（注入资金）
            reserveYes -= sharesOut; // 减少“是”方储备（取出份额）
        } else { // 购买“否”方份额
            sharesOut = (net * reserveNo) / (reserveYes + net); // 根据CPMM公式计算份额
            reserveYes += net; // 增加“是”方储备（注入资金）
            reserveNo -= sharesOut; // 减少“否”方储备（取出份额）
        }
    }

    // 内部函数：计算结算后用户的奖励金额
    function _claimResolved(address user) internal view returns (uint256) {
        uint256 total = reserveYes + reserveNo; // 总储备金
        if (winningOutcome == 0) { // 如果“是”方获胜
            if (reserveYes == 0) return 0; // “是”方储备为空则返回0
            return (yesBalance[user] * total) / reserveYes; // 按用户份额占比分配总储备
        }
        if (reserveNo == 0) return 0; // “否”方储备为空则返回0
        return (noBalance[user] * total) / reserveNo; // 按用户份额占比分配总储备
    }
}
