package middleware

import (
	"net/http"
	"strings"

	"github.com/prediction-did/simple/internal/config"
)

func GeoBlock(cfg *config.Config, logFn func(ip, country, path string, allowed bool)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.GeoBlockEnabled || isExemptPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			country := detectCountry(r)
			allowed := !cfg.BlockedCountries[country]
			if logFn != nil {
				logFn(r.RemoteAddr, country, r.URL.Path, allowed)
			}
			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error":"region_restricted","country":"` + country + `"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isExemptPath(path string) bool {
	switch path {
	case "/health", "/ready", "/metrics", "/compliance/restricted":
		return true
	default:
		return strings.HasPrefix(path, "/events/")
	}
}

func detectCountry(r *http.Request) string {
	if c := r.Header.Get("CF-IPCountry"); c != "" {
		return strings.ToUpper(c)
	}
	if c := r.Header.Get("X-Country-Code"); c != "" {
		return strings.ToUpper(c)
	}
	return "UNKNOWN"
}
