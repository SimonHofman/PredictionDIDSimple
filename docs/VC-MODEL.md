# 可验证凭证（VC）数据模型

## 示例 VC

```json
{
  "@context": ["https://www.w3.org/2018/credentials/v1"],
  "type": ["VerifiableCredential", "VerifiedFan"],
  "issuer": {
    "id": "did:web:prediction-did.local:issuer",
    "name": "Prediction DID Issuer"
  },
  "issuanceDate": "2026-06-03T12:00:00Z",
  "expirationDate": "2027-06-03T12:00:00Z",
  "credentialSubject": {
    "id": "did:pkh:eip155:31337:0xabc...",
    "region": "EU",
    "fanLevel": "verified"
  },
  "proof": {
    "type": "HMAC-SHA256",
    "proofPurpose": "assertionMethod",
    "verificationMethod": "did:web:prediction-did.local:issuer#key-1",
    "proofValue": "<base64>"
  }
}
```

## 签发（管理员）

```http
POST /credentials/issue
X-Admin-Key: dev-admin-key

{
  "address": "0x...",
  "credential_type": "VerifiedFan",
  "claims": { "region": "EU", "fanLevel": "verified" },
  "ttl_hours": 8760
}
```

## 校验

- `POST /auth/verify-vc` — 校验签名与过期
- 受限市场：`markets.requires_vc=true` 时，`GET /markets/:id` 返回 `access.allowed`

## 密钥

生产环境将 `VC_ISSUER_KEY` 存 KMS，勿提交 Git。
