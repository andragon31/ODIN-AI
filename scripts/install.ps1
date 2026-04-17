#Requires -Version 5.1
<#
.SYNOPSIS
    ODIN AI - Instalador para Windows
    Ecosistema nórdico local-first desarrollado por Gentleman Programming.

.DESCRIPTION
    Descarga e instala el binario de ODIN para Windows.
    Soporta instalación vía Go o binario pre-compilado desde GitHub Releases.

.EXAMPLE
    # Ejecutar directamente:
    irm https://raw.githubusercontent.com/andragon31/ODIN-AI/main/scripts/install.ps1 | iex

    # O descargar y ejecutar:
    Invoke-WebRequest -Uri https://raw.githubusercontent.com/andragon31/ODIN-AI/main/scripts/install.ps1 -OutFile install.ps1
    .\install.ps1

    # Forzar método específico:
    .\install.ps1 -Method binary
    .\install.ps1 -Method go
#>

[CmdletBinding()]
param(
    [ValidateSet("auto", "go", "binary")]
    [string]$Method = "auto",

    [string]$InstallDir = ""
)

$ErrorActionPreference = "Stop"

$GITHUB_OWNER = "andragon31"
$GITHUB_REPO = "ODIN-AI"
$BINARY_NAME = "odin"

# ============================================================================
# Logging helpers
# ============================================================================

function Write-Info    { param([string]$Message) Write-Host "[info]    $Message" -ForegroundColor Blue }
function Write-Success { param([string]$Message) Write-Host "[ok]      $Message" -ForegroundColor Green }
function Write-Warn    { param([string]$Message) Write-Host "[warn]    $Message" -ForegroundColor Yellow }
function Write-Err    { param([string]$Message) Write-Host "[error]   $Message" -ForegroundColor Red }
function Write-Step   { param([string]$Message) Write-Host "`n==> $Message" -ForegroundColor Cyan }

function Stop-WithError {
    param([string]$Message)
    Write-Err $Message
    exit 1
}

# ============================================================================
# Banner
# ============================================================================

function Show-Banner {
    Write-Host ""
    Write-Host "   ____  ____  ___ _   _ " -ForegroundColor Cyan
    Write-Host "  / __ \|  _ \|_ _| \ | |" -ForegroundColor Cyan
    Write-Host " | |  | | | | || ||  \| |" -ForegroundColor Cyan
    Write-Host " | |__| | |_| || || |\  |" -ForegroundColor Cyan
    Write-Host "  \____/|____/|___|_| \_|" -ForegroundColor Cyan
    Write-Host "   A I   E C O S Y S T E M   " -ForegroundColor DarkGray
    Write-Host ""
    Write-Host "  El orquestador nórdico local-first para el desarrollo spec-driven." -ForegroundColor DarkGray
    Write-Host ""
}

# ============================================================================
# Platform detection
# ============================================================================

function Get-Platform {
    Write-Step "Detecting platform"

    $arch = if ([Environment]::Is64BitOperatingSystem) {
        if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
    } else {
        Stop-WithError "32-bit Windows is not supported."
    }

    Write-Success "Platform: Windows ($arch)"
    return $arch
}

# ============================================================================
# Prerequisites
# ============================================================================

function Test-Prerequisites {
    Write-Step "Checking prerequisites"

    $missing = @()
    if (-not (Get-Command "git" -ErrorAction SilentlyContinue)) { $missing += "git" }

    if ($missing.Count -gt 0) {
        Stop-WithError "Missing required tools: $($missing -join ', '). Please install them and try again."
    }

    Write-Success "git is available"
}

# ============================================================================
# Install method detection
# ============================================================================

function Get-InstallMethod {
    param([string]$Forced)

    if ($Forced -ne "auto") {
        Write-Info "Using forced method: $Forced"
        return $Forced
    }

    Write-Info "Will use binary with source fallback"
    return "binary"
}

# ============================================================================
# Install via go install
# ============================================================================

