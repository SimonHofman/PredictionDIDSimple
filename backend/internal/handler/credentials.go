// Package handler 可验证凭证（VC）颁发与查询接口
package handler

// 导入依赖
import (
	"encoding/json" // JSON
	"fmt"           // 格式化
	"net/http"      // HTTP
	"time"          // 时间

	"github.com/prediction-did/simple/internal/auth"       // 认证
	"github.com/prediction-did/simple/internal/repository" // 仓储
	"github.com/prediction-did/simple/internal/vc"         // VC 颁发器
)

// issueVCRequest 颁发 VC 请求体
type issueVCRequest struct {
	Address        string                 `json:"address"`         // 目标钱包地址
	CredentialType string                 `json:"credential_type"` // 凭证类型
	Claims         map[string]interface{} `json:"claims"`          // 自定义声明
	TTLHours       int                    `json:"ttl_hours"`       // 有效时长（小时）
}

// issueCredential 管理员颁发一个可验证凭证
func (a *API) issueCredential(w http.ResponseWriter, r *http.Request) {
	var req issueVCRequest
	// 解析请求
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	// 必填校验
	if req.Address == "" || req.CredentialType == "" {
		writeError(w, http.StatusBadRequest, "address and credential_type required")
		return
	}
	// 默认 TTL 为 365 天
	ttl := 24 * time.Hour * 365
	if req.TTLHours > 0 {
		ttl = time.Duration(req.TTLHours) * time.Hour
	}
	// 构建 DID：优先使用已绑定的 DID
	did := fmt.Sprintf("did:pkh:eip155:%d:%s", a.Cfg.ChainID, req.Address)
	if u, _ := a.Users.GetByAddress(r.Context(), req.Address); u != nil && u.DID != nil {
		did = *u.DID // 已绑定的 DID
	}
	// 调用 VCIssuer 签发
	raw, err := a.VCIssuer.Issue(vc.IssueRequest{
		SubjectDID: did,                // 主体 DID
		Type:       req.CredentialType, // 凭证类型
		Claims:     req.Claims,         // 声明
		TTL:        ttl,                // 有效期
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// 存储到数据库
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
	// 返回新凭证 ID 及 JSON
	writeJSON(w, http.StatusCreated, map[string]interface{}{"id": id, "vc": json.RawMessage(raw)})
}

// myCredentials 获取当前用户的凭证列表
func (a *API) myCredentials(w http.ResponseWriter, r *http.Request) {
	addr := auth.AddressFromContext(r.Context()) // 当前用户地址
	list, err := a.Credentials.ListByUser(r.Context(), addr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": list})
}

// verifyVCRequest 验证 VC 请求体
type verifyVCRequest struct {
	VCJSON         json.RawMessage `json:"vc_json"`         // VC JSON
	CredentialType string          `json:"credential_type"` // 凭证类型
	Region         string          `json:"region"`          // 要求匹配的地区
}

// verifyVC 验证一个可验证凭证的有效性
func (a *API) verifyVC(w http.ResponseWriter, r *http.Request) {
	var req verifyVCRequest
	// 解析请求
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	// 校验签名与过期
	if err := a.VCIssuer.Verify(req.VCJSON); err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	// 如果指定了地区，额外校验
	if req.Region != "" {
		got, _ := vc.SubjectRegion(req.VCJSON)
		if got != req.Region {
			writeError(w, http.StatusForbidden, "region not allowed")
			return
		}
	}
	// 验证通过
	writeJSON(w, http.StatusOK, map[string]bool{"valid": true})
}
