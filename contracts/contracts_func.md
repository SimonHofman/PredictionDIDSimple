# Contracts 结构体与函数说明文档

本文档按模块（合约文件）整理 `contracts/contracts/` 目录下 Solidity 源码。
范围不含 `node_modules/`、`scripts/`、`test/`。
`public` 状态变量由编译器自动生成 getter，下文在「状态变量」节列出并注释；显式函数单独成节。

---

## 1. MockUSDC.sol

#### 合约/模块说明

- 测试用 ERC20 抵押代币，6 位小数模拟 USDC（mUSDC）。

#### 规范与约定

- Solidity ^0.8.24；SPDX MIT。

- mint 无权限限制，仅用于测试/本地链。

#### 继承关系

- 继承 OpenZeppelin `ERC20`。

- 覆盖 `decimals()` 返回 6。

#### 事件

- 继承 ERC20 标准 Transfer/Approval 事件。



### 1.1 MockUSDC 状态变量

#### 状态变量说明

- 下列为合约显式声明的状态变量；`public` 变量由编译器自动生成同名 getter。

#### 状态变量（每个变量一行）

- 无额外状态变量（名称/符号/总供给由 ERC20 基类管理）。

---

### 1.2 constructor

- **函数用途**

  - 部署 MockUSDC，初始化 ERC20 名称为 "Mock USDC"、符号 "mUSDC"。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值（构造函数）。

- **函数内校验**

  - 无额外 require。

- **函数实现效果**

  - 调用 ERC20("Mock USDC", "mUSDC") 初始化。

- **错误返回**

  - 无

---

### 1.3 decimals

- **函数用途**

  - 覆盖 ERC20.decimals，固定 6 位小数。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - `uint8`：恒为 6

- **函数内校验**

  - 无。

- **函数实现效果**

  - `pure override` 直接 return 6。

- **错误返回**

  - 无 revert

---

### 1.4 mint

- **函数用途**

  - 向任意地址铸造 mUSDC。

- **函数参数说明**

  - `to`（address）：接收地址

  - `amount`（uint256）：铸造数量（6 位最小单位）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - 无权限校验（测试合约）。

- **函数实现效果**

  - 内部 `_mint(to, amount)`。

- **错误返回**

  - ERC20 内部溢出/零地址由 OZ 处理

---

## 2. IPredictionMarket.sol（interface）

#### 合约/模块说明

- 预测市场通用接口，供 OracleAdapter / OracleAdapterV2 调用结算与作废。

#### 规范与约定

- status 返回 uint8：Open=0、Resolved=1、Voided=2（与实现合约 enum 对应）。

#### 继承关系

- 无继承；由 PredictionMarket、PredictionMarketV3、MultiOutcomeMarket 等实现。



### 2.1 status

- **函数用途**

  - 查询市场状态码。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - `uint8`：市场状态

- **函数内校验**

  - 由实现合约定义。

- **函数实现效果**

  - view 读取 marketStatus/status。

- **错误返回**

  - 无

---

### 2.2 resolve

- **函数用途**

  - 预言机结算市场并指定获胜 outcome。

- **函数参数说明**

  - `winningOutcome`（uint8）：获胜结果索引

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - 实现合约校验 Open 与 outcome 范围。

- **函数实现效果**

  - 写入 winningOutcome，状态→Resolved。

- **错误返回**

  - 非 Open/outcome 无效：revert

---

### 2.3 voidMarket

- **函数用途**

  - 预言机作废开放中的市场。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - 实现合约要求 Open。

- **函数实现效果**

  - 状态→Voided，用户可全额 claim 原投注。

- **错误返回**

  - 非 Open：revert "not open"

---

## 3. PredictionMarket.sol

#### 合约/模块说明

- Phase 1/2 二元（Yes/No）parimutuel 预测市场。

#### 规范与约定

- outcome：0=Yes，1=No。

- SafeERC20 转账；onlyOracle 修饰 resolve/void。

#### 继承关系

- 无显式继承；使用 OZ IERC20/SafeERC20。

- 实现 IPredictionMarket 语义（未 formal inherit）。

#### 事件

- `Bought(user, outcome, amount)`：用户下注

- `Resolved(winningOutcome)`：已结算

- `Claimed(user, amount)`：用户领取

- `MarketVoided()`：市场作废



### 3.1 Status 枚举

#### 枚举说明

- 市场生命周期状态。

#### 枚举值

- `Open`（0）：可下注

- `Resolved`（1）：已结算，可按 parimutuel 领取

- `Voided`（2）：已作废，可全额退回原投注

---

### 3.2 PredictionMarket 状态变量

#### 状态变量说明

- 下列为合约显式声明的状态变量；`public` 变量由编译器自动生成同名 getter。

#### 状态变量（每个变量一行）

