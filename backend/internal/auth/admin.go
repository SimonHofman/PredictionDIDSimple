package auth

import (
	"net/http"
	"strings"
)

func AdminMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if apiKey == "" {
				http.Error(w, `{"error":"admin not configured"}`, http.StatusServiceUnavailable)
				return
			}
			key := r.Header.Get("X-Admin-Key")
			if key == "" {
				key = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			}
			if key != apiKey {
				http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
