package blockchain

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type OracleClient struct {
	client  *ethclient.Client
	adapter common.Address
	abi     abi.ABI
	chainID *big.Int
	auth    *bind.TransactOpts
}

func NewOracleClient(ctx context.Context, rpcURL, adapterAddr, privateKeyHex string, chainID int64) (*OracleClient, error) {
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, err
	}
	ab, err := loadAdapterABI()
	if err != nil {
		return nil, err
	}
	pk, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, err
	}
	auth, err := bindOpts(client, pk, chainID)
	if err != nil {
		return nil, err
	}
	return &OracleClient{
		client:  client,
		adapter: common.HexToAddress(adapterAddr),
		abi:     ab,
		chainID: big.NewInt(chainID),
		auth:    auth,
	}, nil
}

func bindOpts(client *ethclient.Client, pk *ecdsa.PrivateKey, chainID int64) (*bind.TransactOpts, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(chainID))
	if err != nil {
		return nil, err
	}
	auth.Context = context.Background()
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err == nil {
		auth.GasPrice = gasPrice
	}
	return auth, nil
}

func loadAdapterABI() (abi.ABI, error) {
	paths := []string{
		filepath.Join("pkg", "contracts", "OracleAdapter.json"),
		filepath.Join("backend", "pkg", "contracts", "OracleAdapter.json"),
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
	var w struct {
		ABI json.RawMessage `json:"abi"`
	}
	if err := json.Unmarshal(raw, &w); err != nil {
		return abi.ABI{}, err
	}
	return abi.JSON(strings.NewReader(string(w.ABI)))
}

func (c *OracleClient) RequestResolve(ctx context.Context, market string, outcome uint8) (string, error) {
	return c.transact(ctx, "requestResolve", common.HexToAddress(market), outcome)
}

func (c *OracleClient) ConfirmResolve(ctx context.Context, market string) (string, error) {
	return c.transact(ctx, "confirmResolve", common.HexToAddress(market))
}

func (c *OracleClient) ResolveNow(ctx context.Context, market string, outcome uint8) (string, error) {
	return c.transact(ctx, "resolveNow", common.HexToAddress(market), outcome)
}

func (c *OracleClient) VoidMarket(ctx context.Context, market string) (string, error) {
	return c.transact(ctx, "voidMarket", common.HexToAddress(market))
}

func (c *OracleClient) transact(ctx context.Context, method string, args ...interface{}) (string, error) {
	data, err := c.abi.Pack(method, args...)
	if err != nil {
		return "", err
	}
	nonce, err := c.client.PendingNonceAt(ctx, c.auth.From)
	if err != nil {
		return "", err
	}
	gas, err := c.client.EstimateGas(ctx, ethereum.CallMsg{
		From: c.auth.From,
		To:   &c.adapter,
		Data: data,
	})
	if err != nil {
		gas = 500000
	}
	tx := types.NewTransaction(nonce, c.adapter, big.NewInt(0), gas, c.auth.GasPrice, data)
	signed, err := c.auth.Signer(c.auth.From, tx)
	if err != nil {
		return "", err
	}
	if err := c.client.SendTransaction(ctx, signed); err != nil {
		return "", err
	}
	return signed.Hash().Hex(), nil
}

func (c *OracleClient) WaitMined(ctx context.Context, txHash string) error {
	hash := common.HexToHash(txHash)
	for i := 0; i < 30; i++ {
		receipt, err := c.client.TransactionReceipt(ctx, hash)
		if err == nil && receipt != nil {
			if receipt.Status == 0 {
				return fmt.Errorf("tx reverted")
			}
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return fmt.Errorf("timeout waiting for tx")
}
