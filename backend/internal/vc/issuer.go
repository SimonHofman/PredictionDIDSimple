package vc

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Issuer struct {
	key string
}

func NewIssuer(key string) *Issuer {
	return &Issuer{key: key}
}

type IssueRequest struct {
	SubjectDID string
	Type       string
	Claims     map[string]interface{}
	TTL        time.Duration
}

func (i *Issuer) Issue(req IssueRequest) (json.RawMessage, error) {
	if req.TTL == 0 {
		req.TTL = 365 * 24 * time.Hour
	}
	now := time.Now().UTC()
	exp := now.Add(req.TTL)
	claims := req.Claims
	if claims == nil {
		claims = map[string]interface{}{}
	}
	claims["id"] = req.SubjectDID

	vc := map[string]interface{}{
		"@context": []string{
			"https://www.w3.org/2018/credentials/v1",
		},
		"type": []string{"VerifiableCredential", req.Type},
		"issuer": map[string]string{
			"id":   "did:web:prediction-did.local:issuer",
			"name": "Prediction DID Issuer",
		},
		"issuanceDate": now.Format(time.RFC3339),
		"expirationDate": exp.Format(time.RFC3339),
		"credentialSubject": claims,
	}
	raw, err := json.Marshal(vc)
	if err != nil {
		return nil, err
	}
	sig := i.sign(raw)
	proof := map[string]interface{}{
		"type":               "HMAC-SHA256",
		"proofPurpose":       "assertionMethod",
		"verificationMethod": "did:web:prediction-did.local:issuer#key-1",
		"proofValue":         sig,
	}
	vc["proof"] = proof
	return json.Marshal(vc)
}

func (i *Issuer) sign(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(i.key))
	mac.Write(payload)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (i *Issuer) Verify(raw json.RawMessage) error {
	var vc map[string]interface{}
	if err := json.Unmarshal(raw, &vc); err != nil {
		return err
	}
	proof, ok := vc["proof"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing proof")
	}
	proofVal, _ := proof["proofValue"].(string)
	delete(vc, "proof")
	payload, err := json.Marshal(vc)
	if err != nil {
		return err
	}
	if i.sign(payload) != proofVal {
		return fmt.Errorf("invalid vc signature")
	}
	expStr, _ := vc["expirationDate"].(string)
	if expStr != "" {
		exp, err := time.Parse(time.RFC3339, expStr)
		if err == nil && exp.Before(time.Now()) {
			return fmt.Errorf("credential expired")
		}
	}
	return nil
}

func SubjectRegion(raw json.RawMessage) (string, error) {
	var vc map[string]interface{}
	if err := json.Unmarshal(raw, &vc); err != nil {
		return "", err
	}
	sub, ok := vc["credentialSubject"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no subject")
	}
	region, _ := sub["region"].(string)
	return strings.ToUpper(region), nil
}
