// Package middleware 速率限制中间件（基于 IP 的简单滑动窗口）
package middleware

// 导入依赖
import (
	"net"      // 网络工具
	"net/http" // HTTP
	"sync"     // 互斥锁
	"time"     // 时间
)

// ipLimiter 基于 IP 的简单速率限制器（每分钟窗口）
type ipLimiter struct {
	mu       sync.Mutex     // 保护并发访问
	counts   map[string]int // IP -> 请求次数
	windowAt time.Time      // 当前窗口起始时间
	limit    int            // 每分钟允许的最大请求数
}

// newIPLimiter 创建限流器
func newIPLimiter(perMinute int) *ipLimiter {
	return &ipLimiter{
		counts: make(map[string]int), // 初始化计数 map
		limit:  perMinute,            // 设置上限
	}
}

// allow 判断是否允许此 IP 的请求
func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	// 超过一分钟重置窗口
	if now.Sub(l.windowAt) > time.Minute {
		l.counts = make(map[string]int) // 重置
		l.windowAt = now                // 更新窗口起始
	}
	l.counts[ip]++                 // 计数加 1
	return l.counts[ip] <= l.limit // 未超限返回 true
}

// RateLimit 速率限制中间件，每分钟允许 perMinute 次请求
func RateLimit(perMinute int) func(http.Handler) http.Handler {
	lim := newIPLimiter(perMinute) // 创建限流器
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 健康检查路径豁免限流
			if r.URL.Path == "/health" || r.URL.Path == "/ready" {
				next.ServeHTTP(w, r)
				return
			}
			// 提取客户端 IP
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.RemoteAddr // 兜底
			}
			// 判断是否超限
			if !lim.allow(ip) {
				http.Error(w, `{"error":"rate_limited"}`, http.StatusTooManyRequests)
				return
			}
			// 放行
			next.ServeHTTP(w, r)
		})
	}
}
