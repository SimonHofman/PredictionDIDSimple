// SPDX-License-Identifier: MIT
// 许可证标识：MIT开源协议
pragma solidity ^0.8.24;
// 指定Solidity编译器版本为0.8.24及以上

// 预测市场接口 - 定义预测市场合约必须实现的核心方法
interface IPredictionMarket {
    // 获取市场当前状态（0=开放，1=已结算，2=已作废）
    function status() external view returns (uint8);
    // 结算市场，传入获胜结果编号
    function resolve(uint8 winningOutcome) external;
    // 作废市场，允许用户取回押注
    function voidMarket() external;
}