- `collateral`（IERC20 public immutable）：ERC20 抵押代币

- `oracle`（address public immutable）：有权 resolve/void 的预言机地址

- `factory`（address public immutable）：创建该市场的工厂地址

- `matchRef`（bytes32 public immutable）：链上比赛引用哈希

- `question`（string public）：市场问题描述

- `endTime`（uint256 public）：投注截止 Unix 时间戳

- `status`（Status public）：当前市场状态

- `winningOutcome`（uint8 public）：结算后获胜 outcome

- `yesPool`（uint256 public）：Yes 侧池子总量

- `noPool`（uint256 public）：No 侧池子总量

- `yesBalance`（mapping(address=>uint256) public）：用户 Yes 持仓

- `noBalance`（mapping(address=>uint256) public）：用户 No 持仓

- `claimed`（mapping(address=>bool) public）：用户是否已 claim

---

### 3.3 modifier onlyOracle

- **修饰器用途**

  - 限制仅 oracle 地址可调用。

- **函数内校验**

  - `msg.sender == oracle`

- **实现效果**

  - 通过则执行函数体

- **错误返回**

  - `"not oracle"`

---

### 3.4 constructor

- **函数用途**

  - 初始化二元 parimutuel 市场。

- **函数参数说明**

  - `_collateral`（address）：抵押代币

  - `_oracle`（address）：预言机

  - `_factory`（address）：工厂

  - `_matchRef`（bytes32）：比赛引用

  - `_question`（string memory）：问题

  - `_endTime`（uint256）：截止时间

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - `_collateral != 0`

  - `_oracle != 0`

  - `_endTime > block.timestamp`

- **函数实现效果**

  - 赋值各 immutable/状态；status=Open

- **错误返回**

  - `"collateral"`/`"oracle"`/`"end in past"`

---

### 3.5 buy

- **函数用途**

  - 用户向 Yes/No 侧投注。

- **函数参数说明**

  - `outcome`（uint8）：0=Yes，1=No

  - `amount`（uint256）：投注数量

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - status==Open

  - block.timestamp < endTime

  - outcome<=1

  - amount>0

- **函数实现效果**

  - safeTransferFrom 收 collateral

  - 累加 yesBalance/noBalance 与 yesPool/noPool

  - emit Bought

- **错误返回**

  - `"not open"`/`"ended"`/`"invalid outcome"`/`"zero amount"`

---

### 3.6 resolve

- **函数用途**

  - 预言机结算市场。

- **函数参数说明**

  - `_winningOutcome`（uint8）：获胜侧 0 或 1

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyOracle

  - status==Open

  - _winningOutcome<=1

- **函数实现效果**

  - status=Resolved；winningOutcome 写入；emit Resolved

- **错误返回**

  - `"not open"`/`"invalid outcome"`/`"not oracle"`

---

### 3.7 voidMarket

- **函数用途**

  - 预言机作废市场。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyOracle

  - status==Open

- **函数实现效果**

  - status=Voided；emit MarketVoided

- **错误返回**

  - `"not open"`/`"not oracle"`

---

### 3.8 claim

- **函数用途**

  - 用户领取 Resolved 奖金或 Voided 退款。

- **函数参数说明**

  - 无参数（msg.sender）。

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - !claimed[sender]

  - status 为 Resolved 或 Voided

  - payout>0

- **函数实现效果**

  - Resolved：parimutuel `_payoutResolved`

  - Voided：yesBalance+noBalance

  - claimed=true；safeTransfer；emit Claimed

- **错误返回**

  - `"already claimed"`/`"not claimable"`/`"nothing to claim"`

---

### 3.9 _payoutResolved

- **函数用途**

  - 内部 view：计算 Resolved 下用户 payout。

- **函数参数说明**

  - `user`（address）：用户地址

- **函数返回参数说明**

  - `uint256`：应付金额，无 stake 或获胜池为 0 时返回 0

- **函数内校验**

  - winningOutcome 决定 userStake 与 winSideTotal

- **函数实现效果**

  - `(userStake * totalPool) / winSideTotal`

- **错误返回**

  - 无 revert（返回 0）

---

## 4. MarketFactory.sol

#### 合约/模块说明

- Phase 2 市场工厂，Owner 部署 PredictionMarket 实例。

#### 规范与约定

- marketId 从 1 自增；createMarket 仅 Owner。

#### 继承关系

- 继承 OpenZeppelin `Ownable`（部署者为 Owner）。

#### 事件

- `MarketCreated(marketId, market, matchRef, question, endTime)`



### 4.1 MarketFactory 状态变量

#### 状态变量说明

- 下列为合约显式声明的状态变量；`public` 变量由编译器自动生成同名 getter。

#### 状态变量（每个变量一行）

- `collateral`（IERC20 public immutable）：抵押代币

