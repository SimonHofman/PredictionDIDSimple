// package handler 中关于比赛信息的查询接口
package handler

import (
	"net/http" // HTTP
)

// listMatches 分页查询比赛列表
func (a *API) listMatches(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status") // 比赛状态过滤
	limit := queryInt(r, "limit", 20)     // 每页条数
	offset := queryInt(r, "offset", 0)    // 偏移
	list, err := a.Matches.List(r.Context(), status, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": list})
}

// getMatch 返回单场比赛详情，并附带其下的市场列表
func (a *API) getMatch(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	// 查询比赛
	m, err := a.Matches.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "match not found")
		return
	}
	// 取所有市场，再按 match_id 过滤；TODO：可下沉到 SQL 层
	markets, _ := a.Markets.List(r.Context(), "", 50, 0)
	var linked []interface{}
	for _, mk := range markets {
		if mk.MatchID != nil && *mk.MatchID == m.ID {
			linked = append(linked, mk)
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"match": m, "markets": linked})
}
