// Package handler KYC 回调接口
package handler

// 导入依赖
import (
	"crypto/hmac"   // HMAC 签名校验
	"crypto/sha256" // SHA-256 哈希
	"encoding/hex"  // 十六进制编码
	"encoding/json" // JSON 解码
	"io"            // IO 工具
	"net/http"      // HTTP

	"github.com/prediction-did/simple/internal/vc" // VC 颁发
)

// kycWebhook 接收第三方 KYC 回调通知
func (a *API) kycWebhook(w http.ResponseWriter, r *http.Request) {
	// 限制 body 最大 1MB
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "read body")
		return
	}
	// 如果配置了 Webhook Secret 则校验签名
	if a.Cfg.KYCWebhookSecret != "" {
		sig := r.Header.Get("X-KYC-Signature") // 请求中的签名
		// 计算 HMAC-SHA256
		mac := hmac.New(sha256.New, []byte(a.Cfg.KYCWebhookSecret))
		mac.Write(body)
		expected := hex.EncodeToString(mac.Sum(nil))
		// 签名比对
		if sig == "" || !hmac.Equal([]byte(sig), []byte(expected)) {
			writeError(w, http.StatusUnauthorized, "invalid signature")
			return
		}
	}
	// 解析回调负载
	var payload struct {
		ExternalID  string `json:"external_id"`  // 外部 KYC ID
		UserAddress string `json:"user_address"` // 用户钱包地址
		Status      string `json:"status"`       // KYC 状态
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	// 必填校验
	if payload.ExternalID == "" || payload.Status == "" {
		writeError(w, http.StatusBadRequest, "missing fields")
		return
	}
	// 将 KYC 记录写入合规表
	if err := a.Compliance.LogKYC(r.Context(), payload.ExternalID, payload.UserAddress, payload.Status, body); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// 如果通过审核且提供了地址，自动颁发 KYC VC
	if payload.Status == "approved" && payload.UserAddress != "" {
		_, _ = a.VCIssuer.Issue(vc.IssueRequest{
			SubjectDID: "did:ethr:" + payload.UserAddress,         // 构建 DID
			Type:       "KYCVerification",                         // 凭证类型
			Claims:     map[string]interface{}{"kyc": "approved"}, // 声明
		})
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}