function Install-ViaGo {
    Write-Step "Installing via go install"

    $goPackage = "github.com/$($GITHUB_OWNER.ToLower())/$GITHUB_REPO/cmd/$BINARY_NAME@latest"
    Write-Info "Running: go install $goPackage"

    & go install $goPackage
    if ($LASTEXITCODE -ne 0) {
        Stop-WithError "Failed to install via go install. Make sure Go is properly configured."
    }

    $gobin = & go env GOBIN 2>$null
    if (-not $gobin) {
        $gopath = & go env GOPATH 2>$null
        $gobin = Join-Path $gopath "bin"
    }

    if ($env:PATH -notlike "*$gobin*") {
        Write-Warn "$gobin is not in your PATH"
        Write-Warn "Add it to your PATH environment variable."
    }

    Write-Success "Installed $BINARY_NAME via go install"
}

# ============================================================================
# Install via binary download
# ============================================================================

function Get-LatestVersion {
    Write-Info "Fetching latest release from GitHub..."

    $url = "https://api.github.com/repos/$GITHUB_OWNER/$GITHUB_REPO/releases/latest"

    try {
        $response = Invoke-RestMethod -Uri $url -Headers @{ "User-Agent" = "odin-installer" } -TimeoutSec 15
    } catch {
        return $null
    }

    $version = $response.tag_name
    if (-not $version) {
        return $null
    }

    Write-Success "Latest version: $version"
    return $version
}

