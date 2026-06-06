// Package indexer 链上事件索引器，轮询区块日志并同步到数据库
package indexer

// 导入依赖
import (
	"context"       // 上下文
	"encoding/json" // JSON 解析
	"log"           // 日志
	"math/big"      // 大整数
	"os"            // 文件读取
	"path/filepath" // 路径
	"strings"       // 字符串处理
	"time"          // 时间

	"github.com/ethereum/go-ethereum"                      // 过滤查询
	"github.com/ethereum/go-ethereum/accounts/abi"         // ABI 解析
	"github.com/ethereum/go-ethereum/common"               // 通用类型
	"github.com/ethereum/go-ethereum/core/types"           // 日志类型
	"github.com/ethereum/go-ethereum/crypto"               // 哈希工具
	"github.com/ethereum/go-ethereum/ethclient"            // 客户端
	"github.com/prediction-did/simple/internal/config"     // 配置
	"github.com/prediction-did/simple/internal/models"     // 模型
	"github.com/prediction-did/simple/internal/repository" // 仓储
)

// Indexer 索引器核心结构体
type Indexer struct {
	cfg        *config.Config               // 配置
	client     *ethclient.Client            // RPC 客户端
	factoryABI abi.ABI                      // Factory 合约 ABI
	marketABI  abi.ABI                      // Market 合约 ABI
	matches    *repository.MatchRepo        // 比赛仓储
	markets    *repository.MarketRepo       // 市场仓储
	positions  *repository.PositionRepo     // 持仓仓储
	state      *repository.IndexerStateRepo // 索引器状态仓储
}

// New 创建索引器实例
func New(ctx context.Context, cfg *config.Config, matches *repository.MatchRepo, markets *repository.MarketRepo, positions *repository.PositionRepo, state *repository.IndexerStateRepo) (*Indexer, error) {
	// 连接以太坊节点
	client, err := ethclient.DialContext(ctx, cfg.EthRPCURL)
	if err != nil {
		return nil, err
	}
	// 加载 MarketFactory ABI
	factoryABI, err := loadABI("MarketFactory")
	if err != nil {
		return nil, err
	}
	// 加载 PredictionMarket ABI
	marketABI, err := loadABI("PredictionMarket")
	if err != nil {
		return nil, err
	}
	return &Indexer{
		cfg:        cfg,
		client:     client,
		factoryABI: factoryABI,
		marketABI:  marketABI,
		matches:    matches,
		markets:    markets,
		positions:  positions,
		state:      state,
	}, nil
}

