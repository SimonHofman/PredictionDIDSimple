// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

/// @title PredictionMarketV3 — CPMM binary market with LP + fees
contract PredictionMarketV3 is ReentrancyGuard {
    using SafeERC20 for IERC20;

    enum Status { Open, Resolved, Voided }

    IERC20 public immutable collateral;
    address public immutable oracle;
    address public immutable factory;
    bytes32 public immutable matchRef;
    string public question;
    uint256 public endTime;
    uint16 public feeBps;
    uint256 public maxBetPerUser;

    Status public marketStatus;
    uint8 public winningOutcome;

    uint256 public reserveYes;
    uint256 public reserveNo;
    uint256 public totalLPSupply;
    uint256 public collectedFees;

    mapping(address => uint256) public yesBalance;
    mapping(address => uint256) public noBalance;
    mapping(address => uint256) public lpBalance;
    mapping(address => uint256) public userBetTotal;
    mapping(address => bool) public claimed;

    event Bought(address indexed user, uint8 outcome, uint256 amountIn, uint256 sharesOut);
    event LiquidityAdded(address indexed user, uint256 lpMinted);
    event LiquidityRemoved(address indexed user, uint256 lpBurned);
    event Resolved(uint8 winningOutcome);
    event Claimed(address indexed user, uint256 amount);
    event MarketVoided();

    modifier onlyOracle() {
        require(msg.sender == oracle, "not oracle");
        _;
    }

    constructor(
        address _collateral,
        address _oracle,
        address _factory,
        bytes32 _matchRef,
        string memory _question,
        uint256 _endTime,
        uint16 _feeBps,
        uint256 _initialLiquidity,
        uint256 _maxBetPerUser
    ) {
        require(_collateral != address(0) && _oracle != address(0), "zero addr");
        collateral = IERC20(_collateral);
        oracle = _oracle;
        factory = _factory;
        matchRef = _matchRef;
        question = _question;
        endTime = _endTime;
        feeBps = _feeBps;
        maxBetPerUser = _maxBetPerUser;
        marketStatus = Status.Open;
        if (_initialLiquidity > 0) {
            _seedReserves(_initialLiquidity, msg.sender);
        }
    }

    /// @dev Factory seeds pool after transferring collateral to this contract.
    function seedReserves(uint256 perSide, address lpRecipient) external {
        require(msg.sender == factory, "not factory");
        require(totalLPSupply == 0, "already seeded");
        _seedReserves(perSide, lpRecipient);
    }

    function _seedReserves(uint256 perSide, address lpRecipient) internal {
        uint256 total = perSide * 2;
        require(collateral.balanceOf(address(this)) >= total, "insufficient seed");
        reserveYes = perSide;
        reserveNo = perSide;
        lpBalance[lpRecipient] = total;
        totalLPSupply = total;
        emit LiquidityAdded(lpRecipient, total);
    }

    function buy(uint8 outcome, uint256 amountIn) external nonReentrant {
        require(marketStatus == Status.Open, "not open");
        require(block.timestamp < endTime, "ended");
        require(outcome <= 1 && amountIn > 0, "invalid");
        require(userBetTotal[msg.sender] + amountIn <= maxBetPerUser || maxBetPerUser == 0, "max bet");

        uint256 fee = (amountIn * feeBps) / 10_000;
        uint256 net = amountIn - fee;
        collectedFees += fee;
        collateral.safeTransferFrom(msg.sender, address(this), amountIn);

        uint256 sharesOut = _swap(outcome, net);
        userBetTotal[msg.sender] += amountIn;
        if (outcome == 0) {
            yesBalance[msg.sender] += sharesOut;
        } else {
            noBalance[msg.sender] += sharesOut;
        }
        emit Bought(msg.sender, outcome, amountIn, sharesOut);
    }

    function addLiquidity(uint256 amount) external nonReentrant {
        require(marketStatus == Status.Open, "not open");
        require(amount > 0, "zero");
        collateral.safeTransferFrom(msg.sender, address(this), amount);
        uint256 half = amount / 2;
        reserveYes += half;
        reserveNo += amount - half;
        lpBalance[msg.sender] += amount;
        totalLPSupply += amount;
        emit LiquidityAdded(msg.sender, amount);
    }

    function removeLiquidity(uint256 lpAmount) external nonReentrant {
        require(lpAmount > 0 && lpBalance[msg.sender] >= lpAmount, "lp");
        uint256 yesOut = (lpAmount * reserveYes) / totalLPSupply;
        uint256 noOut = (lpAmount * reserveNo) / totalLPSupply;
        reserveYes -= yesOut;
        reserveNo -= noOut;
        lpBalance[msg.sender] -= lpAmount;
        totalLPSupply -= lpAmount;
        collateral.safeTransfer(msg.sender, yesOut + noOut);
        emit LiquidityRemoved(msg.sender, lpAmount);
    }

    function resolve(uint8 _outcome) external onlyOracle {
        require(marketStatus == Status.Open, "not open");
        require(_outcome <= 1, "invalid");
        marketStatus = Status.Resolved;
        winningOutcome = _outcome;
        emit Resolved(_outcome);
    }

    function voidMarket() external onlyOracle {
        require(marketStatus == Status.Open, "not open");
        marketStatus = Status.Voided;
        emit MarketVoided();
    }

    function claim() external nonReentrant {
        require(!claimed[msg.sender], "claimed");
        uint256 payout;
        if (marketStatus == Status.Resolved) {
            payout = _claimResolved(msg.sender);
        } else if (marketStatus == Status.Voided) {
            payout = yesBalance[msg.sender] + noBalance[msg.sender];
        } else {
            revert("not claimable");
        }
        require(payout > 0, "nothing");
        claimed[msg.sender] = true;
        collateral.safeTransfer(msg.sender, payout);
        emit Claimed(msg.sender, payout);
    }

    function status() external view returns (uint8) {
        return uint8(marketStatus);
    }

    function getPoolState() external view returns (uint256 yesR, uint256 noR, uint256 priceYesBps) {
        yesR = reserveYes;
        noR = reserveNo;
        uint256 total = reserveYes + reserveNo;
        priceYesBps = total == 0 ? 5000 : (reserveNo * 10_000) / total;
    }

    function _swap(uint8 outcome, uint256 net) internal returns (uint256 sharesOut) {
        if (outcome == 0) {
            sharesOut = (net * reserveYes) / (reserveNo + net);
            reserveNo += net;
            reserveYes -= sharesOut;
        } else {
            sharesOut = (net * reserveNo) / (reserveYes + net);
            reserveYes += net;
            reserveNo -= sharesOut;
        }
    }

    function _claimResolved(address user) internal view returns (uint256) {
        uint256 total = reserveYes + reserveNo;
        if (winningOutcome == 0) {
            if (reserveYes == 0) return 0;
            return (yesBalance[user] * total) / reserveYes;
        }
        if (reserveNo == 0) return 0;
        return (noBalance[user] * total) / reserveNo;
    }
}
