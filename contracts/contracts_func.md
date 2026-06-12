# Contracts 函数说明文档

本文档按模块（合约文件）整理 `contracts/contracts/` 目录下 Solidity 合约中的函数，包含函数用途、参数说明与返回值说明。公开状态变量（`public`）由编译器自动生成同名 getter，下文仅列出源码中显式定义的函数。

---

## 1. MockUSDC

测试用 ERC20 抵押代币，精度 6 位（模拟 USDC）。

### 1.1 constructor

- **函数用途**
  - 部署 MockUSDC 代币，设置名称为 `"Mock USDC"`、符号为 `"mUSDC"`。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值（构造函数）。

### 1.2 decimals

- **函数用途**
  - 覆盖 ERC20 默认精度，返回 6 位小数（与 USDC 一致）。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `uint8`：固定返回 `6`。

### 1.3 mint

- **函数用途**
  - 向指定地址铸造任意数量的 mUSDC（测试环境无权限限制）。

- **函数参数说明**
  - `to`（`address`）：接收代币的钱包地址。
  - `amount`（`uint256`）：铸造数量，单位为最小单位（6 位小数）。

- **返回参数说明**
  - 无返回值。

---

## 2. IPredictionMarket（interface）

预测市场通用接口，供 Oracle Adapter 调用结算与作废。

### 2.1 status

- **函数用途**
  - 查询市场当前状态码（Open=0、Resolved=1、Voided=2，由实现合约定义）。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `uint8`：市场状态枚举的 uint8 表示。

### 2.2 resolve

- **函数用途**
  - 由预言机调用，将市场设为已结算并指定获胜结果。

- **函数参数说明**
  - `winningOutcome`（`uint8`）：获胜结果索引（二元市场通常为 0=Yes、1=No）。

- **返回参数说明**
  - 无返回值。

### 2.3 voidMarket

- **函数用途**
  - 由预言机调用，将开放中的市场作废，允许用户按原投注全额赎回。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值。

---

## 3. PredictionMarket

二元（Yes/No） parimutuel（彩池）预测市场合约，Phase 1/2 使用。

### 3.1 constructor

- **函数用途**
  - 初始化市场：绑定抵押代币、预言机、工厂、比赛引用、问题描述与结束时间，状态设为 Open。

- **函数参数说明**
  - `_collateral`（`address`）：ERC20 抵押代币合约地址，不可为零地址。
  - `_oracle`（`address`）：有权 resolve/void 的预言机地址，不可为零地址。
  - `_factory`（`address`）：创建该市场的工厂合约地址。
  - `_matchRef`（`bytes32`）：链上比赛引用（通常为 external_id 的 Keccak256 哈希）。
  - `_question`（`string`）：市场问题描述文本。
  - `_endTime`（`uint256`）：投注截止时间戳（Unix 秒），必须大于当前区块时间。

- **返回参数说明**
  - 无返回值；参数校验失败时 revert。

### 3.2 buy

- **函数用途**
  - 用户向 Yes（outcome=0）或 No（outcome=1）侧投注，从用户转入 collateral 并累加个人余额与池子。

- **函数参数说明**
  - `outcome`（`uint8`）：投注方向，0=Yes，1=No。
  - `amount`（`uint256`）：投注金额（collateral 最小单位），必须大于 0。

- **返回参数说明**
  - 无返回值；市场非 Open、已过期、outcome 无效或 amount 为 0 时 revert；成功时 emit `Bought`。

### 3.3 resolve

- **函数用途**
  - 预言机将开放市场结算为 Resolved，记录 winningOutcome（仅 `oracle` 地址可调用）。

- **函数参数说明**
  - `_winningOutcome`（`uint8`）：获胜侧，0 或 1。

- **返回参数说明**
  - 无返回值；非 Open 或 outcome 无效时 revert；成功 emit `Resolved`。

### 3.4 voidMarket

