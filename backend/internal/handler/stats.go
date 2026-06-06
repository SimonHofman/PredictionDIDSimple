// package handler 中关于平台统计与市场盘口的接口
package handler

import (
	"net/http" // HTTP
	"strconv"  // 数字解析
)

// platformStats 返回全平台的聚合统计
func (a *API) platformStats(w http.ResponseWriter, r *http.Request) {
	stats, err := a.Stats.Platform(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// marketPool 返回某个市场的池子状态（用于前端展示曲线）
func (a *API) marketPool(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	m, err := a.Markets.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"market_id":     m.ID,
		"market_type":   m.MarketType,
		"reserve_yes":   m.ReserveYes,
		"reserve_no":    m.ReserveNo,
		"price_yes_bps": m.PriceYesBps,
		"fee_bps":       m.FeeBps,
		"outcome_count": m.OutcomeCount,
	})
}

// marketOrderbook 返回 CPMM 模型下的合成盘口数据
func (a *API) marketOrderbook(w http.ResponseWriter, r *http.Request) {
	// CPMM synthetic orderbook snapshot for binary markets
	// 二元市场的 CPMM 合成盘口快照
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	m, err := a.Markets.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	// 解析 YES 价格（基点）
	yesBps := parseBps(m.PriceYesBps)
	// 用 yes/no 两条记录模拟买盘
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"bids": []map[string]interface{}{
			{"side": "yes", "price_bps": yesBps, "depth": coalesceReserve(m.ReserveYes, m.YesPool)},
			{"side": "no", "price_bps": 10000 - yesBps, "depth": coalesceReserve(m.ReserveNo, m.NoPool)},
		},
		"note": "CPMM snapshot (Phase 3)",
	})
}

// parseBps 解析基点字符串，缺失或 0 时回退为 5000（即 50%）
func parseBps(s string) int {
	if s == "" {
		return 5000
	}
	n, _ := strconv.Atoi(s)
	if n == 0 {
		return 5000
	}
	return n
}

// coalesceReserve 选取首选储备金；为空或 0 时返回备用值
func coalesceReserve(primary, fallback string) string {
	if primary != "" && primary != "0" {
		return primary
	}
	return fallback
}
