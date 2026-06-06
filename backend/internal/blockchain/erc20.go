// Package blockchain ERC20 余额查询工具
package blockchain

// 导入依赖
import (
	"context"  // 上下文
	"math/big" // 大整数

	"github.com/ethereum/go-ethereum"           // 主包（CallMsg）
	"github.com/ethereum/go-ethereum/common"    // 地址等通用类型
	"github.com/ethereum/go-ethereum/ethclient" // 客户端
)

// balanceOf(address) selector
// balanceOf(address) 函数选择器（ERC20 标准：keccak256("balanceOf(address)")[0:4]）
var balanceOfData = common.FromHex("0x70a08231")

// ERC20Balance 查询某账户在某 ERC20 代币合约的余额
func ERC20Balance(ctx context.Context, rpcURL, tokenAddr, holder string) (*big.Int, error) {
	// 拨号 RPC
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, err
	}
	defer client.Close() // 用完关闭

	token := common.HexToAddress(tokenAddr)   // 代币合约地址
	holderAddr := common.HexToAddress(holder) // 持有者地址
	// 拼接 calldata：selector + 32 字节地址参数
	data := append(balanceOfData, common.LeftPadBytes(holderAddr.Bytes(), 32)...)

	// eth_call 静态调用合约
	result, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &token, // 调用目标
		Data: data,   // 函数调用数据
	}, nil)
	if err != nil {
		return nil, err
	}
	// 解析返回值（uint256）
	return new(big.Int).SetBytes(result), nil
}
