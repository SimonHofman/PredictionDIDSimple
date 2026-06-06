// Package middleware 地理封锁中间件
package middleware

// 导入依赖
import (
	"net/http" // HTTP
	"strings"  // 字符串

	"github.com/prediction-did/simple/internal/config" // 配置
)

// GeoBlock 基于 IP 归属国的地理封锁中间件
// logFn 可用于审计日志记录
func GeoBlock(cfg *config.Config, logFn func(ip, country, path string, allowed bool)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 未开启封锁或请求路径豁免时直接放行
			if !cfg.GeoBlockEnabled || isExemptPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			// 检测请求国家
			country := detectCountry(r)
			// 判断是否允许
			allowed := !cfg.BlockedCountries[country]
			// 记录审计日志
			if logFn != nil {
				logFn(r.RemoteAddr, country, r.URL.Path, allowed)
			}
			// 封禁则返回 403
			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"region_restricted","country":"` + country + `"}`))
				return
			}
			// 放行
			next.ServeHTTP(w, r)
		})
	}
}

// isExemptPath 判断路径是否豁免地理封锁
func isExemptPath(path string) bool {
	switch path {
	case "/health", "/ready", "/metrics", "/compliance/restricted":
		return true // 这些路径总是放行
	default:
		return strings.HasPrefix(path, "/events/") // 事件流路径也放行
	}
}

// detectCountry 从请求头推断国家代码
func detectCountry(r *http.Request) string {
	// Cloudflare 注入的头
	if c := r.Header.Get("CF-IPCountry"); c != "" {
		return strings.ToUpper(c)
	}
	// 自定义头
	if c := r.Header.Get("X-Country-Code"); c != "" {
		return strings.ToUpper(c)
	}
	return "UNKNOWN" // 未知国家
}
