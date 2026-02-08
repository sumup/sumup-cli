param(
  [Parameter(Mandatory = $true)]
  [string]$BinaryPath
)

$ErrorActionPreference = "Stop"

# Placeholder script for future Windows Authenticode signing.
# Expected env vars once enabled:
# - WINDOWS_CERT_BASE64
# - WINDOWS_CERT_PASSWORD
#
# Example (to enable later):
# [IO.File]::WriteAllBytes("cert.pfx", [Convert]::FromBase64String($env:WINDOWS_CERT_BASE64))
# signtool sign /f cert.pfx /p $env:WINDOWS_CERT_PASSWORD /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 $BinaryPath

Write-Output "Windows signing is currently disabled. Skipping: $BinaryPath"
