// Package blockchain 区块链交互层（节点连接、ERC20 查询、Oracle 写入等）
package blockchain

// 导入依赖
import (
	"context"     // 上下文
	"log"         // 日志
	"sync/atomic" // 原子值，用于无锁状态共享
	"time"        // 时间

	"github.com/ethereum/go-ethereum/ethclient" // 以太坊客户端
)

// Client 简单的区块链状态客户端，定时 ping 节点
type Client struct {
	url        string       // RPC URL
	expectedID int64        // 期望的 chainId（与 RPC 实际值校验）
	rpcOK      atomic.Bool  // RPC 是否健康
	chainID    atomic.Int64 // 实际 chainId
}

// New 创建 Client
func New(url string, expectedChainID int64) *Client {
	return &Client{
		url:        url,             // 节点地址
		expectedID: expectedChainID, // 期望 chainId
	}
}

// StartBackgroundPing 在后台周期性 ping 节点检查连通性
func (c *Client) StartBackgroundPing(ctx context.Context) {
	go func() {
		c.pingOnce(ctx) // 启动时立即 ping 一次
		// 每 30 秒一次
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done(): // 上下文取消则退出
				return
			case <-ticker.C: // 定时触发
				c.pingOnce(ctx)
			}
		}
	}()
}

// pingOnce 执行一次 ping
func (c *Client) pingOnce(ctx context.Context) {
	// 5 秒超时
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 拨号
	client, err := ethclient.DialContext(pingCtx, c.url)
	if err != nil {
		c.rpcOK.Store(false) // 失败标记不健康
		log.Printf("WARN: RPC dial failed (%s): %v", c.url, err)
		return
	}
	defer client.Close() // 用完关闭

	// 查询 chainId
	chainID, err := client.ChainID(pingCtx)
	if err != nil {
		c.rpcOK.Store(false)
		log.Printf("WARN: RPC ChainID failed: %v", err)
		return
	}

	// 保存实际 chainId
	c.chainID.Store(chainID.Int64())
	// 与期望 chainId 不一致也算不健康
	if c.expectedID > 0 && chainID.Int64() != c.expectedID {
		c.rpcOK.Store(false)
		log.Printf("WARN: RPC chainId %d != expected %d", chainID.Int64(), c.expectedID)
		return
	}

	// 标记健康
	c.rpcOK.Store(true)
	log.Printf("RPC ok: %s chainId=%d", c.url, chainID.Int64())
}

// RPCOK 返回当前节点是否健康
func (c *Client) RPCOK() bool {
	return c.rpcOK.Load()
}

// ChainID 返回最近一次记录的 chainId
func (c *Client) ChainID() int64 {
	return c.chainID.Load()
}