- `oracle`（address public）：新市场默认预言机（可 setOracle 更新）

- `marketCount`（uint256 public）：已创建市场计数

- `markets`（mapping(uint256=>address) public）：marketId→市场地址

---

### 4.2 constructor

- **函数用途**

  - 初始化工厂。

- **函数参数说明**

  - `_collateral`

  - `_oracle`

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - 两者非零地址

- **函数实现效果**

  - Ownable(msg.sender)；保存 collateral/oracle

- **错误返回**

  - `"collateral"`/`"oracle"`

---

### 4.3 setOracle

- **函数用途**

  - Owner 更新后续市场的 oracle。

- **函数参数说明**

  - `_oracle`（address）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyOwner

  - _oracle!=0

- **函数实现效果**

  - oracle=_oracle

- **错误返回**

  - `"oracle"`

---

### 4.4 createMarket

- **函数用途**

  - Owner 部署新 PredictionMarket。

- **函数参数说明**

  - `matchRef`（bytes32）

  - `question`（string calldata）

  - `endTime`（uint256）

- **函数返回参数说明**

  - `market`（address）：新合约地址

  - `marketId`（uint256）：自增 ID

- **函数内校验**

  - onlyOwner

- **函数实现效果**

  - new PredictionMarket(...)

  - marketCount++；markets[id]=market

  - emit MarketCreated

- **错误返回**

  - 构造参数 invalid 时 revert

---

### 4.5 version

- **函数用途**

  - 返回工厂版本字符串。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - `string memory`："2.0.0-phase2"

- **函数内校验**

  - 无

- **函数实现效果**

  - pure 返回常量

- **错误返回**

  - 无

---

## 5. OracleAdapter.sol

#### 合约/模块说明

- 带时间锁的预言机适配器：requestResolve + confirmResolve，或 resolveNow 快速路径。

#### 规范与约定

- ORACLE_ROLE 可发起/确认结算；DEFAULT_ADMIN_ROLE 管理配置。

- pending 映射存 timelock 提议。

#### 继承关系

- 继承 OpenZeppelin `AccessControl`。

#### 事件

- `OracleResolveRequested(market, outcome, executeAfter)`

- `OracleResolveConfirmed(market, outcome)`

- `MarketVoided(market)`



### 5.1 PendingResolve 结构体

#### 结构体说明

- 单市场待确认结算提议。

#### 规范与约定

- 存储于 mapping(address=>PendingResolve) pending。

#### 组合/依赖关系

- requestResolve 写入；confirmResolve 读取并清除 active。

#### 结构体字段

- `outcome`（uint8）：提议获胜 outcome

- `executeAfter`（uint256）：最早可 confirm 的时间戳

- `active`（bool）：提议是否有效

---

### 5.2 OracleAdapter 状态变量

#### 状态变量说明

- 下列为合约显式声明的状态变量；`public` 变量由编译器自动生成同名 getter。

#### 状态变量（每个变量一行）

- `ORACLE_ROLE`（bytes32 public constant）：预言机角色标识

- `timelockDelay`（uint256 public）：request 到 confirm 的最小等待秒数

- `factory`（address public）：关联工厂（预留）

- `pending`（mapping(address=>PendingResolve) public）：市场→待确认提议

---

### 5.3 constructor

- **函数用途**

  - 初始化 AccessControl 与 timelock。

- **函数参数说明**

  - `admin`（address）

  - `_timelockDelay`（uint256）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - 无

- **函数实现效果**

  - admin 获得 DEFAULT_ADMIN_ROLE 与 ORACLE_ROLE

  - timelockDelay=_timelockDelay

- **错误返回**

  - 无

---

### 5.4 setTimelockDelay

- **函数用途**

  - 管理员更新 timelock 秒数。

- **函数参数说明**

  - `delay`（uint256）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyRole(DEFAULT_ADMIN_ROLE)

- **函数实现效果**

  - timelockDelay=delay

- **错误返回**

  - 无 admin 权限 revert

---

### 5.5 setFactory

- **函数用途**

  - 管理员设置 factory 地址。

- **函数参数说明**

  - `_factory`（address）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyRole(DEFAULT_ADMIN_ROLE)

- **函数实现效果**

  - factory=_factory

- **错误返回**

  - 无 admin revert

---

### 5.6 grantOracle

- **函数用途**

  - 管理员授予 ORACLE_ROLE。

- **函数参数说明**

  - `account`（address）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyRole(DEFAULT_ADMIN_ROLE)

- **函数实现效果**

  - _grantRole(ORACLE_ROLE, account)

- **错误返回**

  - 无 admin revert

---

### 5.7 requestResolve

- **函数用途**

  - 预言机提交带 timelock 的结算请求。

