$ProgressPreference = 'SilentlyContinue'
$ErrorActionPreference = 'Continue'
$url = 'https://raw.githubusercontent.com/andragon31/ODIN-AI/ab8b110/scripts/install.ps1'
Write-Host "Downloading..."
try {
    $script = Invoke-WebRequest -Uri $url -UseBasicParsing
    Write-Host "Downloaded $($script.Content.Length) bytes"
    Write-Host "First line: $($script.Content.Split("`n")[0])"
} catch {
    Write-Host "Error: $($_.Exception.Message)"
}