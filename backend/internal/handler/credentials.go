package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/prediction-did/simple/internal/auth"
	"github.com/prediction-did/simple/internal/repository"
	"github.com/prediction-did/simple/internal/vc"
)

type issueVCRequest struct {
	Address        string                 `json:"address"`
	CredentialType string                 `json:"credential_type"`
	Claims         map[string]interface{} `json:"claims"`
	TTLHours       int                    `json:"ttl_hours"`
}

func (a *API) issueCredential(w http.ResponseWriter, r *http.Request) {
	var req issueVCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if req.Address == "" || req.CredentialType == "" {
		writeError(w, http.StatusBadRequest, "address and credential_type required")
		return
	}
	ttl := 24 * time.Hour * 365
	if req.TTLHours > 0 {
		ttl = time.Duration(req.TTLHours) * time.Hour
	}
	did := fmt.Sprintf("did:pkh:eip155:%d:%s", a.Cfg.ChainID, req.Address)
	if u, _ := a.Users.GetByAddress(r.Context(), req.Address); u != nil && u.DID != nil {
		did = *u.DID
	}
	raw, err := a.VCIssuer.Issue(vc.IssueRequest{
		SubjectDID: did,
		Type:       req.CredentialType,
		Claims:     req.Claims,
		TTL:        ttl,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	id, err := a.Credentials.Insert(r.Context(), repository.Credential{
		UserAddress:    req.Address,
		CredentialType: req.CredentialType,
		VCJSON:         raw,
		ExpiresAt:      time.Now().Add(ttl),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"id": id, "vc": json.RawMessage(raw)})
}

func (a *API) myCredentials(w http.ResponseWriter, r *http.Request) {
	addr := auth.AddressFromContext(r.Context())
	list, err := a.Credentials.ListByUser(r.Context(), addr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": list})
}

type verifyVCRequest struct {
	VCJSON         json.RawMessage `json:"vc_json"`
	CredentialType string          `json:"credential_type"`
	Region         string          `json:"region"`
}

func (a *API) verifyVC(w http.ResponseWriter, r *http.Request) {
	var req verifyVCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := a.VCIssuer.Verify(req.VCJSON); err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if req.Region != "" {
		got, _ := vc.SubjectRegion(req.VCJSON)
		if got != req.Region {
			writeError(w, http.StatusForbidden, "region not allowed")
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]bool{"valid": true})
}