- **函数参数说明**

  - `market`（address）

  - `outcome`（uint8）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyRole(ORACLE_ROLE)

  - outcome<=1

  - IPredictionMarket(market).status()==0

- **函数实现效果**

  - executeAfter=block.timestamp+timelockDelay

  - pending[market] 写入

  - emit OracleResolveRequested

- **错误返回**

  - `"invalid outcome"`/`"not open"`

---

### 5.8 confirmResolve

- **函数用途**

  - timelock 到期后确认并 resolve。

- **函数参数说明**

  - `market`（address）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyRole(ORACLE_ROLE)

  - p.active

  - block.timestamp>=p.executeAfter

- **函数实现效果**

  - p.active=false；IPredictionMarket.resolve(outcome)；emit Confirmed

- **错误返回**

  - `"no pending"`/`"timelock"`

---

### 5.9 resolveNow

- **函数用途**

  - timelockDelay==0 时立即 resolve。

- **函数参数说明**

  - `market`

  - `outcome`

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyRole(ORACLE_ROLE)

  - timelockDelay==0

  - outcome<=1

- **函数实现效果**

  - 直接 IPredictionMarket.resolve；emit Confirmed

- **错误返回**

  - `"use request+confirm"`

---

### 5.10 voidMarket

- **函数用途**

  - 预言机作废目标市场。

- **函数参数说明**

  - `market`（address）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyRole(ORACLE_ROLE)

  - market.status()==0

- **函数实现效果**

  - IPredictionMarket.voidMarket；emit MarketVoided

- **错误返回**

  - `"not open"`

---

## 6. OracleAdapterV2.sol

#### 合约/模块说明

- m-of-n 多签预言机：提案 + 批准达 threshold 后 execute resolve。

#### 规范与约定

- outcome 允许 0–7 以支持多结果市场。

- proposeResolve 自动计入发起人一次 approve。

#### 继承关系

- 继承 OpenZeppelin `AccessControl`。

#### 事件

- ProposalCreated/Approved/Executed

- MarketVoided



### 6.1 Proposal 结构体

#### 结构体说明

- 结算提案。

#### 规范与约定

- approvals 达 threshold 时 _execute。

#### 组合/依赖关系

- proposals[id] 存储。

#### 结构体字段

- `market`（address）：目标市场

- `outcome`（uint8）：提议 outcome

- `approvals`（uint256）：当前批准数

- `executed`（bool）：是否已执行

---

### 6.2 OracleAdapterV2 状态变量

#### 状态变量说明

- 下列为合约显式声明的状态变量；`public` 变量由编译器自动生成同名 getter。

#### 状态变量（每个变量一行）

- `ORACLE_ROLE`（bytes32 public constant）

- `threshold`（uint256 public）：执行所需最少 ORACLE 批准数

- `proposalCount`（uint256 public）：提案自增计数

- `proposals`（mapping(uint256=>Proposal) public）

- `approved`（mapping(uint256=>mapping(address=>bool)) public）：提案→预言机→是否已批

---

### 6.3 constructor

- **函数用途**

  - 初始化多签阈值与角色。

- **函数参数说明**

  - `admin`

  - `_threshold`

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - 无

- **函数实现效果**

  - admin 获 ADMIN+ORACLE；threshold=_threshold

- **错误返回**

  - 无

---

### 6.4 setThreshold

- **函数用途**

  - 更新 threshold。

- **函数参数说明**

  - `t`（uint256）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyRole(DEFAULT_ADMIN_ROLE)

  - t>0

- **函数实现效果**

  - threshold=t

- **错误返回**

  - `"threshold"`

---

### 6.5 grantOracle

- **函数用途**

  - 授予 ORACLE_ROLE。

- **函数参数说明**

  - `account`

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyRole(DEFAULT_ADMIN_ROLE)

- **函数实现效果**

  - _grantRole

- **错误返回**

  - 无 admin revert

---

### 6.6 proposeResolve

- **函数用途**

  - 创建提案并自动 _approve 一次。

- **函数参数说明**

  - `market`

  - `outcome`

- **函数返回参数说明**

  - `id`（uint256）：新提案 ID

- **函数内校验**

  - onlyRole(ORACLE_ROLE)

  - outcome<=7

  - market.status()==0

- **函数实现效果**

  - proposalCount++；写入 Proposal；_approve(id)；emit Created

- **错误返回**

  - `"outcome"`/`"not open"`

---

### 6.7 approveResolve

- **函数用途**

  - 对已有提案追加批准。

- **函数参数说明**

  - `id`（uint256）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyRole(ORACLE_ROLE)

- **函数实现效果**

  - _approve(id)

- **错误返回**

  - _approve 内 revert

---

### 6.8 _approve

- **函数用途**

  - 内部：标记 approved、递增 approvals，达 threshold 则 _execute。

