package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/prediction-did/simple/internal/auth"
)

type siweRequest struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

func (a *API) siweAuth(w http.ResponseWriter, r *http.Request) {
	var req siweRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	addr, err := auth.VerifySIWE(auth.SIWEConfig{
		Domain: a.Cfg.SIWEDomain,
		URI:    a.Cfg.SIWEURI,
	}, req.Message, req.Signature)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	user, err := a.Users.UpsertByAddress(r.Context(), addr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	token, err := auth.IssueJWT(a.Cfg.JWTSecret, addr, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

type bindDIDRequest struct {
	DID       string `json:"did"`
	Signature string `json:"signature"`
}

func (a *API) bindDID(w http.ResponseWriter, r *http.Request) {
	addr := auth.AddressFromContext(r.Context())
	var req bindDIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := auth.VerifyDIDBind(a.Cfg.ChainID, addr, req.DID, req.Signature); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := a.Users.BindDID(r.Context(), addr, req.DID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user, _ := a.Users.GetByAddress(r.Context(), addr)
	writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

func (a *API) myPositions(w http.ResponseWriter, r *http.Request) {
	addr := auth.AddressFromContext(r.Context())
	list, err := a.Positions.ListByUser(r.Context(), addr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": list})
}
