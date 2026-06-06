package auth

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey string

const AddressKey ctxKey = "address"

func Middleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(h, "Bearer ")
			claims, err := ParseJWT(secret, token)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), AddressKey, claims.Address)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AddressFromContext(ctx context.Context) string {
	v, _ := ctx.Value(AddressKey).(string)
	return v
}
