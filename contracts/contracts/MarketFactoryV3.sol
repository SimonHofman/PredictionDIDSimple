// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/utils/Pausable.sol";
import "./PredictionMarketV3.sol";
import "./MultiOutcomeMarket.sol";

/// @title MarketFactoryV3 — binary CPMM + multi-outcome markets
contract MarketFactoryV3 is Ownable, Pausable {
    IERC20 public immutable collateral;
    address public oracle;
    uint16 public defaultFeeBps;
    uint256 public defaultMaxBet;

    uint256 public marketCount;
    mapping(uint256 => address) public markets;
    mapping(uint256 => uint8) public marketTypes; // 0=binary v3, 1=multi

    event BinaryMarketCreated(uint256 indexed id, address market, bytes32 matchRef, string question);
    event MultiMarketCreated(uint256 indexed id, address market, uint8 outcomes, string question);

    constructor(address _collateral, address _oracle, uint16 _feeBps) Ownable(msg.sender) {
        collateral = IERC20(_collateral);
        oracle = _oracle;
        defaultFeeBps = _feeBps;
        defaultMaxBet = 10_000 * 1e6;
    }

    function pause() external onlyOwner { _pause(); }
    function unpause() external onlyOwner { _unpause(); }

    function setOracle(address _oracle) external onlyOwner { oracle = _oracle; }
    function setDefaultFeeBps(uint16 bps) external onlyOwner { defaultFeeBps = bps; }

    function createBinaryMarket(
        bytes32 matchRef,
        string calldata question,
        uint256 endTime,
        uint256 initialLiquidity
    ) external onlyOwner whenNotPaused returns (address market, uint256 id) {
        if (initialLiquidity > 0) {
            collateral.transferFrom(msg.sender, address(this), initialLiquidity * 2);
        }
        PredictionMarketV3 m = new PredictionMarketV3(
            address(collateral),
            oracle,
            address(this),
            matchRef,
            question,
            endTime,
            defaultFeeBps,
            0,
            defaultMaxBet
        );
        if (initialLiquidity > 0) {
            collateral.transfer(address(m), initialLiquidity * 2);
            m.seedReserves(initialLiquidity, msg.sender);
        }
        id = ++marketCount;
        market = address(m);
        markets[id] = market;
        marketTypes[id] = 0;
        emit BinaryMarketCreated(id, market, matchRef, question);
    }

    function createMultiMarket(
        bytes32 matchRef,
        string calldata question,
        uint256 endTime,
        uint8 outcomeCount
    ) external onlyOwner whenNotPaused returns (address market, uint256 id) {
        MultiOutcomeMarket m = new MultiOutcomeMarket(
            address(collateral),
            oracle,
            matchRef,
            question,
            endTime,
            outcomeCount,
            defaultFeeBps
        );
        id = ++marketCount;
        market = address(m);
        markets[id] = market;
        marketTypes[id] = 1;
        emit MultiMarketCreated(id, market, outcomeCount, question);
    }

    function version() external pure returns (string memory) {
        return "3.0.0-phase3";
    }
}
