package blockchain

import (
	"context"
	"log"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	url         string
	expectedID  int64
	rpcOK       atomic.Bool
	chainID     atomic.Int64
}

func New(url string, expectedChainID int64) *Client {
	return &Client{
		url:        url,
		expectedID: expectedChainID,
	}
}

func (c *Client) StartBackgroundPing(ctx context.Context) {
	go func() {
		c.pingOnce(ctx)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.pingOnce(ctx)
			}
		}
	}()
}

func (c *Client) pingOnce(ctx context.Context) {
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(pingCtx, c.url)
	if err != nil {
		c.rpcOK.Store(false)
		log.Printf("WARN: RPC dial failed (%s): %v", c.url, err)
		return
	}
	defer client.Close()

	chainID, err := client.ChainID(pingCtx)
	if err != nil {
		c.rpcOK.Store(false)
		log.Printf("WARN: RPC ChainID failed: %v", err)
		return
	}

	c.chainID.Store(chainID.Int64())
	if c.expectedID > 0 && chainID.Int64() != c.expectedID {
		c.rpcOK.Store(false)
		log.Printf("WARN: RPC chainId %d != expected %d", chainID.Int64(), c.expectedID)
		return
	}

	c.rpcOK.Store(true)
	log.Printf("RPC ok: %s chainId=%d", c.url, chainID.Int64())
}

func (c *Client) RPCOK() bool {
	return c.rpcOK.Load()
}

func (c *Client) ChainID() int64 {
	return c.chainID.Load()
}
