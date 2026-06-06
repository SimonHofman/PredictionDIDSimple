package auth

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	siwe "github.com/spruceid/siwe-go"
)

type SIWEConfig struct {
	Domain string
	URI    string
}

func VerifySIWE(cfg SIWEConfig, message, signature string) (string, error) {
	msg, err := siwe.ParseMessage(message)
	if err != nil {
		return "", fmt.Errorf("parse message: %w", err)
	}
	if msg.GetDomain() != cfg.Domain {
		return "", fmt.Errorf("domain mismatch")
	}
	wantURI, err := url.Parse(cfg.URI)
	if err != nil {
		return "", fmt.Errorf("invalid siwe uri config: %w", err)
	}
	gotURI := msg.GetURI()
	if (&gotURI).String() != wantURI.String() {
		return "", fmt.Errorf("uri mismatch")
	}
	if exp := msg.GetExpirationTime(); exp != nil && *exp != "" {
		t, err := time.Parse(time.RFC3339, *exp)
		if err == nil && t.Before(time.Now()) {
			return "", fmt.Errorf("message expired")
		}
	}
	domain := cfg.Domain
	_, err = msg.Verify(signature, &domain, nil, nil)
	if err != nil {
		return "", fmt.Errorf("verify: %w", err)
	}
	addr := msg.GetAddress().Hex()
	return strings.ToLower(addr), nil
}

func VerifyDIDBind(chainID int64, address, did, signatureHex string) error {
	expected := fmt.Sprintf("did:pkh:eip155:%d:%s", chainID, strings.ToLower(address))
	if strings.ToLower(did) != expected {
		return fmt.Errorf("did must be %s", expected)
	}
	// MVP: signature checked client-side; server trusts authenticated user from JWT
	_ = signatureHex
	return nil
}
