// Package blockchain Oracle 客户端，用于向 OracleAdapter 合约写入解析结果
package blockchain

// 导入依赖
import (
	"context"       // 上下文
	"crypto/ecdsa"  // ECDSA 私钥
	"encoding/json" // JSON 解析
	"fmt"           // 错误格式化
	"math/big"      // 大整数
	"os"            // 文件读取
	"path/filepath" // 路径
	"strings"       // 字符串
	"time"          // 时间

	"github.com/ethereum/go-ethereum"                   // 主包
	"github.com/ethereum/go-ethereum/accounts/abi"      // ABI 编解码
	"github.com/ethereum/go-ethereum/accounts/abi/bind" // 交易构建
	"github.com/ethereum/go-ethereum/common"            // 通用类型
	"github.com/ethereum/go-ethereum/core/types"        // 交易类型
	"github.com/ethereum/go-ethereum/crypto"            // 密钥工具
	"github.com/ethereum/go-ethereum/ethclient"         // 以太坊客户端
)

// OracleClient 与 OracleAdapter 合约交互的客户端
type OracleClient struct {
	client  *ethclient.Client  // RPC 客户端
	adapter common.Address     // OracleAdapter 合约地址
	abi     abi.ABI            // ABI 解析器
	chainID *big.Int           // 链 ID
	auth    *bind.TransactOpts // 签名/发送参数
}

// NewOracleClient 创建 Oracle 客户端
func NewOracleClient(ctx context.Context, rpcURL, adapterAddr, privateKeyHex string, chainID int64) (*OracleClient, error) {
	// 连接 RPC
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, err
	}
	// 加载 Adapter ABI
	ab, err := loadAdapterABI()
	if err != nil {
		return nil, err
	}
	// 解析私钥（兼容 0x 前缀）
	pk, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, err
	}
	// 构造交易选项（带签名）
	auth, err := bindOpts(client, pk, chainID)
	if err != nil {
		return nil, err
	}
	// 组装客户端结构
	return &OracleClient{
		client:  client,
		adapter: common.HexToAddress(adapterAddr),
		abi:     ab,
		chainID: big.NewInt(chainID),
		auth:    auth,
	}, nil
}

// bindOpts 创建带链 ID 的交易选项，并尝试设置 gasPrice
func bindOpts(client *ethclient.Client, pk *ecdsa.PrivateKey, chainID int64) (*bind.TransactOpts, error) {
	// 创建键控签名器
	auth, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(chainID))
	if err != nil {
		return nil, err
	}
	auth.Context = context.Background() // 默认上下文
	// 查询建议 gasPrice，失败则保持默认
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err == nil {
		auth.GasPrice = gasPrice
	}
	return auth, nil
}

// loadAdapterABI 从本地 contracts 目录加载 OracleAdapter 的 ABI
func loadAdapterABI() (abi.ABI, error) {
	// 兼容多种运行目录
	paths := []string{
		filepath.Join("pkg", "contracts", "OracleAdapter.json"),
		filepath.Join("backend", "pkg", "contracts", "OracleAdapter.json"),
	}
	var raw []byte
	var err error
	// 逐个尝试读取
	for _, p := range paths {
		raw, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}
	if err != nil {
		return abi.ABI{}, err
	}
	// 取 JSON 中的 abi 字段
	var w struct {
		ABI json.RawMessage `json:"abi"`
	}
	if err := json.Unmarshal(raw, &w); err != nil {
		return abi.ABI{}, err
	}
	// 解析 ABI
	return abi.JSON(strings.NewReader(string(w.ABI)))
}

// RequestResolve 发起结算请求（提议结果）
func (c *OracleClient) RequestResolve(ctx context.Context, market string, outcome uint8) (string, error) {
	return c.transact(ctx, "requestResolve", common.HexToAddress(market), outcome)
}

// ConfirmResolve 确认之前请求过的结算
func (c *OracleClient) ConfirmResolve(ctx context.Context, market string) (string, error) {
	return c.transact(ctx, "confirmResolve", common.HexToAddress(market))
}

// ResolveNow 立即结算（管理员路径）
func (c *OracleClient) ResolveNow(ctx context.Context, market string, outcome uint8) (string, error) {
	return c.transact(ctx, "resolveNow", common.HexToAddress(market), outcome)
}

// VoidMarket 作废市场
func (c *OracleClient) VoidMarket(ctx context.Context, market string) (string, error) {
	return c.transact(ctx, "voidMarket", common.HexToAddress(market))
}

// transact 通用方法：打包参数 -> 估算 gas -> 签名 -> 发送
func (c *OracleClient) transact(ctx context.Context, method string, args ...interface{}) (string, error) {
	// 编码 calldata
	data, err := c.abi.Pack(method, args...)
	if err != nil {
		return "", err
	}
	// 当前 nonce
	nonce, err := c.client.PendingNonceAt(ctx, c.auth.From)
	if err != nil {
		return "", err
	}
	// 估算 gas
	gas, err := c.client.EstimateGas(ctx, ethereum.CallMsg{
		From: c.auth.From,
		To:   &c.adapter,
		Data: data,
	})
	if err != nil {
		// 估算失败给一个保守值
		gas = 500000
	}
	// 构造交易（legacy tx）
	tx := types.NewTransaction(nonce, c.adapter, big.NewInt(0), gas, c.auth.GasPrice, data)
	// 签名
	signed, err := c.auth.Signer(c.auth.From, tx)
	if err != nil {
		return "", err
	}
	// 发送到链上
	if err := c.client.SendTransaction(ctx, signed); err != nil {
		return "", err
	}
	// 返回交易哈希
	return signed.Hash().Hex(), nil
}

// WaitMined 轮询等待交易上链（最多 30 次，每次 2s）
func (c *OracleClient) WaitMined(ctx context.Context, txHash string) error {
	hash := common.HexToHash(txHash) // 解析 hash
	// 最多 30 次轮询
	for i := 0; i < 30; i++ {
		// 查询交易回执
		receipt, err := c.client.TransactionReceipt(ctx, hash)
		if err == nil && receipt != nil {
			// 状态 0 表示 revert
			if receipt.Status == 0 {
				return fmt.Errorf("tx reverted")
			}
			return nil // 成功
		}
		// 等待 2 秒或上下文取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	// 超时
	return fmt.Errorf("timeout waiting for tx")
}