- **函数用途**
  - 预言机将开放市场作废为 Voided，用户后续可全额 claim 原投注（仅 `oracle` 可调用）。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值；非 Open 时 revert；成功 emit `MarketVoided`。

### 3.5 claim

- **函数用途**
  - 用户在 Resolved 时按 parimutuel 比例领取奖金，或在 Voided 时领取 Yes+No 原投注总额；每地址仅可 claim 一次。

- **函数参数说明**
  - 无参数（使用 `msg.sender` 作为领取者）。

- **返回参数说明**
  - 无返回值；已 claim、不可 claim 状态或 payout 为 0 时 revert；成功转账 collateral 并 emit `Claimed`。

### 3.6 _payoutResolved

- **函数用途**
  - 内部 view 函数，计算 Resolved 状态下某用户在获胜侧的 parimutuel 应付金额：`userStake * totalPool / winSideTotal`。

- **函数参数说明**
  - `user`（`address`）：待计算的用户地址。

- **返回参数说明**
  - `uint256`：应付 payout；用户在该侧无 stake 或获胜侧池为 0 时返回 0。

---

## 4. MarketFactory

Phase 2 市场工厂，由 Owner 部署 `PredictionMarket` 二元 parimutuel 市场。

### 4.1 constructor

- **函数用途**
  - 初始化工厂，绑定抵押代币与预言机地址，部署者成为 Owner。

- **函数参数说明**
  - `_collateral`（`address`）：ERC20 抵押代币地址，不可为零地址。
  - `_oracle`（`address`）：新市场的默认预言机地址，不可为零地址。

- **返回参数说明**
  - 无返回值。

### 4.2 setOracle

- **函数用途**
  - Owner 更新后续新创建市场使用的预言机地址。

- **函数参数说明**
  - `_oracle`（`address`）：新预言机地址，不可为零地址。

- **返回参数说明**
  - 无返回值；非 Owner 调用 revert。

### 4.3 createMarket

- **函数用途**
  - Owner 部署新的 `PredictionMarket` 实例，递增 `marketCount` 并写入 `markets` 映射。

- **函数参数说明**
  - `matchRef`（`bytes32`）：比赛引用哈希。
  - `question`（`string`）：市场问题（calldata）。
  - `endTime`（`uint256`）：投注截止时间戳。

- **返回参数说明**
  - `market`（`address`）：新部署的市场合约地址。
  - `marketId`（`uint256`）：自增的市场 ID（从 1 起）。
  - 成功 emit `MarketCreated`；非 Owner 调用 revert。

### 4.4 version

- **函数用途**
  - 返回工厂合约版本字符串。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `string`：固定返回 `"2.0.0-phase2"`。

---

## 5. OracleAdapter

带时间锁的预言机适配器，为 `IPredictionMarket` 提供 requestResolve + confirmResolve 或 resolveNow 快速路径。

### 5.1 constructor

- **函数用途**
  - 初始化 AccessControl，授予 admin 默认管理员与 ORACLE_ROLE，设置 timelock 延迟秒数。

- **函数参数说明**
  - `admin`（`address`）：管理员地址，同时获得 DEFAULT_ADMIN_ROLE 与 ORACLE_ROLE。
  - `_timelockDelay`（`uint256`）：requestResolve 到 confirmResolve 的最小等待秒数。

- **返回参数说明**
  - 无返回值。

### 5.2 setTimelockDelay

- **函数用途**
  - 管理员更新全局 timelock 延迟秒数。

- **函数参数说明**
  - `delay`（`uint256`）：新的延迟秒数。

- **返回参数说明**
  - 无返回值；需 DEFAULT_ADMIN_ROLE。

### 5.3 setFactory

- **函数用途**
  - 管理员设置关联的 MarketFactory 地址（当前 resolve 流程未强制校验，预留扩展）。

- **函数参数说明**
  - `_factory`（`address`）：工厂合约地址。

- **返回参数说明**
  - 无返回值；需 DEFAULT_ADMIN_ROLE。

### 5.4 grantOracle

