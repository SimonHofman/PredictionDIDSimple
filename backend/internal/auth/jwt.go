// Package auth JWT 颁发与解析
package auth

// 导入依赖
import (
	"fmt"     // 错误格式化
	"strings" // 字符串处理
	"time"    // 时间计算

	"github.com/golang-jwt/jwt/v5" // JWT v5 库
)

// Claims 自定义 JWT Claims，包含钱包地址
type Claims struct {
	Address              string `json:"address"` // 钱包地址（小写）
	jwt.RegisteredClaims        // 标准注册项（exp/iat 等）
}

// IssueJWT 颁发 JWT
// secret 是签名密钥；address 钱包地址；ttl 有效期
func IssueJWT(secret, address string, ttl time.Duration) (string, error) {
	// 构造 claims
	claims := Claims{
		Address: strings.ToLower(address), // 地址转小写规范化
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),          // 签发时间
		},
	}
	// 使用 HS256 签名
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 输出签名后的字符串
	return token.SignedString([]byte(secret))
}

// ParseJWT 解析并校验 JWT 字符串
func ParseJWT(secret, tokenStr string) (*Claims, error) {
	// 调用 jwt.ParseWithClaims 解析
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		// 限制签名算法必须是 HS256
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		// 返回密钥
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err // 解析失败
	}
	// 类型断言
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		// 无效 token
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