- **函数参数说明**

  - `id`

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - !p.executed

  - !approved[id][sender]

- **函数实现效果**

  - approved=true；approvals++；emit Approved；可能 _execute

- **错误返回**

  - `"executed"`/`"approved"`

---

### 6.9 _execute

- **函数用途**

  - 内部：标记 executed 并 resolve。

- **函数参数说明**

  - `id`

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - !p.executed

- **函数实现效果**

  - p.executed=true；IPredictionMarket.resolve；emit Executed

- **错误返回**

  - `"executed"`

---

### 6.10 voidMarket

- **函数用途**

  - 作废开放市场。

- **函数参数说明**

  - `market`

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyRole(ORACLE_ROLE)

  - status==0

- **函数实现效果**

  - voidMarket；emit MarketVoided

- **错误返回**

  - `"not open"`

---

## 7. PredictionMarketV3.sol

#### 合约/模块说明

- Phase 3 二元 CPMM 市场：恒定乘积做市、LP、手续费、单用户 maxBet。

#### 规范与约定

- feeBps 基数 10000；buy 扣 fee 后 _swap。

- ReentrancyGuard + nonReentrant 保护 claim/LP/buy。

#### 继承关系

- 继承 OpenZeppelin `ReentrancyGuard`。

- 实现 IPredictionMarket.status/resolve/voidMarket 语义。

#### 事件

- Bought

- LiquidityAdded/Removed

- Resolved

- Claimed

- MarketVoided



### 7.1 Status 枚举

#### 枚举说明

- 同 PredictionMarket：Open/Resolved/Voided。

#### 枚举值

- Open(0)

- Resolved(1)

- Voided(2)

---

### 7.2 PredictionMarketV3 状态变量

#### 状态变量说明

- 下列为合约显式声明的状态变量；`public` 变量由编译器自动生成同名 getter。

#### 状态变量（每个变量一行）

- `collateral`（IERC20 public immutable）

- `oracle`（address public immutable）

- `factory`（address public immutable）

- `matchRef`（bytes32 public immutable）

- `question`（string public）

- `endTime`（uint256 public）

- `feeBps`（uint16 public）：交易手续费基点

- `maxBetPerUser`（uint256 public）：单用户累计最大投注，0=不限制

- `marketStatus`（Status public）

- `winningOutcome`（uint8 public）

- `reserveYes/reserveNo`（uint256 public）：CPMM 两侧储备

- `totalLPSupply`（uint256 public）：LP 总供给

- `collectedFees`（uint256 public）：累计手续费

- `yesBalance/noBalance`（mapping public）：用户份额

- `lpBalance`（mapping public）：用户 LP

- `userBetTotal`（mapping public）：用户累计投注

- `claimed`（mapping public）：是否已 claim

---

### 7.3 modifier onlyOracle

- **修饰器用途**

  - 仅 oracle 可调 resolve/void。

- **函数内校验**

  - msg.sender==oracle

- **实现效果**

  - 执行函数体

- **错误返回**

  - `"not oracle"`

---

### 7.4 constructor

- **函数用途**

  - 初始化 CPMM 市场。

- **函数参数说明**

  - `_collateral`

  - `_oracle`

  - `_factory`

  - `_matchRef`

  - `_question`

  - `_endTime`

  - `_feeBps`

  - `_initialLiquidity`

  - `_maxBetPerUser`

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - collateral 与 oracle 非零

- **函数实现效果**

  - 赋值字段；marketStatus=Open

  - 若 _initialLiquidity>0 则 _seedReserves(..., msg.sender)

- **错误返回**

  - `"zero addr"`

---

### 7.5 seedReserves

- **函数用途**

  - 工厂在转入 collateral 后 seed 储备。

- **函数参数说明**

  - `perSide`（uint256）

  - `lpRecipient`（address）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - msg.sender==factory

  - totalLPSupply==0

- **函数实现效果**

  - _seedReserves

- **错误返回**

  - `"not factory"`/`"already seeded"`

---

### 7.6 _seedReserves

- **函数用途**

  - 内部初始化 reserve 与 LP。

- **函数参数说明**

  - `perSide`

  - `lpRecipient`

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - collateral.balanceOf(this) >= perSide*2

- **函数实现效果**

  - reserveYes=reserveNo=perSide；mint LP；emit LiquidityAdded

- **错误返回**

  - `"insufficient seed"`

---

### 7.7 buy

- **函数用途**

  - CPMM 购买 Yes/No 份额。

- **函数参数说明**

  - `outcome`（uint8）

  - `amountIn`（uint256）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - nonReentrant

  - Open

  - 未过 endTime

  - outcome<=1

  - amountIn>0

  - maxBetPerUser==0 或 userBetTotal+amountIn<=maxBet