- **函数用途**
  - 管理员为账户授予 ORACLE_ROLE，允许其发起/确认结算与作废。

- **函数参数说明**
  - `account`（`address`）：待授权地址。

- **返回参数说明**
  - 无返回值；需 DEFAULT_ADMIN_ROLE。

### 5.5 requestResolve

- **函数用途**
  - 预言机提交结算提议，写入 `pending[market]` 并设置 `executeAfter = block.timestamp + timelockDelay`；要求市场 status 为 Open（0）。

- **函数参数说明**
  - `market`（`address`）：目标预测市场合约地址。
  - `outcome`（`uint8`）：提议的获胜结果，二元市场为 0 或 1。

- **返回参数说明**
  - 无返回值；outcome 无效、市场非 Open 或非 ORACLE_ROLE 时 revert；成功 emit `OracleResolveRequested`。

### 5.6 confirmResolve

- **函数用途**
  - 在时间锁到期后，预言机确认 pending 提议并调用市场的 `resolve(outcome)`。

- **函数参数说明**
  - `market`（`address`）：目标市场合约地址。

- **返回参数说明**
  - 无返回值；无 pending、未到期 timelock 或非 ORACLE_ROLE 时 revert；成功 emit `OracleResolveConfirmed`。

### 5.7 resolveNow

- **函数用途**
  - 当 `timelockDelay == 0` 时的快速结算路径，直接调用市场 `resolve`，跳过 request/confirm。

- **函数参数说明**
  - `market`（`address`）：目标市场合约地址。
  - `outcome`（`uint8`）：获胜结果，0 或 1。

- **返回参数说明**
  - 无返回值；timelockDelay 非 0、outcome 无效或非 ORACLE_ROLE 时 revert；成功 emit `OracleResolveConfirmed`。

### 5.8 voidMarket

- **函数用途**
  - 预言机调用目标开放市场的 `voidMarket()` 作废市场。

- **函数参数说明**
  - `market`（`address`）：目标市场合约地址。

- **返回参数说明**
  - 无返回值；市场非 Open 或非 ORACLE_ROLE 时 revert；成功 emit `MarketVoided`。

---

## 6. OracleAdapterV2

m-of-n 多签预言机适配器，通过提案与审批达到 threshold 后执行 resolve。

### 6.1 constructor

- **函数用途**
  - 初始化 AccessControl 与多签阈值，admin 获得管理员与 ORACLE_ROLE。

- **函数参数说明**
  - `admin`（`address`）：管理员地址。
  - `_threshold`（`uint256`）：执行 resolve 所需的最少 ORACLE 批准数。

- **返回参数说明**
  - 无返回值。

### 6.2 setThreshold

- **函数用途**
  - 管理员更新多签批准阈值。

- **函数参数说明**
  - `t`（`uint256`）：新阈值，必须大于 0。

- **返回参数说明**
  - 无返回值；阈值为 0 或非管理员时 revert。

### 6.3 grantOracle

- **函数用途**
  - 管理员为账户授予 ORACLE_ROLE。

- **函数参数说明**
  - `account`（`address`）：待授权地址。

- **返回参数说明**
  - 无返回值；需 DEFAULT_ADMIN_ROLE。

### 6.4 proposeResolve

- **函数用途**
  - 预言机创建结算提案并自动计入发起人一次批准；若批准数已达 threshold 则立即执行。

- **函数参数说明**
  - `market`（`address`）：目标市场合约地址。
  - `outcome`（`uint8`）：获胜结果，允许 0–7（支持多结果市场）。

- **返回参数说明**
  - `id`（`uint256`）：新提案 ID（自增 `proposalCount`）。
  - 市场非 Open 或 outcome 无效时 revert；成功 emit `ProposalCreated`。

### 6.5 approveResolve

- **函数用途**
  - 预言机对已有提案追加批准；达到 threshold 时内部调用 `_execute`。

- **函数参数说明**
  - `id`（`uint256`）：提案 ID。

