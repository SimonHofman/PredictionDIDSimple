// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import "@openzeppelin/contracts/access/AccessControl.sol";
import "./interfaces/IPredictionMarket.sol";

/// @title OracleAdapterV2 — m-of-n multisig resolve + void
contract OracleAdapterV2 is AccessControl {
    bytes32 public constant ORACLE_ROLE = keccak256("ORACLE_ROLE");

    uint256 public threshold;
    uint256 public proposalCount;

    struct Proposal {
        address market;
        uint8 outcome;
        uint256 approvals;
        bool executed;
    }

    mapping(uint256 => Proposal) public proposals;
    mapping(uint256 => mapping(address => bool)) public approved;

    event ProposalCreated(uint256 indexed id, address indexed market, uint8 outcome);
    event ProposalApproved(uint256 indexed id, address indexed oracle);
    event ProposalExecuted(uint256 indexed id, uint8 outcome);
    event MarketVoided(address indexed market);

    constructor(address admin, uint256 _threshold) {
        _grantRole(DEFAULT_ADMIN_ROLE, admin);
        _grantRole(ORACLE_ROLE, admin);
        threshold = _threshold;
    }

    function setThreshold(uint256 t) external onlyRole(DEFAULT_ADMIN_ROLE) {
        require(t > 0, "threshold");
        threshold = t;
    }

    function grantOracle(address account) external onlyRole(DEFAULT_ADMIN_ROLE) {
        _grantRole(ORACLE_ROLE, account);
    }

    function proposeResolve(address market, uint8 outcome) external onlyRole(ORACLE_ROLE) returns (uint256 id) {
        require(outcome <= 7, "outcome");
        require(IPredictionMarket(market).status() == 0, "not open");
        id = ++proposalCount;
        proposals[id] = Proposal({market: market, outcome: outcome, approvals: 0, executed: false});
        _approve(id);
        emit ProposalCreated(id, market, outcome);
    }

    function approveResolve(uint256 id) external onlyRole(ORACLE_ROLE) {
        _approve(id);
    }

    function _approve(uint256 id) internal {
        Proposal storage p = proposals[id];
        require(!p.executed, "executed");
        require(!approved[id][msg.sender], "approved");
        approved[id][msg.sender] = true;
        p.approvals++;
        emit ProposalApproved(id, msg.sender);
        if (p.approvals >= threshold) {
            _execute(id);
        }
    }

    function _execute(uint256 id) internal {
        Proposal storage p = proposals[id];
        require(!p.executed, "executed");
        p.executed = true;
        IPredictionMarket(p.market).resolve(p.outcome);
        emit ProposalExecuted(id, p.outcome);
    }

    function voidMarket(address market) external onlyRole(ORACLE_ROLE) {
        require(IPredictionMarket(market).status() == 0, "not open");
        IPredictionMarket(market).voidMarket();
        emit MarketVoided(market);
    }
}