- **函数实现效果**

  - 扣 fee；safeTransferFrom

  - _swap 得 sharesOut

  - 更新 yesBalance/noBalance 与 userBetTotal

  - emit Bought

- **错误返回**

  - 各类 require 字符串

---

### 7.8 addLiquidity

- **函数用途**

  - 追加流动性，1:1 注入 reserveYes/No。

- **函数参数说明**

  - `amount`（uint256）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - nonReentrant

  - Open

  - amount>0

- **函数实现效果**

  - half 加 reserveYes，余量加 reserveNo

  - mint LP

  - emit LiquidityAdded

- **错误返回**

  - `"not open"`/`"zero"`

---

### 7.9 removeLiquidity

- **函数用途**

  - 销毁 LP 按比例取回 collateral。

- **函数参数说明**

  - `lpAmount`（uint256）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - nonReentrant

  - lpAmount>0

  - lpBalance>=lpAmount

- **函数实现效果**

  - 按 reserve 比例计算 yesOut/noOut

  - 更新 reserve/LP/totalLPSupply

  - safeTransfer

  - emit LiquidityRemoved

- **错误返回**

  - `"lp"`

---

### 7.10 resolve

- **函数用途**

  - 预言机结算。

- **函数参数说明**

  - `_outcome`（uint8）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyOracle

  - Open

  - _outcome<=1

- **函数实现效果**

  - marketStatus=Resolved；emit Resolved

- **错误返回**

  - `"not open"`/`"invalid"`

---

### 7.11 voidMarket

- **函数用途**

  - 预言机作废。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyOracle

  - Open

- **函数实现效果**

  - marketStatus=Voided

- **错误返回**

  - `"not open"`

---

### 7.12 claim

- **函数用途**

  - 领取 Resolved/Voided 收益。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - nonReentrant

  - !claimed

  - payout>0

  - Resolved 或 Voided

- **函数实现效果**

  - Resolved: _claimResolved

  - Voided: yes+no balance

  - claimed=true；transfer

- **错误返回**

  - `"claimed"`/`"not claimable"`/`"nothing"`

---

### 7.13 status

- **函数用途**

  - IPredictionMarket 兼容 getter。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - `uint8`：marketStatus

- **函数内校验**

  - 无

- **函数实现效果**

  - return uint8(marketStatus)

- **错误返回**

  - 无

---

### 7.14 getPoolState

- **函数用途**

  - 查询 CPMM 储备与 YES 价格 bps。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - `yesR`（uint256）：reserveYes

  - `noR`（uint256）：reserveNo

  - `priceYesBps`（uint256）：(reserveNo*10000)/total，total=0 时 5000

- **函数内校验**

  - 无

- **函数实现效果**

  - view 返回三值

- **错误返回**

  - 无

---

### 7.15 _swap

- **函数用途**

  - 内部 CPMM 交换。

- **函数参数说明**

  - `outcome`（uint8）

  - `net`（uint256）：扣费后净输入

- **函数返回参数说明**

  - `sharesOut`（uint256）：输出份额

- **函数内校验**

  - outcome 0 或 1

- **函数实现效果**

  - outcome=0：减 reserveYes 增 reserveNo

  - outcome=1：减 reserveNo 增 reserveYes

- **错误返回**

  - 无 revert（依赖 reserve 充足）

---

### 7.16 _claimResolved

- **函数用途**

  - 内部 view：Resolved 领取额。

- **函数参数说明**

  - `user`（address）

- **函数返回参数说明**

  - `uint256` payout

- **函数内校验**

  - 获胜侧 reserve 为 0 返回 0

- **函数实现效果**

  - (balance*total)/winReserve

- **错误返回**

  - 无

---

## 8. MultiOutcomeMarket.sol

#### 合约/模块说明

- 2–8 结果 parimutuel 多选项市场，含 feeBps。

#### 规范与约定

- pool[] 各 outcome 池；stake[user][outcome] 用户投注。

- fee 从 amount 扣除，net 入池。

#### 继承关系

- 继承 `ReentrancyGuard`。

- 实现 status/resolve/voidMarket。

#### 事件

- Bought

- Resolved

- Claimed

- MarketVoided



### 8.1 Status 枚举

#### 枚举说明

- Open/Resolved/Voided。

#### 枚举值

- 同前

---

### 8.2 MultiOutcomeMarket 状态变量

#### 状态变量说明

- 下列为合约显式声明的状态变量；`public` 变量由编译器自动生成同名 getter。

#### 状态变量（每个变量一行）

- `collateral`（IERC20 public immutable）

- `oracle`（address public immutable）

- `matchRef`（bytes32 public immutable）

- `question`（string public）

- `endTime`（uint256 public）

- `outcomeCount`（uint8 public）：结果数 2–8

- `feeBps`（uint16 public）

- `marketStatus`（Status public）

- `winningOutcome`（uint8 public）

