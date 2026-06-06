// Package handler 健康检查与就绪检测接口
package handler

// 导入依赖
import (
	"net/http" // HTTP

	"github.com/jackc/pgx/v5/pgxpool"                      // 数据库连接池
	"github.com/prediction-did/simple/internal/blockchain" // 区块链客户端
	"github.com/redis/go-redis/v9"                         // Redis 客户端
)

// Health 健康检查处理器，聚合后端各组件状态
type Health struct {
	DB    *pgxpool.Pool      // 数据库连接池
	Redis *redis.Client      // Redis 客户端
	Chain *blockchain.Client // 区块链客户端
}

// RegisterRoutes 注册健康检查路由
func (h *Health) RegisterRoutes(mux interface {
	Get(string, http.HandlerFunc)
}) {
	mux.Get("/health", h.health) // 存活探针
	mux.Get("/ready", h.ready)   // 就绪探针
}

// health 简单存活检查（始终返回 ok）
func (h *Health) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ready 就绪检查：检测 DB/Redis/RPC 各组件是否可用
func (h *Health) ready(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ready := true // 总体就绪标记
	body := map[string]interface{}{
		"status": "ready",
	}

	// 数据库检测
	if h.DB != nil {
		if err := h.DB.Ping(ctx); err != nil {
			ready = false // 不就绪
			body["db_ok"] = false
			body["db_error"] = err.Error()
		} else {
			body["db_ok"] = true
		}
	}

	// Redis 检测（可降级运行）
	if h.Redis != nil {
		if err := h.Redis.Ping(ctx).Err(); err != nil {
			body["redis_ok"] = false
			body["redis_degraded"] = true // 降级标记
		} else {
			body["redis_ok"] = true
		}
	}

	// 区块链 RPC 检测
	if h.Chain != nil {
		body["rpc_ok"] = h.Chain.RPCOK() // RPC 是否健康
		if cid := h.Chain.ChainID(); cid > 0 {
			body["chain_id"] = cid // 当前链 ID
		}
	}

	// 返回结果
	if !ready {
		body["status"] = "not ready"
		writeJSON(w, http.StatusServiceUnavailable, body) // 503
		return
	}
	writeJSON(w, http.StatusOK, body) // 200
}
