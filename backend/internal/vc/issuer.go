// Package vc 提供可验证凭证（Verifiable Credentials）的签发和验证功能
package vc

// 导入依赖
import (
	"crypto/hmac"     // HMAC 消息认证码
	"crypto/sha256"   // SHA-256 哈希算法
	"encoding/base64" // Base64 编解码
	"encoding/json"   // JSON 序列化和反序列化
	"fmt"             // 格式化输出
	"strings"         // 字符串处理
	"time"            // 时间处理
)

// Issuer 可验证凭证签发器结构体
// 使用 HMAC-SHA256 算法对凭证进行签名
type Issuer struct {
	key string // HMAC 签名密钥
}

// NewIssuer 创建新的凭证签发器实例
// 参数 key: 用于 HMAC 签名的密钥
// 返回: Issuer 指针
func NewIssuer(key string) *Issuer {
	return &Issuer{key: key} // 初始化并返回签发器实例
}

// IssueRequest 凭证签发请求结构体
// 包含签发一个可验证凭证所需的所有参数
type IssueRequest struct {
	SubjectDID string                 // 凭证主体的 DID（去中心化身份标识）
	Type       string                 // 凭证类型（如 "KYCCredential"）
	Claims     map[string]interface{} // 凭证声明内容（键值对）
	TTL        time.Duration          // 凭证有效期
}

// Issue 签发一个新的可验证凭证
// 按照 W3C 可验证凭证规范构建凭证 JSON，并附加 HMAC-SHA256 签名证明
// 参数 req: 凭证签发请求
// 返回: 序列化后的 JSON 凭证和错误信息
func (i *Issuer) Issue(req IssueRequest) (json.RawMessage, error) {
	if req.TTL == 0 {
		req.TTL = 365 * 24 * time.Hour // 默认有效期为 1 年
	}
	now := time.Now().UTC() // 获取当前 UTC 时间
	exp := now.Add(req.TTL) // 计算过期时间
	claims := req.Claims    // 获取凭证声明
	if claims == nil {
		claims = map[string]interface{}{} // 如果声明为空则初始化空 map
	}
	claims["id"] = req.SubjectDID // 将主体 DID 添加到声明中

	// 构建符合 W3C VC 规范的凭证结构
	vc := map[string]interface{}{
		"@context": []string{
			"https://www.w3.org/2018/credentials/v1", // W3C 凭证上下文
		},
		"type": []string{"VerifiableCredential", req.Type}, // 凭证类型数组
		"issuer": map[string]string{
			"id":   "did:web:prediction-did.local:issuer", // 签发者 DID
			"name": "Prediction DID Issuer",               // 签发者名称
		},
		"issuanceDate":      now.Format(time.RFC3339), // 签发日期（RFC3339 格式）
		"expirationDate":    exp.Format(time.RFC3339), // 过期日期（RFC3339 格式）
		"credentialSubject": claims,                   // 凭证主体声明
	}
	raw, err := json.Marshal(vc) // 将凭证序列化为 JSON
	if err != nil {
		return nil, err // 序列化失败时返回错误
	}
	sig := i.sign(raw) // 对凭证 JSON 进行 HMAC 签名
	// 构建签名证明对象
	proof := map[string]interface{}{
		"type":               "HMAC-SHA256",                               // 签名算法类型
		"proofPurpose":       "assertionMethod",                           // 证明用途
		"verificationMethod": "did:web:prediction-did.local:issuer#key-1", // 验证方法标识
		"proofValue":         sig,                                         // 签名值（Base64 编码）
	}
	vc["proof"] = proof     // 将证明附加到凭证
	return json.Marshal(vc) // 返回包含证明的完整凭证 JSON
}

// sign 使用 HMAC-SHA256 对数据进行签名
// 参数 payload: 需要签名的原始数据
// 返回: Base64 编码的签名字符串
func (i *Issuer) sign(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(i.key))             // 创建 HMAC 实例
	mac.Write(payload)                                     // 写入待签名数据
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)) // 返回 Base64 编码的签名
}

// Verify 验证可验证凭证的签名和有效期
// 参数 raw: 凭证的 JSON 原始数据
// 返回: 验证失败时的错误信息，nil 表示验证通过
func (i *Issuer) Verify(raw json.RawMessage) error {
	var vc map[string]interface{} // 存储解析后的凭证
	if err := json.Unmarshal(raw, &vc); err != nil {
		return err // JSON 解析失败
	}
	// 提取证明对象
	proof, ok := vc["proof"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("missing proof") // 缺少签名证明
	}
	proofVal, _ := proof["proofValue"].(string) // 获取签名值
	delete(vc, "proof")                         // 删除证明字段以重建原始负载
	payload, err := json.Marshal(vc)            // 重新序列化（不含证明）
	if err != nil {
		return err // 序列化失败
	}
	// 重新计算签名并与原签名比较
	if i.sign(payload) != proofVal {
		return fmt.Errorf("invalid vc signature") // 签名不匹配
	}
	// 检查凭证是否已过期
	expStr, _ := vc["expirationDate"].(string)
	if expStr != "" {
		exp, err := time.Parse(time.RFC3339, expStr) // 解析过期时间
		if err == nil && exp.Before(time.Now()) {
			return fmt.Errorf("credential expired") // 凭证已过期
		}
	}
	return nil // 验证通过
}

// SubjectRegion 从可验证凭证中提取主体的地区信息
// 参数 raw: 凭证的 JSON 原始数据
// 返回: 大写的地区代码字符串和错误信息
func SubjectRegion(raw json.RawMessage) (string, error) {
	var vc map[string]interface{} // 存储解析后的凭证
	if err := json.Unmarshal(raw, &vc); err != nil {
		return "", err // JSON 解析失败
	}
	// 提取凭证主体声明
	sub, ok := vc["credentialSubject"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no subject") // 缺少主体声明
	}
	region, _ := sub["region"].(string) // 获取地区字段
	return strings.ToUpper(region), nil // 返回大写的地区代码
}