- **返回参数说明**
  - 无返回值；提案已执行、重复批准或非 ORACLE_ROLE 时 revert；成功 emit `ProposalApproved`。

### 6.6 _approve

- **函数用途**
  - 内部逻辑：标记 `approved[id][msg.sender]`，递增 `approvals`，达 threshold 时调用 `_execute`。

- **函数参数说明**
  - `id`（`uint256`）：提案 ID。

- **返回参数说明**
  - 无返回值。

### 6.7 _execute

- **函数用途**
  - 内部执行：标记提案已执行，调用 `IPredictionMarket(market).resolve(outcome)`。

- **函数参数说明**
  - `id`（`uint256`）：提案 ID。

- **返回参数说明**
  - 无返回值；已执行时 revert；成功 emit `ProposalExecuted`。

### 6.8 voidMarket

- **函数用途**
  - 预言机直接作废开放中的目标市场。

- **函数参数说明**
  - `market`（`address`）：目标市场合约地址。

- **返回参数说明**
  - 无返回值；市场非 Open 或非 ORACLE_ROLE 时 revert；成功 emit `MarketVoided`。

---

## 7. PredictionMarketV3

Phase 3 二元 CPMM（恒定乘积做市）市场，支持 LP 流动性、手续费与用户最大投注限制。

### 7.1 constructor

- **函数用途**
  - 初始化 CPMM 二元市场参数；若 `_initialLiquidity > 0` 则在构造时调用 `_seedReserves` 初始化储备（需合约已有足够 collateral 余额）。

- **函数参数说明**
  - `_collateral`（`address`）：ERC20 抵押代币地址，不可为零地址。
  - `_oracle`（`address`）：预言机地址，不可为零地址。
  - `_factory`（`address`）：工厂合约地址。
  - `_matchRef`（`bytes32`）：比赛引用。
  - `_question`（`string`）：市场问题。
  - `_endTime`（`uint256`）：投注截止时间。
  - `_feeBps`（`uint16`）：交易手续费基点（10000 = 100%）。
  - `_initialLiquidity`（`uint256`）：每侧初始流动性；为 0 时不在构造时 seed。
  - `_maxBetPerUser`（`uint256`）：单用户累计最大投注额；为 0 表示不限制。

- **返回参数说明**
  - 无返回值；地址为零时 revert。

### 7.2 seedReserves

- **函数用途**
  - 由工厂在转入 collateral 后调用，为市场两侧各注入 `perSide` 储备并铸造 LP 给 `lpRecipient`；仅允许首次 seed。

- **函数参数说明**
  - `perSide`（`uint256`）：YES 与 NO 各侧的储备数量（两侧相等）。
  - `lpRecipient`（`address`）：接收 LP 份额的地址（通常为创建者）。

- **返回参数说明**
  - 无返回值；非 factory 调用、已 seed 或合约 collateral 余额不足时 revert；成功 emit `LiquidityAdded`。

### 7.3 _seedReserves

- **函数用途**
  - 内部初始化：`reserveYes = reserveNo = perSide`，`lpBalance[lpRecipient] = totalLPSupply = perSide * 2`。

- **函数参数说明**
  - `perSide`（`uint256`）：每侧储备量。
  - `lpRecipient`（`address`）：LP 接收者。

- **返回参数说明**
  - 无返回值；collateral 余额不足时 revert。

### 7.4 buy

- **函数用途**
  - 用户按 CPMM 曲线购买 Yes/No 份额：扣除 feeBps 手续费后执行 `_swap`，更新用户 yes/no 余额与 `userBetTotal`。

- **函数参数说明**
  - `outcome`（`uint8`）：购买方向，0=Yes，1=No。
  - `amountIn`（`uint256`）：投入的 collateral 总量（含手续费）。

- **返回参数说明**
  - 无返回值；市场非 Open、已结束、参数无效或超过 maxBetPerUser 时 revert；成功 emit `Bought`（含 sharesOut）。

### 7.5 addLiquidity