- `pool`（uint256[] public）：各 outcome 池（public 带 index getter）

- `stake`（mapping(address=>mapping(uint8=>uint256)) public）

- `claimed`（mapping(address=>bool) public）

---

### 8.3 modifier onlyOracle

- **修饰器用途**

  - 仅 oracle。

- **函数内校验**

  - msg.sender==oracle

- **实现效果**

  - 执行函数体

- **错误返回**

  - `"not oracle"`

---

### 8.4 constructor

- **函数用途**

  - 初始化多结果市场。

- **函数参数说明**

  - `_collateral`

  - `_oracle`

  - `_matchRef`

  - `_question`

  - `_endTime`

  - `_outcomeCount`

  - `_feeBps`

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - _outcomeCount 在 2–8

- **函数实现效果**

  - 初始化 pool 数组长度为 outcomeCount

- **错误返回**

  - `"outcomes"`

---

### 8.5 status

- **函数用途**

  - 返回 marketStatus uint8。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - `uint8`

- **函数内校验**

  - 无

- **函数实现效果**

  - view

- **错误返回**

  - 无

---

### 8.6 buy

- **函数用途**

  - 向指定 outcome 投注。

- **函数参数说明**

  - `outcome`

  - `amount`

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - nonReentrant

  - Open 且未过 endTime

  - outcome<outcomeCount

  - amount>0

- **函数实现效果**

  - 扣 fee；transferFrom；pool 与 stake 累加 net；emit Bought

- **错误返回**

  - `"closed"`/`"invalid"`

---

### 8.7 resolve

- **函数用途**

  - 预言机结算。

- **函数参数说明**

  - `_outcome`

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyOracle

  - Open

  - _outcome<outcomeCount

- **函数实现效果**

  - Resolved；emit

- **错误返回**

  - `"invalid"`

---

### 8.8 voidMarket

- **函数用途**

  - 作废。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyOracle

  - Open

- **函数实现效果**

  - Voided

- **错误返回**

  - `"not open"`

---

### 8.9 claim

- **函数用途**

  - 领取。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - nonReentrant

  - !claimed

  - payout>0

  - Resolved: winPool>0

  - Voided: 各 stake 之和

- **函数实现效果**

  - parimutuel 或全额退；transfer

- **错误返回**

  - `"claimed"`/`"empty"`/`"nothing"`/`"not claimable"`

---

## 9. MarketFactoryV3.sol

#### 合约/模块说明

- Phase 3 工厂：部署 PredictionMarketV3 与 MultiOutcomeMarket。

- Pausable 暂停创建。

#### 规范与约定

- marketTypes: 0=binary v3, 1=multi。

- createBinaryMarket 可选 initialLiquidity seed。

#### 继承关系

- 继承 `Ownable` 与 `Pausable`（多继承）。

#### 事件

- BinaryMarketCreated

- MultiMarketCreated



### 9.1 MarketFactoryV3 状态变量

#### 状态变量说明

- 下列为合约显式声明的状态变量；`public` 变量由编译器自动生成同名 getter。

#### 状态变量（每个变量一行）

- `collateral`（IERC20 public immutable）

- `oracle`（address public）

- `defaultFeeBps`（uint16 public）

- `defaultMaxBet`（uint256 public）：默认 10000*1e6

- `marketCount`（uint256 public）

- `markets`（mapping(uint256=>address) public）

- `marketTypes`（mapping(uint256=>uint8) public）

---

### 9.2 constructor

- **函数用途**

  - 初始化工厂默认值。

- **函数参数说明**

  - `_collateral`

  - `_oracle`

  - `_feeBps`

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - 无

- **函数实现效果**

  - Ownable(msg.sender)

  - defaultMaxBet=10000*1e6

- **错误返回**

  - 无

---

### 9.3 pause

- **函数用途**

  - Owner 暂停工厂。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyOwner

- **函数实现效果**

  - _pause()

- **错误返回**

  - 非 Owner revert

---

### 9.4 unpause

- **函数用途**

  - Owner 恢复。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyOwner

- **函数实现效果**

  - _unpause()

- **错误返回**

  - 非 Owner revert

---

### 9.5 setOracle

- **函数用途**

  - 更新 oracle。

- **函数参数说明**

  - `_oracle`

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyOwner

- **函数实现效果**

  - oracle=_oracle

- **错误返回**

  - 非 Owner revert

---

### 9.6 setDefaultFeeBps

- **函数用途**

  - 更新默认 fee。

- **函数参数说明**

  - `bps`（uint16）

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - onlyOwner

- **函数实现效果**

  - defaultFeeBps=bps

- **错误返回**

  - 非 Owner revert

---

### 9.7 createBinaryMarket

- **函数用途**

  - 部署 V3 二元市场。

