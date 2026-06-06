// Package auth 提供管理员认证、JWT、SIWE 钱包登录等认证能力
package auth

// 导入依赖
import (
	"net/http" // HTTP 服务
	"strings"  // 字符串处理
)

// AdminMiddleware 管理员鉴权中间件
// 通过 X-Admin-Key 或 Authorization: Bearer <key> 头校验
func AdminMiddleware(apiKey string) func(http.Handler) http.Handler {
	// 返回中间件
	return func(next http.Handler) http.Handler {
		// 实际请求处理函数
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 服务端未配置 apiKey 时直接返回不可用
			if apiKey == "" {
				http.Error(w, `{"error":"admin not configured"}`, http.StatusServiceUnavailable)
				return
			}
			// 优先从 X-Admin-Key 取
			key := r.Header.Get("X-Admin-Key")
			if key == "" {
				// 退回 Authorization Bearer
				key = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			}
			// 鉴权失败返回 403
			if key != apiKey {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}
			// 通过则放行
			next.ServeHTTP(w, r)
		})
	}
}
