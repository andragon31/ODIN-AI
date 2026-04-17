$ProgressPreference = 'SilentlyContinue'
$url = 'https://raw.githubusercontent.com/andragon31/ODIN-AI/ab8b110/scripts/install.ps1'
$r = Invoke-WebRequest -Uri $url -UseBasicParsing
$r.Content | Select-Object -First 10