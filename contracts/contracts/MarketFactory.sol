// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "./PredictionMarket.sol";

/// @title MarketFactory — deploys Yes/No prediction markets
contract MarketFactory is Ownable {
    IERC20 public immutable collateral;
    address public oracle;

    uint256 public marketCount;

    mapping(uint256 => address) public markets;

    event MarketCreated(
        uint256 indexed marketId,
        address indexed market,
        bytes32 indexed matchRef,
        string question,
        uint256 endTime
    );

    constructor(address _collateral, address _oracle) Ownable(msg.sender) {
        require(_collateral != address(0), "collateral");
        require(_oracle != address(0), "oracle");
        collateral = IERC20(_collateral);
        oracle = _oracle;
    }

    function setOracle(address _oracle) external onlyOwner {
        require(_oracle != address(0), "oracle");
        oracle = _oracle;
    }

    function createMarket(
        bytes32 matchRef,
        string calldata question,
        uint256 endTime
    ) external onlyOwner returns (address market, uint256 marketId) {
        PredictionMarket m = new PredictionMarket(
            address(collateral),
            oracle,
            address(this),
            matchRef,
            question,
            endTime
        );
        marketId = ++marketCount;
        market = address(m);
        markets[marketId] = market;
        emit MarketCreated(marketId, market, matchRef, question, endTime);
    }

    function version() external pure returns (string memory) {
        return "2.0.0-phase2";
    }
}
