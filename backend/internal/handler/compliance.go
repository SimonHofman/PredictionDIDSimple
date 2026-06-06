// Package handler 合规检查接口
package handler

// 导入依赖
import (
	"net/http" // HTTP
)

// complianceRestricted 返回当前请求者所在国家的合规限制信息
func (a *API) complianceRestricted(w http.ResponseWriter, r *http.Request) {
	// 优先取 Cloudflare 注入的国家头
	country := r.Header.Get("CF-IPCountry")
	if country == "" {
		// 退回到自定义头
		country = r.Header.Get("X-Country-Code")
	}
	if country == "" {
		country = "UNKNOWN" // 未知
	}
	// 判断是否在封禁列表中
	blocked := a.Cfg.BlockedCountries[country]
	// 返回合规信息
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"country":             country,                  // 国家代码
		"restricted":          blocked,                  // 是否受限
		"compliance_required": a.Cfg.ComplianceRequired, // 是否要求合规
		"environment":         a.Cfg.Environment,        // 运行环境
	})
}
