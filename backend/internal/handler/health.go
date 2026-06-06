package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prediction-did/simple/internal/blockchain"
	"github.com/redis/go-redis/v9"
)

type Health struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
	Chain *blockchain.Client
}

func (h *Health) RegisterRoutes(mux interface {
	Get(string, http.HandlerFunc)
}) {
	mux.Get("/health", h.health)
	mux.Get("/ready", h.ready)
}

func (h *Health) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Health) ready(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ready := true
	body := map[string]interface{}{
		"status": "ready",
	}

	if h.DB != nil {
		if err := h.DB.Ping(ctx); err != nil {
			ready = false
			body["db_ok"] = false
			body["db_error"] = err.Error()
		} else {
			body["db_ok"] = true
		}
	}

	if h.Redis != nil {
		if err := h.Redis.Ping(ctx).Err(); err != nil {
			body["redis_ok"] = false
			body["redis_degraded"] = true
		} else {
			body["redis_ok"] = true
		}
	}

	if h.Chain != nil {
		body["rpc_ok"] = h.Chain.RPCOK()
		if cid := h.Chain.ChainID(); cid > 0 {
			body["chain_id"] = cid
		}
	}

	if !ready {
		body["status"] = "not ready"
		writeJSON(w, http.StatusServiceUnavailable, body)
		return
	}
	writeJSON(w, http.StatusOK, body)
}
