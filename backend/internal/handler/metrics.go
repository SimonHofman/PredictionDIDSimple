// package handler 中暴露 Prometheus 指标的接口
package handler

import (
	"fmt"     // 写入指标文本
	"net/http" // HTTP
)

// prometheusMetrics 以 Prometheus 文本格式暴露关键指标
// 当前主要统计 Oracle 作业的各类状态计数
func (a *API) prometheusMetrics(w http.ResponseWriter, r *http.Request) {
	// 取最近 1000 条作业（足够覆盖近期状态）
	jobs, _ := a.OracleJobs.ListAll(r.Context(), "", 1000)
	// 各状态计数器
	pending, manual, confirmed, failed := 0, 0, 0, 0
	for _, j := range jobs {
		switch j.Status {
		case "pending":
			pending++
		case "manual_review":
			manual++
		case "confirmed":
			confirmed++
		case "failed":
			failed++
		}
	}
	// Prometheus 期望的 Content-Type
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	// 输出 HELP 注释与数据行
	fmt.Fprintf(w, "# HELP oracle_jobs_total Oracle jobs by status\n")
	fmt.Fprintf(w, "oracle_jobs_pending %d\n", pending)
	fmt.Fprintf(w, "oracle_jobs_manual_review %d\n", manual)
	fmt.Fprintf(w, "oracle_jobs_confirmed %d\n", confirmed)
	fmt.Fprintf(w, "oracle_jobs_failed %d\n", failed)
}
