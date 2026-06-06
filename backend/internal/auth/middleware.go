// Package auth JWT 校验中间件，保护需要登录的接口
package auth

// 导入依赖
import (
	"context"  // 上下文
	"net/http" // HTTP
	"strings"  // 字符串
)

// ctxKey 自定义 context key 类型，避免冲突
type ctxKey string

// AddressKey 用于从 context 取得已认证的地址
const AddressKey ctxKey = "address"

// Middleware 校验 Authorization: Bearer <jwt> 头并将地址注入 context
func Middleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 取 Authorization 头
			h := r.Header.Get("Authorization")
			// 必须以 Bearer 开头
			if !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			// 提取 token 字符串
			token := strings.TrimPrefix(h, "Bearer ")
			// 解析 token
			claims, err := ParseJWT(secret, token)
			if err != nil {
				// 解析失败返回 401
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			// 将地址注入到 context
			ctx := context.WithValue(r.Context(), AddressKey, claims.Address)
			// 继续处理
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AddressFromContext 从 context 获取已认证的钱包地址
func AddressFromContext(ctx context.Context) string {
	v, _ := ctx.Value(AddressKey).(string) // 类型断言失败则空字符串
	return v
}
