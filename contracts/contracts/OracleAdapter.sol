// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "./interfaces/IPredictionMarket.sol";

/// @title OracleAdapter — timelocked resolve + void for prediction markets
contract OracleAdapter is AccessControl {
    bytes32 public constant ORACLE_ROLE = keccak256("ORACLE_ROLE");

    uint256 public timelockDelay;
    address public factory;

    struct PendingResolve {
        uint8 outcome;
        uint256 executeAfter;
        bool active;
    }

    mapping(address => PendingResolve) public pending;

    event OracleResolveRequested(
        address indexed market,
        uint8 outcome,
        uint256 executeAfter
    );
    event OracleResolveConfirmed(address indexed market, uint8 outcome);
    event MarketVoided(address indexed market);

    constructor(address admin, uint256 _timelockDelay) {
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(ORACLE_ROLE, admin);
        timelockDelay = _timelockDelay;
    }

    function setTimelockDelay(uint256 delay) external onlyRole(DEFAULT_ADMIN_ROLE) {
        timelockDelay = delay;
    }

    function setFactory(address _factory) external onlyRole(DEFAULT_ADMIN_ROLE) {
        factory = _factory;
    }

    function grantOracle(address account) external onlyRole(DEFAULT_ADMIN_ROLE) {
        _grantRole(ORACLE_ROLE, account);
    }

    function requestResolve(address market, uint8 outcome) external onlyRole(ORACLE_ROLE) {
        require(outcome <= 1, "invalid outcome");
        require(IPredictionMarket(market).status() == 0, "not open");
        uint256 executeAfter = block.timestamp + timelockDelay;
        pending[market] = PendingResolve({
            outcome: outcome,
            executeAfter: executeAfter,
            active: true
        });
        emit OracleResolveRequested(market, outcome, executeAfter);
    }

    function confirmResolve(address market) external onlyRole(ORACLE_ROLE) {
        PendingResolve storage p = pending[market];
        require(p.active, "no pending");
        require(block.timestamp >= p.executeAfter, "timelock");
        p.active = false;
        IPredictionMarket(market).resolve(p.outcome);
        emit OracleResolveConfirmed(market, p.outcome);
    }

    /// @notice Fast path when timelock is 0
    function resolveNow(address market, uint8 outcome) external onlyRole(ORACLE_ROLE) {
        require(timelockDelay == 0, "use request+confirm");
        require(outcome <= 1, "invalid outcome");
        IPredictionMarket(market).resolve(outcome);
        emit OracleResolveConfirmed(market, outcome);
    }

    function voidMarket(address market) external onlyRole(ORACLE_ROLE) {
        require(IPredictionMarket(market).status() == 0, "not open");
        IPredictionMarket(market).voidMarket();
        emit MarketVoided(market);
    }
}
