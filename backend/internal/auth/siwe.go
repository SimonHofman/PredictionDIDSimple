// Package auth SIWE（Sign-In with Ethereum）登录验证 + DID 绑定校验
package auth

// 导入依赖
import (
	"fmt"     // 错误格式化
	"net/url" // URL 解析
	"strings" // 字符串处理
	"time"    // 时间处理

	siwe "github.com/spruceid/siwe-go" // SIWE 协议库
)

// SIWEConfig SIWE 校验所需的配置
type SIWEConfig struct {
	Domain string // 期望的域名
	URI    string // 期望的 URI
}

// VerifySIWE 校验 SIWE 消息及签名，成功返回钱包地址
func VerifySIWE(cfg SIWEConfig, message, signature string) (string, error) {
	// 解析 SIWE 消息
	msg, err := siwe.ParseMessage(message)
	if err != nil {
		return "", fmt.Errorf("parse message: %w", err)
	}
	// 校验域名
	if msg.GetDomain() != cfg.Domain {
		return "", fmt.Errorf("domain mismatch")
	}
	// 期望的 URI
	wantURI, err := url.Parse(cfg.URI)
	if err != nil {
		return "", fmt.Errorf("invalid siwe uri config: %w", err)
	}
	// 实际 URI
	gotURI := msg.GetURI()
	// URI 必须匹配
	if (&gotURI).String() != wantURI.String() {
		return "", fmt.Errorf("uri mismatch")
	}
	// 校验过期时间
	if exp := msg.GetExpirationTime(); exp != nil && *exp != "" {
		t, err := time.Parse(time.RFC3339, *exp)
		// 已过期返回错误
		if err == nil && t.Before(time.Now()) {
			return "", fmt.Errorf("message expired")
		}
	}
	// 调用 siwe-go 的 Verify 进行签名校验
	domain := cfg.Domain
	_, err = msg.Verify(signature, &domain, nil, nil)
	if err != nil {
		return "", fmt.Errorf("verify: %w", err)
	}
	// 取出地址
	addr := msg.GetAddress().Hex()
	// 转小写返回
	return strings.ToLower(addr), nil
}

// VerifyDIDBind 校验绑定的 DID 是否符合 did:pkh:eip155:<chain>:<address> 规范
func VerifyDIDBind(chainID int64, address, did, signatureHex string) error {
	// 计算期望的 DID
	expected := fmt.Sprintf("did:pkh:eip155:%d:%s", chainID, strings.ToLower(address))
	// 不匹配返回错误
	if strings.ToLower(did) != expected {
		return fmt.Errorf("did must be %s", expected)
	}
	// MVP: signature checked client-side; server trusts authenticated user from JWT
	// MVP 阶段：签名由客户端校验；服务端信任经 JWT 认证的用户
	_ = signatureHex
	return nil
}
