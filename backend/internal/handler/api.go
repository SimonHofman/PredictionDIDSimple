package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/prediction-did/simple/internal/auth"
	"github.com/prediction-did/simple/internal/blockchain"
	"github.com/prediction-did/simple/internal/config"
	"github.com/prediction-did/simple/internal/repository"
	"github.com/prediction-did/simple/internal/vc"
)

type API struct {
	Cfg          *config.Config
	Matches      *repository.MatchRepo
	Markets      *repository.MarketRepo
	Users        *repository.UserRepo
	Positions    *repository.PositionRepo
	OracleJobs   *repository.OracleJobRepo
	Credentials  *repository.CredentialRepo
	Stats        *repository.StatsRepo
	Compliance   *repository.ComplianceRepo
	VCIssuer     *vc.Issuer
	OracleChain  *blockchain.OracleClient
}

func (a *API) RegisterRoutes(r chi.Router) {
	r.Get("/matches", a.listMatches)
	r.Get("/matches/{id}", a.getMatch)
	r.Get("/markets", a.listMarkets)
	r.Get("/markets/{id}", a.getMarket)
	r.Get("/markets/{id}/pool", a.marketPool)
	r.Get("/markets/{id}/orderbook", a.marketOrderbook)
	r.Get("/stats/platform", a.platformStats)
	r.Get("/compliance/restricted", a.complianceRestricted)
	r.Post("/kyc/webhook", a.kycWebhook)
	r.Get("/events/scores", a.streamScores)
	r.Get("/metrics", a.prometheusMetrics)

	r.Post("/auth/siwe", a.siweAuth)
	r.Post("/auth/verify-vc", a.verifyVC)

	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(a.Cfg.JWTSecret))
		r.Get("/me/positions", a.myPositions)
		r.Post("/users/bind-did", a.bindDID)
		r.Get("/users/me/credentials", a.myCredentials)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.AdminMiddleware(a.Cfg.AdminAPIKey))
		r.Get("/admin/oracle-jobs", a.listOracleJobs)
		r.Post("/admin/markets", a.registerMarket)
		r.Post("/admin/markets/{id}/void", a.voidMarket)
		r.Post("/admin/oracle-jobs/{id}/retry", a.retryOracleJob)
		r.Post("/credentials/issue", a.issueCredential)
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}
