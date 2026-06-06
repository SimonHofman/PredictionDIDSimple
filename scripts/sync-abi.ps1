# Sync contract ABIs to backend and frontend
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location (Join-Path $Root "contracts")
npm run compile
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
npm run export-abi
exit $LASTEXITCODE
