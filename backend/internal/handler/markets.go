// package handler 中市场列表与详情的处理函数
package handler

import (
	"net/http" // HTTP

	"github.com/prediction-did/simple/internal/auth" // 用户认证上下文
)

// listMarkets 分页查询市场列表
// 查询参数：status, limit, offset
func (a *API) listMarkets(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")    // 状态过滤（OPEN/RESOLVED/...）
	limit := queryInt(r, "limit", 20)         // 默认每页 20 条
	offset := queryInt(r, "offset", 0)        // 默认从 0 偏移
	list, err := a.Markets.List(r.Context(), status, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// 同时返回链 ID 与抵押代币地址，方便前端构造合约调用
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items":              list,
		"collateral_address": a.Cfg.CollateralAddress,
		"chain_id":           a.Cfg.ChainID,
	})
}

// getMarket 返回单个市场详情
// 同时返回访问门禁信息（access），告诉前端是否需要 VC 才能交易
func (a *API) getMarket(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	// 查询市场实体
	m, err := a.Markets.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "market not found")
		return
	}
	// 默认放行
	gate := map[string]interface{}{"allowed": true}
	// 若需要 VC，则检查当前用户是否持有
	if m.RequiresVC {
		addr := auth.AddressFromContext(r.Context())
		ok, _ := a.Credentials.HasValidType(r.Context(), addr, "VerifiedFan")
		gate["allowed"] = ok
		gate["requires_vc"] = true
		gate["credential_type"] = "VerifiedFan"
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"market":             m,
		"collateral_address": a.Cfg.CollateralAddress,
		"chain_id":           a.Cfg.ChainID,
		"access":             gate,
	})
}
