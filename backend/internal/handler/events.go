// Package handler SSE 事件流接口（比分推送）
package handler

// 导入依赖
import (
	"encoding/json" // JSON 序列化
	"fmt"           // 格式化输出
	"net/http"      // HTTP
	"time"          // 定时器
)

// streamScores 以 SSE（Server-Sent Events）方式实时推送比赛比分
func (a *API) streamScores(w http.ResponseWriter, r *http.Request) {
	// 设置 SSE 必要的 HTTP 头
	w.Header().Set("Content-Type", "text/event-stream") // SSE 格式
	w.Header().Set("Cache-Control", "no-cache")         // 禁止缓存
	w.Header().Set("Connection", "keep-alive")          // 长连接
	// 检查是否支持 Flush
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	// 每 5 秒推送一次数据
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done(): // 客户端断开
			return
		case <-ticker.C: // 定时触发
			// 查询最新比赛
			matches, err := a.Matches.List(r.Context(), "", 50, 0)
			if err != nil {
				continue // 出错跳过
			}
			// 序列化为 JSON
			payload, _ := json.Marshal(map[string]interface{}{"items": matches, "ts": time.Now().UTC()})
			// 输出 SSE data 帧
			fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush() // 立即推送
		}
	}
}
