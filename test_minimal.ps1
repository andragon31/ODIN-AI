#Requires -Version 5.1
$ErrorActionPreference = "Stop"
$BINARY_NAME = "odin"
$GITHUB_OWNER = "andragon31"
$GITHUB_REPO = "ODIN-AI"

Write-Host ""
Write-Host "   ODIN Installer" -ForegroundColor Cyan
Write-Host ""

Write-Host "Testing..."
Write-Host "BINARY_NAME=$BINARY_NAME"
Write-Host "GITHUB_OWNER=$GITHUB_OWNER"

$missing = @()
if (-not (Get-Command "git" -ErrorAction SilentlyContinue)) { $missing += "git" }
if ($missing.Count -gt 0) {
    Write-Host "Missing: $($missing -join ', ')" -ForegroundColor Red
} else {
    Write-Host "git found" -ForegroundColor Green
}

Write-Host "Done"