function Install-ViaBinary {
    param([string]$Arch)

    Write-Step "Installing pre-built binary"

    $version = Get-LatestVersion

    if (-not $version) {
        Write-Warn "No GitHub release found. Building from source..."
        Install-ViaSource
        return
    }

    $versionNumber = $version.TrimStart("v")

    $archiveName = "${BINARY_NAME}_${versionNumber}_windows_${Arch}.zip"
    $downloadUrl = "https://github.com/$GITHUB_OWNER/$GITHUB_REPO/releases/download/$version/$archiveName"
    $checksumsUrl = "https://github.com/$GITHUB_OWNER/$GITHUB_REPO/releases/download/$version/checksums.txt"

    $tmpDir = Join-Path $env:TEMP "odin-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    try {
        Write-Info "Downloading $archiveName..."
        $archivePath = Join-Path $tmpDir $archiveName
        Invoke-WebRequest -Uri $downloadUrl -OutFile $archivePath -UseBasicParsing -TimeoutSec 60

        $fileSize = (Get-Item $archivePath).Length
        if ($fileSize -lt 1000) {
            Write-Warn "Downloaded file too small. Building from source..."
            Install-ViaSource
            return
        }
        Write-Success "Downloaded $archiveName ($fileSize bytes)"

        Write-Info "Verifying checksum..."
        try {
            $checksumsPath = Join-Path $tmpDir "checksums.txt"
            Invoke-WebRequest -Uri $checksumsUrl -OutFile $checksumsPath -UseBasicParsing

            $checksums = Get-Content $checksumsPath
            $expectedLine = $checksums | Where-Object { $_ -match $archiveName }
            if ($expectedLine) {
                $expectedChecksum = ($expectedLine -split "\s+")[0]
                $actualChecksum = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLower()

                if ($actualChecksum -ne $expectedChecksum) {
                    Write-Warn "Checksum mismatch. Building from source..."
                    Install-ViaSource
                    return
                }
                Write-Success "Checksum verified"
            } else {
                Write-Warn "Archive not found in checksums.txt - skipping verification"
            }
        } catch {
            Write-Warn "Could not download checksums.txt - skipping verification"
        }

        Write-Info "Extracting $BINARY_NAME..."
        Expand-Archive -Path $archivePath -DestinationPath $tmpDir -Force

        $binaryPath = Join-Path $tmpDir "$BINARY_NAME.exe"
        if (-not (Test-Path $binaryPath)) {
            $binaryPath = Join-Path $tmpDir "bin\$BINARY_NAME.exe"
        }
        if (-not (Test-Path $binaryPath)) {
            Write-Warn "Binary not found in archive. Building from source..."
            Install-ViaSource
            return
        }

        $installDir = $InstallDir
        if (-not $installDir) {
            $installDir = Join-Path $env:LOCALAPPDATA "odin\bin"
        }

        if (-not (Test-Path $installDir)) {
            New-Item -ItemType Directory -Path $installDir -Force | Out-Null
        }

        $destPath = Join-Path $installDir "$BINARY_NAME.exe"
        Write-Info "Installing to $destPath..."
        Copy-Item -Path $binaryPath -Destination $destPath -Force

        Write-Success "Installed $BINARY_NAME to $destPath"

        if ($env:PATH -notlike "*$installDir*") {
            Write-Warn "$installDir is not in your PATH"
            Write-Host ""
            Write-Warn "Run this to add it permanently:"
            Write-Host "  [Environment]::SetEnvironmentVariable('PATH', `$env:PATH + ';$installDir', 'User')" -ForegroundColor DarkGray
            Write-Host ""
        }
    } finally {
        Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

function Install-ViaSource {
    Write-Step "Building from source"

    if (-not (Get-Command "go" -ErrorAction SilentlyContinue)) {
        Stop-WithError "Go is required to build from source. Install Go from https://go.dev/dl/"
    }

    $INSTALL_ROOT = Join-Path $HOME ".odin"
    $INSTALL_BIN = Join-Path $INSTALL_ROOT "bin"

    if (-not (Test-Path $INSTALL_BIN)) {
        New-Item -ItemType Directory -Path $INSTALL_BIN -Force | Out-Null
    }

    $sourcePath = "./cmd/odin"
    $binaryPath = Join-Path $INSTALL_BIN "$BINARY_NAME.exe"

    Write-Info "Compiling..."
    & go build -o $binaryPath $sourcePath
    if ($LASTEXITCODE -ne 0) {
        Stop-WithError "Failed to compile ODIN. Check that source exists at $sourcePath"
    }

    Write-Success "Binary compiled and installed at $binaryPath"

    if ($env:PATH -notlike "*$INSTALL_BIN*") {
        $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
        $newPath = "$userPath;$INSTALL_BIN"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Write-Warn "Added $INSTALL_BIN to PATH. Restart your terminal."
    }
}

# ============================================================================
# Verify installation
# ============================================================================

function Test-Installation {
    Write-Step "Verifying installation"

    $env:PATH = [Environment]::GetEnvironmentVariable("PATH", "Machine") + ";" + [Environment]::GetEnvironmentVariable("PATH", "User")

    $cmd = Get-Command $BINARY_NAME -ErrorAction SilentlyContinue
    if ($cmd) {
        $versionOutput = & $BINARY_NAME version 2>&1
        Write-Success "$BINARY_NAME is installed: $versionOutput"
        return
    }

    $gopath = $null
    if (Get-Command "go" -ErrorAction SilentlyContinue) {
        $gopath = & go env GOPATH 2>$null
    }
    $locations = @(
        (Join-Path $env:LOCALAPPDATA "odin\bin\$BINARY_NAME.exe")
    )
    if ($gopath) {
        $locations += (Join-Path $gopath "bin\$BINARY_NAME.exe")
    }

    foreach ($loc in $locations) {
        if ($loc -and (Test-Path $loc)) {
            $versionOutput = & $loc version 2>&1
            Write-Success "Found $BINARY_NAME at $loc`: $versionOutput"
            Write-Warn "Binary location is not in your PATH. Add it to use '$BINARY_NAME' directly."
            return
        }
    }

    Write-Warn "Could not verify installation. You may need to restart your terminal."
}

# ============================================================================
# Next steps
# ============================================================================

function Show-NextSteps {
    Write-Host ""
    Write-Host "Installation complete!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor White
    Write-Host "  1. Run 'odin init' to initialize the orchestrator" -ForegroundColor Cyan
    Write-Host "  2. Launch 'odin tui' for interactive mode" -ForegroundColor Cyan
    Write-Host "  3. Configure your models with 'odin router selection'" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "For help: $BINARY_NAME --help" -ForegroundColor DarkGray
    Write-Host "Docs:     https://github.com/$GITHUB_OWNER/$GITHUB_REPO" -ForegroundColor DarkGray
    Write-Host ""
}

# ============================================================================
# Main
# ============================================================================

function Main {
    Show-Banner

    $arch = Get-Platform
    Test-Prerequisites

    $installMethod = Get-InstallMethod -Forced $Method

    switch ($installMethod) {
        "go"     { Install-ViaGo }
        "binary" { Install-ViaBinary -Arch $arch }
    }

    Test-Installation
    Show-NextSteps
}

Main
