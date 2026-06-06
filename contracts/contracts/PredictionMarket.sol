// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

/// @title PredictionMarket — parimutuel Yes(0) / No(1) market
contract PredictionMarket {
    using SafeERC20 for IERC20;

    enum Status {
        Open,
        Resolved,
        Voided
    }

    IERC20 public immutable collateral;
    address public immutable oracle;
    address public immutable factory;
    bytes32 public immutable matchRef;
    string public question;
    uint256 public endTime;

    Status public status;
    uint8 public winningOutcome;

    uint256 public yesPool;
    uint256 public noPool;

    mapping(address => uint256) public yesBalance;
    mapping(address => uint256) public noBalance;
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
        address _factory,
        bytes32 _matchRef,
        string memory _question,
        uint256 _endTime
    ) {
        require(_collateral != address(0), "collateral");
        require(_oracle != address(0), "oracle");
        require(_endTime > block.timestamp, "end in past");
        collateral = IERC20(_collateral);
        oracle = _oracle;
        factory = _factory;
        matchRef = _matchRef;
        question = _question;
        endTime = _endTime;
        status = Status.Open;
    }

    function buy(uint8 outcome, uint256 amount) external {
        require(status == Status.Open, "not open");
        require(block.timestamp < endTime, "ended");
        require(outcome <= 1, "invalid outcome");
        require(amount > 0, "zero amount");

        collateral.safeTransferFrom(msg.sender, address(this), amount);

        if (outcome == 0) {
            yesBalance[msg.sender] += amount;
            yesPool += amount;
        } else {
            noBalance[msg.sender] += amount;
            noPool += amount;
        }

        emit Bought(msg.sender, outcome, amount);
    }

    function resolve(uint8 _winningOutcome) external onlyOracle {
        require(status == Status.Open, "not open");
        require(_winningOutcome <= 1, "invalid outcome");
        status = Status.Resolved;
        winningOutcome = _winningOutcome;
        emit Resolved(_winningOutcome);
    }

    function voidMarket() external onlyOracle {
        require(status == Status.Open, "not open");
        status = Status.Voided;
        emit MarketVoided();
    }

    function claim() external {
        require(!claimed[msg.sender], "already claimed");
        uint256 payout;

        if (status == Status.Resolved) {
            payout = _payoutResolved(msg.sender);
        } else if (status == Status.Voided) {
            payout = yesBalance[msg.sender] + noBalance[msg.sender];
        } else {
            revert("not claimable");
        }

        require(payout > 0, "nothing to claim");
        claimed[msg.sender] = true;
        collateral.safeTransfer(msg.sender, payout);
        emit Claimed(msg.sender, payout);
    }

    function _payoutResolved(address user) internal view returns (uint256) {
        uint256 userStake;
        uint256 winSideTotal;
        uint256 totalPool = yesPool + noPool;

        if (winningOutcome == 0) {
            userStake = yesBalance[user];
            winSideTotal = yesPool;
        } else {
            userStake = noBalance[user];
            winSideTotal = noPool;
        }

        if (userStake == 0 || winSideTotal == 0) {
            return 0;
        }
        return (userStake * totalPool) / winSideTotal;
    }
}
