// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

/// @title MultiOutcomeMarket — N-outcome parimutuel (2-8 outcomes)
contract MultiOutcomeMarket is ReentrancyGuard {
    using SafeERC20 for IERC20;

    enum Status { Open, Resolved, Voided }

    IERC20 public immutable collateral;
    address public immutable oracle;
    bytes32 public immutable matchRef;
    string public question;
    uint256 public endTime;
    uint8 public outcomeCount;
    uint16 public feeBps;

    Status public marketStatus;
    uint8 public winningOutcome;

    uint256[] public pool;
    mapping(address => mapping(uint8 => uint256)) public stake;
    mapping(address => bool) public claimed;

    event Bought(address indexed user, uint8 outcome, uint256 amount);
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
        bytes32 _matchRef,
        string memory _question,
        uint256 _endTime,
        uint8 _outcomeCount,
        uint16 _feeBps
    ) {
        require(_outcomeCount >= 2 && _outcomeCount <= 8, "outcomes");
        collateral = IERC20(_collateral);
        oracle = _oracle;
        matchRef = _matchRef;
        question = _question;
        endTime = _endTime;
        outcomeCount = _outcomeCount;
        feeBps = _feeBps;
        for (uint8 i = 0; i < _outcomeCount; i++) {
            pool.push(0);
        }
    }

    function status() external view returns (uint8) {
        return uint8(marketStatus);
    }

    function buy(uint8 outcome, uint256 amount) external nonReentrant {
        require(marketStatus == Status.Open && block.timestamp < endTime, "closed");
        require(outcome < outcomeCount && amount > 0, "invalid");
        uint256 fee = (amount * feeBps) / 10_000;
        uint256 net = amount - fee;
        collateral.safeTransferFrom(msg.sender, address(this), amount);
        pool[outcome] += net;
        stake[msg.sender][outcome] += net;
        emit Bought(msg.sender, outcome, amount);
    }

    function resolve(uint8 _outcome) external onlyOracle {
        require(marketStatus == Status.Open && _outcome < outcomeCount, "invalid");
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
            uint256 winPool = pool[winningOutcome];
            require(winPool > 0, "empty");
            uint256 total;
            for (uint8 i = 0; i < outcomeCount; i++) {
                total += pool[i];
            }
            payout = (stake[msg.sender][winningOutcome] * total) / winPool;
        } else if (marketStatus == Status.Voided) {
            for (uint8 i = 0; i < outcomeCount; i++) {
                payout += stake[msg.sender][i];
            }
        } else {
            revert("not claimable");
        }
        require(payout > 0, "nothing");
        claimed[msg.sender] = true;
        collateral.safeTransfer(msg.sender, payout);
        emit Claimed(msg.sender, payout);
    }
}
