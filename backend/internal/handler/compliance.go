package handler

import (
	"net/http"
)

func (a *API) complianceRestricted(w http.ResponseWriter, r *http.Request) {
	country := r.Header.Get("CF-IPCountry")
	if country == "" {
		country = r.Header.Get("X-Country-Code")
	}
	if country == "" {
		country = "UNKNOWN"
	}
	blocked := a.Cfg.BlockedCountries[country]
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"country":             country,
		"restricted":          blocked,
		"compliance_required": a.Cfg.ComplianceRequired,
		"environment":         a.Cfg.Environment,
	})
}
