param(
  [Parameter(Mandatory = $true)]
  [string]$BinaryPath
)

$ErrorActionPreference = "Stop"

# Cross-compiled Windows binaries are signed only when this hook runs on Windows
# with Authenticode credentials available.
if (-not $IsWindows) {
  Write-Output "Not running on Windows; skipping Authenticode signing for $BinaryPath"
  exit 0
}

if (-not $env:WINDOWS_CERT_BASE64 -or -not $env:WINDOWS_CERT_PASSWORD) {
  Write-Output "WINDOWS_CERT_BASE64 / WINDOWS_CERT_PASSWORD not set; skipping Authenticode signing for $BinaryPath"
  exit 0
}

$certPath = Join-Path ([System.IO.Path]::GetTempPath()) ("sumup-windows-" + [guid]::NewGuid().ToString() + ".pfx")
try {
  [IO.File]::WriteAllBytes($certPath, [Convert]::FromBase64String($env:WINDOWS_CERT_BASE64))

  $signtool = $null
  $candidates = @(
    "${env:ProgramFiles(x86)}\Windows Kits\10\bin\*\x64\signtool.exe",
    "${env:ProgramFiles}\Windows Kits\10\bin\*\x64\signtool.exe"
  )
  foreach ($pattern in $candidates) {
    $match = Get-Item $pattern -ErrorAction SilentlyContinue | Sort-Object FullName -Descending | Select-Object -First 1
    if ($match) {
      $signtool = $match.FullName
      break
    }
  }
  if (-not $signtool) {
    $signtool = Get-Command signtool.exe -ErrorAction SilentlyContinue | Select-Object -ExpandProperty Source
  }
  if (-not $signtool) {
    throw "signtool.exe not found; install Windows SDK signing tools"
  }

  & $signtool sign `
    /f $certPath `
    /p $env:WINDOWS_CERT_PASSWORD `
    /fd SHA256 `
    /tr http://timestamp.digicert.com `
    /td SHA256 `
    $BinaryPath
  if ($LASTEXITCODE -ne 0) {
    throw "signtool failed with exit code $LASTEXITCODE"
  }

  Write-Output "Signed $BinaryPath"
}
finally {
  if (Test-Path $certPath) {
    Remove-Item -Force $certPath
  }
}