- **函数用途**
  - 用户追加流动性：collateral 一半加 reserveYes、另一半加 reserveNo，按 1:1  mint LP。

- **函数参数说明**
  - `amount`（`uint256`）：投入的 collateral 总量，必须大于 0。

- **返回参数说明**
  - 无返回值；市场非 Open 或 amount 为 0 时 revert；成功 emit `LiquidityAdded`。

### 7.6 removeLiquidity

- **函数用途**
  - 用户销毁 LP，按当前 reserve 比例取回 yes/no 侧 collateral 总和。

- **函数参数说明**
  - `lpAmount`（`uint256`）：要销毁的 LP 数量，不得超过用户 lpBalance。

- **返回参数说明**
  - 无返回值；LP 不足或 lpAmount 为 0 时 revert；成功转账并 emit `LiquidityRemoved`。

### 7.7 resolve

- **函数用途**
  - 预言机将市场设为 Resolved 并记录 winningOutcome（仅 oracle 可调用）。

- **函数参数说明**
  - `_outcome`（`uint8`）：获胜结果，0 或 1。

- **返回参数说明**
  - 无返回值；非 Open 或 outcome 无效时 revert；成功 emit `Resolved`。

### 7.8 voidMarket

- **函数用途**
  - 预言机将开放市场作废为 Voided。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值；非 Open 时 revert；成功 emit `MarketVoided`。

### 7.9 claim

- **函数用途**
  - Resolved 时按获胜侧份额占 reserve 比例领取；Voided 时领取 yes+no 余额总和；每地址一次。

- **函数参数说明**
  - 无参数（`msg.sender` 为领取者）。

- **返回参数说明**
  - 无返回值；已 claim、不可 claim 或 payout 为 0 时 revert；成功转账 emit `Claimed`。

### 7.10 status

- **函数用途**
  - 实现 `IPredictionMarket` 兼容接口，返回 `marketStatus` 的 uint8 值。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `uint8`：Open=0，Resolved=1，Voided=2。

### 7.11 getPoolState

- **函数用途**
  - 查询当前 CPMM 储备与隐含 YES 价格（基点）。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `yesR`（`uint256`）：YES 侧 reserveYes。
  - `noR`（`uint256`）：NO 侧 reserveNo。
  - `priceYesBps`（`uint256`）：YES 隐含价格基点，公式 `(reserveNo * 10000) / (reserveYes + reserveNo)`；总储备为 0 时默认 5000。

### 7.12 _swap

- **函数用途**
  - 内部 CPMM 交换：根据 outcome 用 net 输入更新 reserve 并计算输出 sharesOut。

- **函数参数说明**
  - `outcome`（`uint8`）：0 买 Yes（减 reserveYes、增 reserveNo），1 买 No。
  - `net`（`uint256`）：扣除手续费后的净输入量。

- **返回参数说明**
  - `sharesOut`（`uint256`）：用户获得的 Yes/No 份额数量。

### 7.13 _claimResolved

- **函数用途**
  - 内部 view：Resolved 状态下计算用户按获胜侧份额占该侧 reserve 比例的可领取金额。

- **函数参数说明**
  - `user`（`address`）：用户地址。

- **返回参数说明**
  - `uint256`：应付 payout；获胜侧 reserve 为 0 时返回 0。

---

## 8. MultiOutcomeMarket

2–8 结果的多 outcome parimutuel 市场，支持 feeBps 手续费。

### 8.1 constructor

- **函数用途**
  - 初始化多结果市场，创建 `outcomeCount` 长度的 pool 数组并设为 Open。

- **函数参数说明**
  - `_collateral`（`address`）：ERC20 抵押代币地址。
  - `_oracle`（`address`）：预言机地址。
  - `_matchRef`（`bytes32`）：比赛引用。
  - `_question`（`string`）：市场问题。
  - `_endTime`（`uint256`）：投注截止时间。
  - `_outcomeCount`（`uint8`）：结果数量，必须在 2–8 之间。
  - `_feeBps`（`uint16`）：手续费基点。