- **函数参数说明**

  - `matchRef`

  - `question`

  - `endTime`

  - `initialLiquidity`

- **函数返回参数说明**

  - `market`（address）

  - `id`（uint256）

- **函数内校验**

  - onlyOwner

  - whenNotPaused

- **函数实现效果**

  - initialLiquidity>0 时 transferFrom Owner 2*liquidity

  - new PredictionMarketV3(..., initialLiquidity=0 构造)

  - transfer 至 market 并 seedReserves

  - marketCount++；marketTypes=0

  - emit BinaryMarketCreated

- **错误返回**

  - pause/transfer/seed revert

---

### 9.8 createMultiMarket

- **函数用途**

  - 部署多结果市场。

- **函数参数说明**

  - `matchRef`

  - `question`

  - `endTime`

  - `outcomeCount`

- **函数返回参数说明**

  - `market`

  - `id`

- **函数内校验**

  - onlyOwner

  - whenNotPaused

- **函数实现效果**

  - new MultiOutcomeMarket

  - marketTypes=1

  - emit MultiMarketCreated

- **错误返回**

  - 构造 outcome 校验 revert

---

### 9.9 version

- **函数用途**

  - 版本字符串。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - `string`："3.0.0-phase3"

- **函数内校验**

  - 无

- **函数实现效果**

  - pure

- **错误返回**

  - 无

---

## 10. DIDRegistry.sol

#### 合约/模块说明

- 可选链上 DID 哈希绑定：用户 EIP-191 签名证明后写入 didHashOf。

#### 规范与约定

- digest = keccak256("BindDID:", msg.sender, didHash) 的 eth signed message。

#### 继承关系

- 继承 OpenZeppelin `Ownable`。

- 使用 ECDSA + MessageHashUtils。

#### 事件

- `DidBound(account, didHash)`



### 10.1 DIDRegistry 状态变量

#### 状态变量说明

- 下列为合约显式声明的状态变量；`public` 变量由编译器自动生成同名 getter。

#### 状态变量（每个变量一行）

- `didHashOf`（mapping(address=>bytes32) public）：账户→DID 内容 Keccak256 哈希

---

### 10.2 constructor

- **函数用途**

  - 部署注册表，部署者为 Owner。

- **函数参数说明**

  - 无参数。

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - 无

- **函数实现效果**

  - Ownable(msg.sender)

- **错误返回**

  - 无

---

### 10.3 bindDid

- **函数用途**

  - 用户绑定 DID 哈希。

- **函数参数说明**

  - `didHash`（bytes32）：DID 内容哈希，不可为零

  - `signature`（bytes calldata）：对 digest 的 ECDSA 签名

- **函数返回参数说明**

  - 无返回值。

- **函数内校验**

  - didHash != 0

  - recover(signature) == msg.sender

- **函数实现效果**

  - didHashOf[msg.sender]=didHash；emit DidBound

- **错误返回**

  - `"empty did"`/`"invalid sig"`

---

### 10.4 resolveDid

- **函数用途**

  - 查询账户绑定的 didHash。

- **函数参数说明**

  - `account`（address）

- **函数返回参数说明**

  - `bytes32`：didHash，未绑定为 0

- **函数内校验**

  - 无

- **函数实现效果**

  - view 读 mapping

- **错误返回**

  - 无

---

## 附录：编译器自动生成的 public getter



下列 `public` 状态变量除上文注释外，Solidity 自动生成同名 view 函数（mapping 需传入 key/index）：



1. **PredictionMarket**：collateral、oracle、factory、matchRef、question、endTime、status、winningOutcome、yesPool、noPool、yesBalance(addr)、noBalance(addr)、claimed(addr)



2. **MarketFactory**：collateral、oracle、marketCount、markets(id)



3. **OracleAdapter**：ORACLE_ROLE、timelockDelay、factory、pending(market)



4. **OracleAdapterV2**：ORACLE_ROLE、threshold、proposalCount、proposals(id)、approved(id,oracle)



5. **PredictionMarketV3**：collateral、oracle、factory、matchRef、question、endTime、feeBps、maxBetPerUser、marketStatus、winningOutcome、reserveYes、reserveNo、totalLPSupply、collectedFees、yesBalance、noBalance、lpBalance、userBetTotal、claimed



6. **MultiOutcomeMarket**：collateral、oracle、matchRef、question、endTime、outcomeCount、feeBps、marketStatus、winningOutcome、pool(i)、stake(user,outcome)、claimed(user)



7. **MarketFactoryV3**：collateral、oracle、defaultFeeBps、defaultMaxBet、marketCount、markets(id)、marketTypes(id)



8. **DIDRegistry**：didHashOf(account)



9. **MockUSDC**（ERC20）：name、symbol、decimals、totalSupply、balanceOf、allowance 等标准 ERC20 getter
