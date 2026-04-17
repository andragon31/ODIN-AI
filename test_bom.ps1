$bytes = [System.IO.File]::ReadAllBytes($env:TEMP + "\test_install.ps1")
Write-Host "First 4 bytes: $($bytes[0]),$($bytes[1]),$($bytes[2]),$($bytes[3])"
if ($bytes[0] -eq 0xEF -and $bytes[1] -eq 0xBB -and $bytes[2] -eq 0xBF) {
    Write-Host "UTF-8 BOM detected"
} elseif ($bytes[0] -eq 0xFF -and $bytes[1] -eq 0xFE) {
    Write-Host "UTF-16 LE BOM detected"
} elseif ($bytes[0] -eq 0xFE -and $bytes[1] -eq 0xFF) {
    Write-Host "UTF-16 BE BOM detected"
} else {
    Write-Host "No BOM - raw bytes"
}
Write-Host "Total size: $($bytes.Length)"
Write-Host "Content start:"
Get-Content $env:TEMP"\test_install.ps1" -TotalCount 3