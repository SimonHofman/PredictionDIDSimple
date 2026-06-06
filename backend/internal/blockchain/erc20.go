package blockchain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
)

// balanceOf(address) selector
var balanceOfData = common.FromHex("0x70a08231")

func ERC20Balance(ctx context.Context, rpcURL, tokenAddr, holder string) (*big.Int, error) {
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	token := common.HexToAddress(tokenAddr)
	holderAddr := common.HexToAddress(holder)
	data := append(balanceOfData, common.LeftPadBytes(holderAddr.Bytes(), 32)...)

	result, err := client.CallContract(ctx, ethereum.CallMsg{
		To:   &token,
		Data: data,
	}, nil)
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(result), nil
}
