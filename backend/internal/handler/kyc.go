package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/prediction-did/simple/internal/vc"
)

func (a *API) kycWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body")
		return
	}
	if a.Cfg.KYCWebhookSecret != "" {
		sig := r.Header.Get("X-KYC-Signature")
		mac := hmac.New(sha256.New, []byte(a.Cfg.KYCWebhookSecret))
		mac.Write(body)
		expected := hex.EncodeToString(mac.Sum(nil))
		if sig == "" || !hmac.Equal([]byte(sig), []byte(expected)) {
			writeError(w, http.StatusUnauthorized, "invalid signature")
			return
		}
	}
	var payload struct {
		ExternalID  string `json:"external_id"`
		UserAddress string `json:"user_address"`
		Status      string `json:"status"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if payload.ExternalID == "" || payload.Status == "" {
		writeError(w, http.StatusBadRequest, "missing fields")
		return
	}
	if err := a.Compliance.LogKYC(r.Context(), payload.ExternalID, payload.UserAddress, payload.Status, body); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if payload.Status == "approved" && payload.UserAddress != "" {
		_, _ = a.VCIssuer.Issue(vc.IssueRequest{
			SubjectDID: "did:ethr:" + payload.UserAddress,
			Type:       "KYCVerification",
			Claims:     map[string]interface{}{"kyc": "approved"},
		})
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}
