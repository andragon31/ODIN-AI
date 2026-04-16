# ODIN AI - Windows Installer
# One-liner: irm https://raw.githubusercontent.com/andragon31/ODIN-AI/main/install.ps1 | iex

param(
    [string]$Version = "1.0.0",
    [string]$InstallPath = "$env:LOCALAPPDATA\Programs\ODIN",
    [switch]$AddToPath,
    [switch]$DryRun,
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"

# Colors
function Write-Info { param([string]$Msg) Write-Host "[INFO] $Msg" -ForegroundColor Cyan }
function Write-Success { param([string]$Msg) Write-Host "[SUCCESS] $Msg" -ForegroundColor Green }
function Write-Warn { param([string]$Msg) Write-Host "[WARN] $Msg" -ForegroundColor Yellow }
function Write-Err { param([string]$Msg) Write-Host "[ERROR] $Msg" -ForegroundColor Red }

# Banner
function Show-Banner {
    Write-Host ""
    Write-Host "  ██╗    ██╗██╗██╗  ██╗██╗    ██╗ █████╗ ███████╗" -ForegroundColor Green
    Write-Host "  ██║    ██║██║██║ ██╔╝██║    ██║██╔══██╗██╔════╝" -ForegroundColor Green
    Write-Host "  ██║ █╗ ██║██║█████╔╝ ██║    ██║███████║███████╗" -ForegroundColor Green
    Write-Host "  ██║███╗██║██║██╔══██╗ ██║    ██║██╔══██║╚════██║" -ForegroundColor Green
    Write-Host "  ╚███╔███╔╝██║██║  ██║██║    ██║██║  ██║███████║" -ForegroundColor Green
    Write-Host "   ╚══╝╚══╝ ╚═╝╚═╝  ╚═╝╚═╝    ╚═╝╚═╝  ╚═╝╚══════╝" -ForegroundColor Green
    Write-Host ""
    Write-Host "  Local-First AI Ecosystem · 100% OSS" -ForegroundColor White
    Write-Host ""
}

# Detect architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { return "amd64" }
    }
}

# Detect OS
function Get-OS {
    if ($IsWindows) { return "windows" }
    return "windows"
}

# Get latest release info from GitHub
function Get-LatestRelease {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/andragon31/ODIN-AI/releases/latest" -TimeoutSec 10
        return @{
            TagName = $response.tag_name -replace "v", ""
            Assets = $response.assets
        }
    } catch {
        Write-Warn "Could not fetch latest release, using default version: $Version"
        return @{
            TagName = $Version
            Assets = @()
        }
    }
}

# Download file
function Download-File {
    param(
        [string]$Url,
        [string]$OutFile
    )

    Write-Info "Downloading from $Url..."
    try {
        Invoke-WebRequest -Uri $Url -OutFile $OutFile -UseBasicParsing
    } catch {
        Write-Err "Download failed: $_"
        throw
    }
}

# Verify checksum
function Verify-Checksum {
    param(
        [string]$File,
        [string]$Expected
    )

    if (-not $Expected) {
        Write-Warn "No checksum provided, skipping verification"
        return $true
    }

    Write-Info "Verifying checksum..."
    $hash = (Get-FileHash -Path $File -Algorithm SHA256).Hash.ToLower()

    if ($hash -eq $Expected.ToLower()) {
        Write-Success "Checksum verified!"
        return $true
    } else {
        Write-Err "Checksum mismatch! Expected: $Expected, Got: $hash"
        return $false
    }
}

# Create directories
function Initialize-Directories {
    $configDir = "$env:USERPROFILE\.odin"

    Write-Info "Creating directories..."
    New-Item -ItemType Directory -Force -Path $configDir | Out-Null
    New-Item -ItemType Directory -Force -Path "$configDir\rules" | Out-Null
    New-Item -ItemType Directory -Force -Path "$configDir\themes" | Out-Null
    New-Item -ItemType Directory -Force -Path "$configDir\plugins" | Out-Null
    New-Item -ItemType Directory -Force -Path "$configDir\sessions" | Out-Null
    New-Item -ItemType Directory -Force -Path "$configDir\logs" | Out-Null
    New-Item -ItemType Directory -Force -Path "$configDir\backups" | Out-Null

    Write-Success "Directories created at $configDir"
}

# Add to PATH
function Add-ToPath {
    param([string]$Path)

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -like "*$Path*") {
        Write-Info "ODIN bin directory already in PATH"
        return
    }

    $newPath = "$userPath;$Path"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Success "Added $Path to PATH"
    Write-Info "Please restart your terminal or run: refreshenv"
}

