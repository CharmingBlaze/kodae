$ErrorActionPreference = "Stop"

param(
  [Parameter(Mandatory = $true)][string]$ClioBinary,
  [Parameter(Mandatory = $true)][string]$ZigBinary,
  [Parameter(Mandatory = $true)][string]$Platform,
  [string]$OutDir = "dist"
)

$bundleRoot = Join-Path $OutDir ("clio-" + $Platform)
$binDir = Join-Path $bundleRoot "bin"
$toolchainDir = Join-Path $bundleRoot "toolchain\zig"

New-Item -ItemType Directory -Force -Path $binDir | Out-Null
New-Item -ItemType Directory -Force -Path $toolchainDir | Out-Null

Copy-Item -LiteralPath $ClioBinary -Destination (Join-Path $binDir "clio.exe") -Force
Copy-Item -LiteralPath $ZigBinary -Destination (Join-Path $toolchainDir "zig.exe") -Force

$zipPath = Join-Path $OutDir ("clio-" + $Platform + ".zip")
if (Test-Path $zipPath) {
  Remove-Item -LiteralPath $zipPath -Force
}
Compress-Archive -Path $bundleRoot -DestinationPath $zipPath -Force
Write-Host "portable bundle written: $zipPath"
