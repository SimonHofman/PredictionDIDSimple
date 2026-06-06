package indexer

import (
	"context"
	"encoding/json"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prediction-did/simple/internal/config"
	"github.com/prediction-did/simple/internal/models"
	"github.com/prediction-did/simple/internal/repository"
)

type Indexer struct {
	cfg       *config.Config
	client    *ethclient.Client
	factoryABI abi.ABI
	marketABI  abi.ABI
	matches    *repository.MatchRepo
	markets    *repository.MarketRepo
	positions  *repository.PositionRepo
	state      *repository.IndexerStateRepo
}

func New(ctx context.Context, cfg *config.Config, matches *repository.MatchRepo, markets *repository.MarketRepo, positions *repository.PositionRepo, state *repository.IndexerStateRepo) (*Indexer, error) {
	client, err := ethclient.DialContext(ctx, cfg.EthRPCURL)
	if err != nil {
		return nil, err
	}
	factoryABI, err := loadABI("MarketFactory")
	if err != nil {
		return nil, err
	}
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

func loadABI(name string) (abi.ABI, error) {
	paths := []string{
		filepath.Join("pkg", "contracts", name+".json"),
		filepath.Join("backend", "pkg", "contracts", name+".json"),
	}
	var raw []byte
	var err error
	for _, p := range paths {
		raw, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		return abi.ABI{}, err
	}
	var wrapper struct {
		ABI json.RawMessage `json:"abi"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return abi.ABI{}, err
	}
	return abi.JSON(strings.NewReader(string(wrapper.ABI)))
}

func (idx *Indexer) Run(ctx context.Context) error {
	if idx.cfg.FactoryAddress == "" {
		log.Println("indexer: MARKET_FACTORY_ADDRESS not set, skipping")
		return nil
	}
	factoryAddr := common.HexToAddress(idx.cfg.FactoryAddress)
	ticker := time.NewTicker(time.Duration(idx.cfg.IndexerPollSeconds) * time.Second)
	defer ticker.Stop()

	for {
		if err := idx.poll(ctx, factoryAddr); err != nil {
			log.Printf("indexer poll error: %v", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (idx *Indexer) poll(ctx context.Context, factoryAddr common.Address) error {
	last, err := idx.state.GetLastBlock(ctx)
	if err != nil {
		return err
	}
	if last == 0 && idx.cfg.IndexerStartBlock > 0 {
		last = idx.cfg.IndexerStartBlock - 1
	}

	head, err := idx.client.BlockNumber(ctx)
	if err != nil {
		return err
	}
	if last >= head {
		return nil
	}
	from := last + 1
	to := head
	if to-from > 2000 {
		to = from + 2000
	}

	if err := idx.scanFactory(ctx, factoryAddr, from, to); err != nil {
		return err
	}
	if err := idx.scanMarkets(ctx, from, to); err != nil {
		return err
	}
	return idx.state.SetLastBlock(ctx, to, idx.cfg.FactoryAddress)
}

func (idx *Indexer) scanFactory(ctx context.Context, factory common.Address, from, to uint64) error {
	createdSig := idx.factoryABI.Events["MarketCreated"].ID
	q := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(from)),
		ToBlock:   big.NewInt(int64(to)),
		Addresses: []common.Address{factory},
		Topics:    [][]common.Hash{{createdSig}},
	}
	logs, err := idx.client.FilterLogs(ctx, q)
	if err != nil {
		return err
	}
	for _, lg := range logs {
		if err := idx.handleMarketCreated(ctx, lg); err != nil {
			log.Printf("MarketCreated: %v", err)
		}
	}
	return nil
}

func (idx *Indexer) handleMarketCreated(ctx context.Context, lg types.Log) error {
	if len(lg.Topics) < 4 {
		return nil
	}
	marketID := new(big.Int).SetBytes(lg.Topics[1].Bytes())
	marketAddr := common.BytesToAddress(lg.Topics[2].Bytes())
	matchRef := lg.Topics[3]

	vals, err := idx.factoryABI.Unpack("MarketCreated", lg.Data)
	if err != nil {
		return err
	}
	question := vals[0].(string)
	endTime := vals[1].(*big.Int)

	matchRefHex := matchRef.Hex()
	externalID := matchRefToExternal(matchRefHex)

	var matchID *int64
	if m, err := idx.matches.GetByExternalID(ctx, externalID); err == nil {
		matchID = &m.ID
	}

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

func matchRefToExternal(matchRefHex string) string {
	known := []struct {
		external string
	}{
		{"wc-2026-semi-001"},
		{"wc-2026-semi-002"},
		{"wc-2026-final-001"},
	}
	for _, k := range known {
		h := crypto.Keccak256Hash([]byte(k.external))
		if strings.EqualFold(h.Hex(), matchRefHex) {
			return k.external
		}
	}
	return matchRefHex
}

func (idx *Indexer) scanMarkets(ctx context.Context, from, to uint64) error {
	rows, err := idx.markets.List(ctx, "", 500, 0)
	if err != nil {
		return err
	}
	boughtSig := idx.marketABI.Events["Bought"].ID
	resolvedSig := idx.marketABI.Events["Resolved"].ID
	claimedSig := idx.marketABI.Events["Claimed"].ID

	for _, mk := range rows {
		addr := common.HexToAddress(mk.MarketAddress)
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
		for _, lg := range logs {
			switch lg.Topics[0] {
			case boughtSig:
				_ = idx.handleBought(ctx, mk.ID, lg)
			case resolvedSig:
				_ = idx.handleResolved(ctx, mk.MarketAddress, lg)
			case claimedSig:
				_ = idx.handleClaimed(ctx, mk.ID, lg)
			}
		}
	}
	return nil
}

func (idx *Indexer) handleBought(ctx context.Context, marketID int64, lg types.Log) error {
	vals, err := idx.marketABI.Unpack("Bought", lg.Data)
	if err != nil {
		return err
	}
	outcome := vals[0].(uint8)
	amount := vals[1].(*big.Int)
	user := common.BytesToAddress(lg.Topics[1].Bytes())
	amt := amount.String()
	if err := idx.positions.InsertTrade(ctx, marketID, lg.TxHash.Hex(), int(lg.Index), int64(lg.BlockNumber), user.Hex(), int(outcome), amt); err != nil {
		return err
	}
	return idx.positions.AddTrade(ctx, marketID, user.Hex(), int(outcome), amt)
}

func (idx *Indexer) handleResolved(ctx context.Context, marketAddress string, lg types.Log) error {
	vals, err := idx.marketABI.Unpack("Resolved", lg.Data)
	if err != nil {
		return err
	}
	outcome := int(vals[0].(uint8))
	return idx.markets.UpdateResolved(ctx, marketAddress, outcome, "0", "0")
}

func (idx *Indexer) handleClaimed(ctx context.Context, marketID int64, lg types.Log) error {
	user := common.BytesToAddress(lg.Topics[1].Bytes())
	return idx.positions.SetClaimed(ctx, marketID, user.Hex())
}