// loadABI 从 pkg/contracts 目录加载合约 ABI
func loadABI(name string) (abi.ABI, error) {
	// 候选路径
	paths := []string{
		filepath.Join("pkg", "contracts", name+".json"),
		filepath.Join("backend", "pkg", "contracts", name+".json"),
	}
	var raw []byte
	var err error
	// 逐一尝试读取
	for _, p := range paths {
		raw, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		return abi.ABI{}, err
	}
	// 提取 abi 字段
	var wrapper struct {
		ABI json.RawMessage `json:"abi"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return abi.ABI{}, err
	}
	// 解析 ABI JSON
	return abi.JSON(strings.NewReader(string(wrapper.ABI)))
}

// Run 启动索引器主循环
func (idx *Indexer) Run(ctx context.Context) error {
	// 未配置 Factory 地址时跳过
	if idx.cfg.FactoryAddress == "" {
		log.Println("indexer: MARKET_FACTORY_ADDRESS not set, skipping")
		return nil
	}
	factoryAddr := common.HexToAddress(idx.cfg.FactoryAddress) // Factory 合约地址
	// 创建轮询定时器
	ticker := time.NewTicker(time.Duration(idx.cfg.IndexerPollSeconds) * time.Second)
	defer ticker.Stop()

	for {
		// 执行一次区块扫描
		if err := idx.poll(ctx, factoryAddr); err != nil {
			log.Printf("indexer poll error: %v", err)
		}
		select {
		case <-ctx.Done(): // 收到取消信号
			return ctx.Err()
		case <-ticker.C: // 等待下次轮询
		}
	}
}

// poll 执行一次区块范围扫描
func (idx *Indexer) poll(ctx context.Context, factoryAddr common.Address) error {
	// 获取上次扫描到的区块号
	last, err := idx.state.GetLastBlock(ctx)
	if err != nil {
		return err
	}
	// 首次运行使用配置的起始区块
	if last == 0 && idx.cfg.IndexerStartBlock > 0 {
		last = idx.cfg.IndexerStartBlock - 1
	}

	// 当前链上最新区块
	head, err := idx.client.BlockNumber(ctx)
	if err != nil {
		return err
	}
	// 没有新区块则跳过
	if last >= head {
		return nil
	}
	from := last + 1 // 起始
	to := head       // 结束
	// 每次最多扫 2000 个区块避免 RPC 超时
	if to-from > 2000 {
		to = from + 2000
	}

	// 扫描 Factory 日志
	if err := idx.scanFactory(ctx, factoryAddr, from, to); err != nil {
		return err
	}
	// 扫描各市场日志
	if err := idx.scanMarkets(ctx, from, to); err != nil {
		return err
	}
	// 更新已处理区块号
	return idx.state.SetLastBlock(ctx, to, idx.cfg.FactoryAddress)
}

// scanFactory 扫描 Factory 合约的 MarketCreated 事件
func (idx *Indexer) scanFactory(ctx context.Context, factory common.Address, from, to uint64) error {
	createdSig := idx.factoryABI.Events["MarketCreated"].ID // 事件签名
	// 构造过滤查询
	q := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(from)),
		ToBlock:   big.NewInt(int64(to)),
		Addresses: []common.Address{factory},
		Topics:    [][]common.Hash{{createdSig}},
	}
	// 查询日志
	logs, err := idx.client.FilterLogs(ctx, q)
	if err != nil {
		return err
	}
	// 逐条处理
	for _, lg := range logs {
		if err := idx.handleMarketCreated(ctx, lg); err != nil {
			log.Printf("MarketCreated: %v", err)
		}
	}
	return nil
}

// handleMarketCreated 处理单条 MarketCreated 日志
func (idx *Indexer) handleMarketCreated(ctx context.Context, lg types.Log) error {
	// 需要至少 4 个 topic
	if len(lg.Topics) < 4 {
		return nil
	}
	marketID := new(big.Int).SetBytes(lg.Topics[1].Bytes())   // 链上市场 ID
	marketAddr := common.BytesToAddress(lg.Topics[2].Bytes()) // 市场合约地址
	matchRef := lg.Topics[3]                                  // 比赛引用哈希

	// 解析非索引参数
	vals, err := idx.factoryABI.Unpack("MarketCreated", lg.Data)
	if err != nil {
		return err
	}
	question := vals[0].(string)  // 预测问题
	endTime := vals[1].(*big.Int) // 结束时间

	matchRefHex := matchRef.Hex()
	// 尝试反查 external_id
	externalID := matchRefToExternal(matchRefHex)

	// 查找对应比赛
	var matchID *int64
	if m, err := idx.matches.GetByExternalID(ctx, externalID); err == nil {
		matchID = &m.ID
	}

	// 组装市场模型并入库
	mk := models.Market{
		MatchID:         matchID,
		ChainID:         idx.cfg.ChainID,
		FactoryAddress:  strings.ToLower(idx.cfg.FactoryAddress),
		MarketAddress:   strings.ToLower(marketAddr.Hex()),
		OnChainMarketID: marketID.Int64(),
		MatchRef:        matchRefHex,
		Question:        question,
		EndTime:         time.Unix(endTime.Int64(), 0),
		Status:          "OPEN",
	}
	_, err = idx.markets.InsertFromChain(ctx, mk)
	return err
}

// matchRefToExternal 将 matchRef 哈希反查为可读 external_id
func matchRefToExternal(matchRefHex string) string {
	// 已知的 external_id 集合
	known := []struct {
		external string
	}{
		{"wc-2026-semi-001"},
		{"wc-2026-semi-002"},
		{"wc-2026-final-001"},
	}
	// 逐个计算 keccak256 并比对
	for _, k := range known {
		h := crypto.Keccak256Hash([]byte(k.external))
		if strings.EqualFold(h.Hex(), matchRefHex) {
			return k.external
		}
	}
	// 未匹配时返回原始十六进制值
	return matchRefHex
}

// scanMarkets 扫描所有已知市场的 Bought/Resolved/Claimed 事件
func (idx *Indexer) scanMarkets(ctx context.Context, from, to uint64) error {
	// 查出所有市场
	rows, err := idx.markets.List(ctx, "", 500, 0)
	if err != nil {
		return err
	}
	// 各事件签名
	boughtSig := idx.marketABI.Events["Bought"].ID
	resolvedSig := idx.marketABI.Events["Resolved"].ID
	claimedSig := idx.marketABI.Events["Claimed"].ID

	// 遍历每个市场
	for _, mk := range rows {
		addr := common.HexToAddress(mk.MarketAddress)
		// 过滤查询
		q := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(from)),
			ToBlock:   big.NewInt(int64(to)),
			Addresses: []common.Address{addr},
			Topics:    [][]common.Hash{{boughtSig, resolvedSig, claimedSig}},
		}
		logs, err := idx.client.FilterLogs(ctx, q)
		if err != nil {
			return err
		}
		// 按事件类型分发处理
		for _, lg := range logs {
			switch lg.Topics[0] {
			case boughtSig:
				_ = idx.handleBought(ctx, mk.ID, lg) // 处理买入
			case resolvedSig:
				_ = idx.handleResolved(ctx, mk.MarketAddress, lg) // 处理结算
			case claimedSig:
				_ = idx.handleClaimed(ctx, mk.ID, lg) // 处理领取
			}
		}
	}
	return nil
}

// handleBought 处理 Bought 事件（用户买入）
func (idx *Indexer) handleBought(ctx context.Context, marketID int64, lg types.Log) error {
	// 解析事件参数
	vals, err := idx.marketABI.Unpack("Bought", lg.Data)
	if err != nil {
		return err
	}
	outcome := vals[0].(uint8)                          // 结果索引
	amount := vals[1].(*big.Int)                        // 买入金额
	user := common.BytesToAddress(lg.Topics[1].Bytes()) // 买入者地址
	amt := amount.String()                              // 字符串金额
	// 插入交易记录
	if err := idx.positions.InsertTrade(ctx, marketID, lg.TxHash.Hex(), int(lg.Index), int64(lg.BlockNumber), user.Hex(), int(outcome), amt); err != nil {
		return err
	}
	// 更新持仓聚合
	return idx.positions.AddTrade(ctx, marketID, user.Hex(), int(outcome), amt)
}

// handleResolved 处理 Resolved 事件（市场结算）
func (idx *Indexer) handleResolved(ctx context.Context, marketAddress string, lg types.Log) error {
	// 解析结算结果
	vals, err := idx.marketABI.Unpack("Resolved", lg.Data)
	if err != nil {
		return err
	}
	outcome := int(vals[0].(uint8)) // 获胜结果
	// 更新市场为已结算
	return idx.markets.UpdateResolved(ctx, marketAddress, outcome, "0", "0")
}

// handleClaimed 处理 Claimed 事件（用户领取奖金）
func (idx *Indexer) handleClaimed(ctx context.Context, marketID int64, lg types.Log) error {
	user := common.BytesToAddress(lg.Topics[1].Bytes()) // 领取用户
	// 标记为已领取
	return idx.positions.SetClaimed(ctx, marketID, user.Hex())
}