# Install ODIN
function Install-ODIN {
    param(
        [string]$Version,
        [string]$InstallPath
    )

    Show-Banner
    Write-Info "Installing ODIN AI v$Version..."

    $os = Get-OS
    $arch = Get-Architecture
    $archiveName = "odin-$os-$arch.tar.gz"
    $exeName = "odin.exe"

    # Get release info
    $release = Get-LatestRelease
    $Version = $release.TagName

    Write-Info "Target: $os/$arch"
    Write-Info "Version: $Version"

    # Create temp directory
    $tempDir = Join-Path $env:TEMP "odin-install-$(Get-Random)"
    New-Item -ItemType Directory -Force -Path $tempDir | Out-Null

    try {
        # Download URL
        $downloadUrl = "https://github.com/andragon31/ODIN-AI/releases/download/v$Version/$archiveName"

        # If no pre-built binary, build locally
        $shouldBuild = $true
        try {
            $response = Invoke-WebRequest -Uri $downloadUrl -UseBasicParsing -TimeoutSec 5
            if ($response.StatusCode -eq 200) {
                $shouldBuild = $false
            }
        } catch {
            Write-Warn "Pre-built binary not found, will try to build locally"
        }

        if ($shouldBuild) {
            Write-Info "Pre-built binary not available for this version"
            Write-Info "Please use: git clone https://github.com/andragon31/ODIN-AI.git"
            Write-Info "Then run: go build -o odin.exe ./cmd/odin"
            return
        }

        # Download
        $archivePath = Join-Path $tempDir $archiveName
        Download-File -Url $downloadUrl -OutFile $archivePath

        # Create install directory
        New-Item -ItemType Directory -Force -Path $InstallPath | Out-Null

        # Extract
        Write-Info "Extracting..."
        tar -xzf $archivePath -C $InstallPath

        # Make executable
        $exePath = Join-Path $InstallPath $exeName
        if (Test-Path $exePath) {
            Write-Success "Binary installed at $exePath"
        } else {
            # Try finding the exe
            $found = Get-ChildItem -Path $InstallPath -Filter "*.exe" -Recurse | Select-Object -First 1
            if ($found) {
                Write-Success "Binary installed at $($found.FullName)"
                $exePath = $found.FullName
            }
        }

        # Initialize directories
        Initialize-Directories

        # Add to PATH
        if ($AddToPath -or ([System.SecurityPrincipal.WindowsPrincipal][System.SecurityPrincipal.WindowsIdentity]::GetCurrent()).IsInRole([System.SecurityPrincipal.WindowsBuiltInRole]::Administrator)) {
            Add-ToPath -Path $InstallPath
        } else {
            Write-Warn "Not adding to PATH (requires admin or -AddToPath flag)"
        }

        Write-Success "ODIN AI v$Version installed successfully!"
        Write-Host ""
        Write-Host "Run 'odin --help' to get started" -ForegroundColor White
        Write-Host "Or use: $exePath --help" -ForegroundColor White

    } finally {
        # Cleanup
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# Uninstall ODIN
function Uninstall-ODIN {
    Show-Banner
    Write-Warn "Uninstalling ODIN AI..."

    $exePath = Join-Path $InstallPath "odin.exe"
    if (Test-Path $exePath) {
        Remove-Item -Path $InstallPath -Recurse -Force
        Write-Success "ODIN uninstalled from $InstallPath"
    } else {
        Write-Err "ODIN not found at $InstallPath"
    }

    Write-Info "Configuration files kept at $env:USERPROFILE\.odin"
    Write-Info "To remove all data: Remove-Item -Recurse $env:USERPROFILE\.odin"
}

# Dry run
function Dry-Run {
    Show-Banner
    $os = Get-OS
    $arch = Get-Architecture

    Write-Info "Dry run - would install:"
    Write-Host "  Version:    $Version"
    Write-Host "  OS:         $os"
    Write-Host "  Arch:       $arch"
    Write-Host "  Install to:  $InstallPath"
    Write-Host "  Config at:  $env:USERPROFILE\.odin"
    Write-Host ""
    Write-Info "Download URL would be:"
    Write-Host "  https://github.com/andragon31/ODIN-AI/releases/download/v$Version/odin-$os-$arch.tar.gz"
}

# Main
function Main {
    if ($Uninstall) {
        Uninstall-ODIN
    } elseif ($DryRun) {
        Dry-Run
    } else {
        Install-ODIN -Version $Version -InstallPath $InstallPath
    }
}

Main