- **返回参数说明**
  - 无返回值；outcomeCount 不在范围内时 revert。

### 8.2 status

- **函数用途**
  - 返回 `marketStatus` 的 uint8，供 Oracle Adapter 查询。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `uint8`：Open=0，Resolved=1，Voided=2。

### 8.3 buy

- **函数用途**
  - 用户向指定 outcome 投注，扣除 fee 后 net 计入 pool 与用户 stake。

- **函数参数说明**
  - `outcome`（`uint8`）：结果索引，必须 `< outcomeCount`。
  - `amount`（`uint256`）：投注总量（含 fee），必须大于 0。

- **返回参数说明**
  - 无返回值；市场已关闭或参数无效时 revert；成功 emit `Bought`。

### 8.4 resolve

- **函数用途**
  - 预言机结算市场，指定 winningOutcome（仅 oracle 可调用）。

- **函数参数说明**
  - `_outcome`（`uint8`）：获胜结果索引，必须 `< outcomeCount`。

- **返回参数说明**
  - 无返回值；非 Open 或 outcome 无效时 revert；成功 emit `Resolved`。

### 8.5 voidMarket

- **函数用途**
  - 预言机将开放市场作废。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值；非 Open 时 revert；成功 emit `MarketVoided`。

### 8.6 claim

- **函数用途**
  - Resolved 时按 parimutuel 从总池分配：`stake[user][winningOutcome] * totalPool / winPool`；Voided 时退还各 outcome stake 之和。

- **函数参数说明**
  - 无参数（`msg.sender` 为领取者）。

- **返回参数说明**
  - 无返回值；已 claim、不可 claim、获胜池为空或 payout 为 0 时 revert；成功转账 emit `Claimed`。

---

## 9. MarketFactoryV3

Phase 3 市场工厂，支持部署 CPMM 二元市场（PredictionMarketV3）与多结果市场（MultiOutcomeMarket），含 Pausable 暂停机制。

### 9.1 constructor

- **函数用途**
  - 初始化抵押代币、预言机、默认手续费与默认单用户最大投注（10_000 * 1e6）。

- **函数参数说明**
  - `_collateral`（`address`）：ERC20 抵押代币地址。
  - `_oracle`（`address`）：新市场默认预言机地址。
  - `_feeBps`（`uint16`）：默认手续费基点。

- **返回参数说明**
  - 无返回值；部署者成为 Owner。

### 9.2 pause

- **函数用途**
  - Owner 暂停工厂，阻止新的 createBinaryMarket / createMultiMarket。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值；非 Owner 时 revert。

### 9.3 unpause

- **函数用途**
  - Owner 恢复工厂，允许创建市场。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值；非 Owner 时 revert。

### 9.4 setOracle

- **函数用途**
  - Owner 更新后续新市场的预言机地址。

- **函数参数说明**
  - `_oracle`（`address`）：新预言机地址。

- **返回参数说明**
  - 无返回值。

### 9.5 setDefaultFeeBps

- **函数用途**
  - Owner 更新新市场的默认手续费基点。

- **函数参数说明**
  - `bps`（`uint16`）：新默认 feeBps。

- **返回参数说明**
  - 无返回值。

### 9.6 createBinaryMarket

- **函数用途**
  - Owner 部署 `PredictionMarketV3`；若 `initialLiquidity > 0`，从 Owner 转入 `initialLiquidity * 2` collateral 并 seed 储备，LP 给 Owner。

- **函数参数说明**
  - `matchRef`（`bytes32`）：比赛引用。
  - `question`（`string`）：市场问题。
  - `endTime`（`uint256`）：投注截止时间。
  - `initialLiquidity`（`uint256`）：每侧初始流动性；为 0 时不 seed。

- **返回参数说明**
  - `market`（`address`）：新部署的 PredictionMarketV3 地址。
  - `id`（`uint256`）：自增市场 ID；`marketTypes[id] = 0`（binary v3）。
  - 工厂暂停或非 Owner 时 revert；成功 emit `BinaryMarketCreated`。

