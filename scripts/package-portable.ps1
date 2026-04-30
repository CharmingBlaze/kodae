$ErrorActionPreference = "Stop"

param(
  [Parameter(Mandatory = $true)][string]$KodaeBinary,
  [Parameter(Mandatory = $true)][string]$Platform,
  [string]$OutDir = "dist"
)

$bundleRoot = Join-Path $OutDir ("kodae-" + $Platform)
$binDir = Join-Path $bundleRoot "bin"

New-Item -ItemType Directory -Force -Path $binDir | Out-Null
Copy-Item -LiteralPath $KodaeBinary -Destination (Join-Path $binDir "kodae.exe") -Force

$zipPath = Join-Path $OutDir ("kodae-" + $Platform + ".zip")
if (Test-Path $zipPath) {
  Remove-Item -LiteralPath $zipPath -Force
}
Compress-Archive -Path $bundleRoot -DestinationPath $zipPath -Force
Write-Host "portable bundle written: $zipPath"
