// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IPredictionMarket {
    function status() external view returns (uint8);
    function resolve(uint8 winningOutcome) external;
    function voidMarket() external;
}
