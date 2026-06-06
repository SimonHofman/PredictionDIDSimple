# Audit Checklist (template)

| Area | Item | Status |
|------|------|--------|
| Access | Oracle roles, factory owner, pause | ☐ |
| Funds | Reentrancy guards, pull payments | ☐ |
| Math | CPMM rounding, fee accounting | ☐ |
| Oracle | Multisig threshold, double-execute | ☐ |
| LP | `removeLiquidity` pro-rata drain | ☐ |
| Compliance | Geo/KYC off-chain only — not on-chain | ☐ |

Record findings in issue tracker; tag releases `v3.x-auditN`.
