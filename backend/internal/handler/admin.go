// Package handler 管理员接口处理函数
package handler

// 导入依赖
import (
	"encoding/json" // JSON 解码
	"net/http"      // HTTP

	"github.com/prediction-did/simple/internal/repository" // 仓储层
)

// listOracleJobs 列出 Oracle 任务（支持按 status 过滤）
func (a *API) listOracleJobs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status") // 可选查询参数
	// 获取最多 100 条任务
	jobs, err := a.OracleJobs.ListAll(r.Context(), status, 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// 返回任务列表
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": jobs})
}

// voidMarketRequest 作废市场请求体
type voidMarketRequest struct {
	Reason string `json:"reason"` // 作废原因
}

// voidMarket 管理员作废某个市场
func (a *API) voidMarket(w http.ResponseWriter, r *http.Request) {
	// 解析 URL 中的市场 ID
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	// 查找市场
	mk, err := a.Markets.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "market not found")
		return
	}
	// 如果配置了 Oracle 链客户端则调用链上作废
	if a.OracleChain != nil {
		if _, err := a.OracleChain.VoidMarket(r.Context(), mk.MarketAddress); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	// 更新数据库状态为 VOID
	if err := a.Markets.SetVoid(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// 关联的比赛状态改为 CANCELLED
	if mk.MatchID != nil {
		_ = a.Matches.SetStatus(r.Context(), *mk.MatchID, "CANCELLED")
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "void"})
}

// registerMarketRequest 注册市场请求体
type registerMarketRequest struct {
	MatchID          int64  `json:"match_id"`          // 关联比赛 ID
	MarketAddress    string `json:"market_address"`    // 合约地址
	Question         string `json:"question"`          // 预测问题
	RequiresVC       bool   `json:"requires_vc"`       // 是否需要 VC
	RestrictedRegion string `json:"restricted_region"` // 限制地区
	ResolutionRule   string `json:"resolution_rule"`   // 解析规则
}

// registerMarket 管理员注册新市场
func (a *API) registerMarket(w http.ResponseWriter, r *http.Request) {
	var req registerMarketRequest
	// 解析请求体
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	// 写入数据库
	err := a.Markets.RegisterAdmin(r.Context(), repository.AdminMarketUpdate{
		MatchID:          req.MatchID,
		RequiresVC:       req.RequiresVC,
		RestrictedRegion: req.RestrictedRegion,
		ResolutionRule:   req.ResolutionRule,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "registered"})
}

// retryOracleJob 管理员重试失败的 Oracle 任务
func (a *API) retryOracleJob(w http.ResponseWriter, r *http.Request) {
	// 解析任务 ID
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	// 将状态重置为 pending，清除错误消息
	_ = a.OracleJobs.UpdateStatus(r.Context(), id, "pending", map[string]interface{}{
		"error_message": nil,
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "pending"})
}
