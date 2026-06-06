/**
 * 可验证凭证（VC）卡片组件
 * 展示单个可验证凭证的类型、过期时间和详细主题信息
 * @param {{ credential: object }} props - 凭证数据对象
 */
export default function VCCard({ credential }) {
  // 初始化 VC 解析对象
  let vc = {};
  try {
    // 尝试解析 VC JSON 数据（可能是字符串或对象格式）
    vc = typeof credential.vc_json === 'string'
      ? JSON.parse(credential.vc_json)
      : credential.vc_json;
  } catch {
    // 解析失败时使用空对象
    vc = {};
  }
  // 提取凭证主题信息
  const subject = vc.credentialSubject || {};
  // 提取凭证类型列表（数组则用逗号连接）
  const types = Array.isArray(vc.type) ? vc.type.join(', ') : vc.type;
  return (
    <div className="card">
      {/* 显示凭证类型名称（加粗） */}
      <strong>{credential.credential_type}</strong>
      {/* 显示 VC 类型列表 */}
      <p>类型：{types}</p>
      {/* 显示凭证过期时间 */}
      <p>过期：{new Date(credential.expires_at).toLocaleString()}</p>
      {/* 以 JSON 格式显示凭证主题详细信息 */}
      <pre style={{ fontSize: '0.75rem' }}>{JSON.stringify(subject, null, 2)}</pre>
    </div>
  );
}
