package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type ipLimiter struct {
	mu       sync.Mutex
	counts   map[string]int
	windowAt time.Time
	limit    int
}

func newIPLimiter(perMinute int) *ipLimiter {
	return &ipLimiter{
		counts: make(map[string]int),
		limit:  perMinute,
	}
}

func (l *ipLimiter) allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if now.Sub(l.windowAt) > time.Minute {
		l.counts = make(map[string]int)
		l.windowAt = now
	}
	l.counts[ip]++
	return l.counts[ip] <= l.limit
}

func RateLimit(perMinute int) func(http.Handler) http.Handler {
	lim := newIPLimiter(perMinute)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" || r.URL.Path == "/ready" {
				next.ServeHTTP(w, r)
				return
			}
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.RemoteAddr
			}
			if !lim.allow(ip) {
				http.Error(w, `{"error":"rate_limited"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
