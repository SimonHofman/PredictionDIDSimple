// Package handler 认证相关 HTTP 处理器
package handler

// 导入依赖
import (
	"encoding/json" // JSON 解码
	"net/http"      // HTTP
	"time"          // 时间

	"github.com/prediction-did/simple/internal/auth" // 认证工具
)

// siweRequest SIWE 登录请求体
type siweRequest struct {
	Message   string `json:"message"`   // SIWE 签名消息
	Signature string `json:"signature"` // 钱包签名
}

// siweAuth 处理 SIWE 登录：验证签名 -> 创建/获取用户 -> 颁发 JWT
func (a *API) siweAuth(w http.ResponseWriter, r *http.Request) {
	var req siweRequest
	// 解码请求体
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	// 验证 SIWE 签名
	addr, err := auth.VerifySIWE(auth.SIWEConfig{
		Domain: a.Cfg.SIWEDomain, // 期望域名
		URI:    a.Cfg.SIWEURI,    // 期望 URI
	}, req.Message, req.Signature)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	// 创建或获取用户
	user, err := a.Users.UpsertByAddress(r.Context(), addr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// 颁发 JWT（24 小时有效期）
	token, err := auth.IssueJWT(a.Cfg.JWTSecret, addr, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// 返回 token 和用户信息
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user":  user,
	})
}

// bindDIDRequest 绑定 DID 请求体
type bindDIDRequest struct {
	DID       string `json:"did"`       // DID 标识符
	Signature string `json:"signature"` // 签名
}

// bindDID 将 DID 绑定到已认证用户
func (a *API) bindDID(w http.ResponseWriter, r *http.Request) {
	// 从 context 获取已认证地址
	addr := auth.AddressFromContext(r.Context())
	var req bindDIDRequest
	// 解码请求
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}
	// 校验 DID 格式与地址匹配
	if err := auth.VerifyDIDBind(a.Cfg.ChainID, addr, req.DID, req.Signature); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	// 存入数据库
	if err := a.Users.BindDID(r.Context(), addr, req.DID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	// 返回更新后的用户
	user, _ := a.Users.GetByAddress(r.Context(), addr)
	writeJSON(w, http.StatusOK, map[string]interface{}{"user": user})
}

// myPositions 获取当前登录用户的持仓列表
func (a *API) myPositions(w http.ResponseWriter, r *http.Request) {
	// 从 context 获取地址
	addr := auth.AddressFromContext(r.Context())
	// 查询持仓
	list, err := a.Positions.ListByUser(r.Context(), addr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"items": list})
}
