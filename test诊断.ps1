$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

$BINARY_NAME = "odin"
$GITHUB_OWNER = "andragon31"
$GITHUB_REPO = "ODIN-AI"

Write-Host ""
Write-Host "   ODIN Installer Test" -ForegroundColor Cyan
Write-Host ""

try {
    Write-Host "Testing GitHub API..."
    $url = "https://api.github.com/repos/$GITHUB_OWNER/$GITHUB_REPO/releases/latest"
    $response = Invoke-RestMethod -Uri $url -Headers @{ "User-Agent" = "odin-installer" } -TimeoutSec 15
    Write-Host "API Response: $($response.tag_name)"
} catch {
    Write-Host "API Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")