export default function VCCard({ credential }) {
  let vc = {};
  try {
    vc = typeof credential.vc_json === 'string'
      ? JSON.parse(credential.vc_json)
      : credential.vc_json;
  } catch {
    vc = {};
  }
  const subject = vc.credentialSubject || {};
  const types = Array.isArray(vc.type) ? vc.type.join(', ') : vc.type;
  return (
    <div className="card">
      <strong>{credential.credential_type}</strong>
      <p>类型：{types}</p>
      <p>过期：{new Date(credential.expires_at).toLocaleString()}</p>
      <pre style={{ fontSize: '0.75rem' }}>{JSON.stringify(subject, null, 2)}</pre>
    </div>
  );
}
