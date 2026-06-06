package handler

import (
	"encoding/json"
	"net/http"

	"github.com/prediction-did/simple/internal/repository"
)

func (a *API) listOracleJobs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	jobs, err := a.OracleJobs.ListAll(r.Context(), status, 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": jobs})
}

type voidMarketRequest struct {
	Reason string `json:"reason"`
}

func (a *API) voidMarket(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	mk, err := a.Markets.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "market not found")
		return
	}
	if a.OracleChain != nil {
		if _, err := a.OracleChain.VoidMarket(r.Context(), mk.MarketAddress); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if err := a.Markets.SetVoid(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if mk.MatchID != nil {
		_ = a.Matches.SetStatus(r.Context(), *mk.MatchID, "CANCELLED")
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "void"})
}

type registerMarketRequest struct {
	MatchID          int64  `json:"match_id"`
	MarketAddress    string `json:"market_address"`
	Question         string `json:"question"`
	RequiresVC       bool   `json:"requires_vc"`
	RestrictedRegion string `json:"restricted_region"`
	ResolutionRule   string `json:"resolution_rule"`
}

func (a *API) registerMarket(w http.ResponseWriter, r *http.Request) {
	var req registerMarketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	err := a.Markets.RegisterAdmin(r.Context(), repository.AdminMarketUpdate{
		MatchID:          req.MatchID,
		RequiresVC:       req.RequiresVC,
		RestrictedRegion: req.RestrictedRegion,
		ResolutionRule:   req.ResolutionRule,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "registered"})
}

func (a *API) retryOracleJob(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	_ = a.OracleJobs.UpdateStatus(r.Context(), id, "pending", map[string]interface{}{
		"error_message": nil,
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "pending"})
}