### 9.7 createMultiMarket

- **函数用途**
  - Owner 部署 `MultiOutcomeMarket` 多结果 parimutuel 市场。

- **函数参数说明**
  - `matchRef`（`bytes32`）：比赛引用。
  - `question`（`string`）：市场问题。
  - `endTime`（`uint256`）：投注截止时间。
  - `outcomeCount`（`uint8`）：结果数量（2–8，由 MultiOutcomeMarket 构造校验）。

- **返回参数说明**
  - `market`（`address`）：新部署的 MultiOutcomeMarket 地址。
  - `id`（`uint256`）：自增市场 ID；`marketTypes[id] = 1`（multi）。
  - 工厂暂停或非 Owner 时 revert；成功 emit `MultiMarketCreated`。

### 9.8 version

- **函数用途**
  - 返回工厂版本字符串。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - `string`：固定返回 `"3.0.0-phase3"`。

---

## 10. DIDRegistry

可选链上 DID 哈希绑定注册表，用户通过以太坊签名证明身份后绑定 `didHash`。

### 10.1 constructor

- **函数用途**
  - 部署 DID 注册表，部署者成为 Owner。

- **函数参数说明**
  - 无参数。

- **返回参数说明**
  - 无返回值。

### 10.2 bindDid

- **函数用途**
  - 用户绑定 DID 哈希：验证对 `keccak256("BindDID:", msg.sender, didHash)` 的 EIP-191 签名，签名者必须为 `msg.sender`，通过后写入 `didHashOf[msg.sender]`。

- **函数参数说明**
  - `didHash`（`bytes32`）：DID 内容的 Keccak256 哈希，不可为零。
  - `signature`（`bytes`）：用户对上述 digest 的 ECDSA 签名（calldata）。

- **返回参数说明**
  - 无返回值；didHash 为空或签名无效时 revert；成功 emit `DidBound`。

### 10.3 resolveDid

- **函数用途**
  - 查询某账户已绑定的 DID 哈希。

- **函数参数说明**
  - `account`（`address`）：待查询的钱包地址。

- **返回参数说明**
  - `bytes32`：该账户绑定的 didHash；未绑定时为 `bytes32(0)`。

---

## 附录：public 状态变量自动 getter

以下 `public` 变量由 Solidity 编译器生成同名 view getter（单参数 mapping 需传入 key），未在上文逐条列出：

- **PredictionMarket**：`collateral`、`oracle`、`factory`、`matchRef`、`question`、`endTime`、`status`、`winningOutcome`、`yesPool`、`noPool`、`yesBalance(address)`、`noBalance(address)`、`claimed(address)`
- **MarketFactory**：`collateral`、`oracle`、`marketCount`、`markets(uint256)`
- **OracleAdapter**：`ORACLE_ROLE`、`timelockDelay`、`factory`、`pending(address)`
- **OracleAdapterV2**：`ORACLE_ROLE`、`threshold`、`proposalCount`、`proposals(uint256)`、`approved(uint256,address)`
- **PredictionMarketV3**：`collateral`、`oracle`、`factory`、`matchRef`、`question`、`endTime`、`feeBps`、`maxBetPerUser`、`marketStatus`、`winningOutcome`、`reserveYes`、`reserveNo`、`totalLPSupply`、`collectedFees`、`yesBalance(address)`、`noBalance(address)`、`lpBalance(address)`、`userBetTotal(address)`、`claimed(address)`
- **MultiOutcomeMarket**：`collateral`、`oracle`、`matchRef`、`question`、`endTime`、`outcomeCount`、`feeBps`、`marketStatus`、`winningOutcome`、`pool(uint256)`、`stake(address,uint8)`、`claimed(address)`
- **MarketFactoryV3**：`collateral`、`oracle`、`defaultFeeBps`、`defaultMaxBet`、`marketCount`、`markets(uint256)`、`marketTypes(uint256)`
- **DIDRegistry**：`didHashOf(address)`
