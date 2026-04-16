# ODIN AI - Windows Installer
# One-liner: irm https://raw.githubusercontent.com/andragon31/ODIN-AI/main/install.ps1 | iex

param(
    [string]$Version = "1.0.0",
    [string]$InstallPath = "$env:LOCALAPPDATA\Programs\ODIN",
    [switch]$DryRun,
    [switch]$Uninstall,
    [switch]$SkipSetup
)

$ErrorActionPreference = "Stop"

# Colors
function Write-Info { param([string]$Msg) Write-Host "[INFO] $Msg" -ForegroundColor Cyan }
function Write-Success { param([string]$Msg) Write-Host "[SUCCESS] $Msg" -ForegroundColor Green }
function Write-Warn { param([string]$Msg) Write-Host "[WARN] $Msg" -ForegroundColor Yellow }
function Write-Err { param([string]$Msg) Write-Host "[ERROR] $Msg" -ForegroundColor Red }
function Write-Step { param([string]$Num, [string]$Msg) Write-Host "`n  [$Num] $Msg" -ForegroundColor Magenta }
function Write-Prompt { param([string]$Msg) Write-Host $Msg -ForegroundColor Yellow }

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
    $env:Path = $newPath  # Update current session
    Write-Success "Added $Path to PATH"
}

# Get ODIN executable path
function Get-OdinPath {
    param([string]$InstallPath)
    $exePath = Join-Path $InstallPath "odin.exe"
    if (Test-Path $exePath) {
        return $exePath
    }
    # Try finding the exe
    $found = Get-ChildItem -Path $InstallPath -Filter "*.exe" -Recurse | Select-Object -First 1
    return $found.FullName
}

# Run ODIN command
function Invoke-Odin {
    param([string]$Args, [string]$InstallPath)
    $exePath = Get-OdinPath -InstallPath $InstallPath
    if ($exePath) {
        & $exePath $Args
    } else {
        Write-Err "ODIN executable not found"
    }
}

# ============================================================================
# POST-INSTALLATION SETUP WIZARD
# ============================================================================

function Show-SetupWizard {
    param([string]$InstallPath)

    $exePath = Get-OdinPath -InstallPath $InstallPath

    Write-Host ""
    Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Magenta
    Write-Host "                    CONFIGURACIÓN INICIAL" -ForegroundColor Magenta
    Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Magenta

    Write-Step "1" "Inicializando ODIN..."
    Write-Host "   Ejecutando: odin init"
    & $exePath init --quiet 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Success "   ODIN inicializado correctamente"
    } else {
        Write-Warn "   ODIN init completado (puede que ya existiera)"
    }

    Write-Step "2" "Verificando instalación..."
    Write-Host "   Ejecutando: odin --version"
    $versionOutput = & $exePath --version 2>$null
    Write-Host "   Versión: $versionOutput" -ForegroundColor White

    Write-Step "3" "Configurando Router (Provider de IA)..."
    Write-Host ""
    Write-Host "   ODIN soporta múltiples providers de IA:" -ForegroundColor White
    Write-Host "   1. Ollama (local, gratis) - Recommended" -ForegroundColor Cyan
    Write-Host "   2. OpenRouter (API, económico)" -ForegroundColor Cyan
    Write-Host "   3. Anthropic (API, premium)" -ForegroundColor Cyan
    Write-Host ""
    Write-Prompt "   ¿Tienes Ollama instalado localmente? (S/n)"
    $response = Read-Host "   "
    if ($response -ne "n" -and $response -ne "N") {
        Write-Info "   Configurando Ollama como provider principal..."
        & $exePath router set ollama --quiet 2>$null
        Write-Success "   Ollama configurado"
    } else {
        Write-Info "   Puedes configurar después con: odin router set <provider>"
    }

    Write-Step "4" "Configurando Heimdall (Seguridad)..."
    Write-Host "   Heimdall proporciona análisis de seguridad SAST" -ForegroundColor White
    Write-Host "   Para instalar reglas personalizadas:" -ForegroundColor White
    Write-Host "   odin heimdall hook-install" -ForegroundColor Cyan

    Write-Step "5" "Sincronización con Bifrost..."
    Write-Host "   Bifrost permite sincronizar tu configuración" -ForegroundColor White
    Write-Host "   Para inicializar: odin sync init" -ForegroundColor Cyan

    Write-Host ""
    Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Magenta
    Write-Host "                    ¡INSTALACIÓN COMPLETA!" -ForegroundColor Magenta
    Write-Host "════════════════════════════════════════════════════════════" -ForegroundColor Magenta

    Write-Host ""
    Write-Host "  Comandos disponibles:" -ForegroundColor White
    Write-Host ""
    Write-Host "  odin status              Ver estado del sistema" -ForegroundColor Green
    Write-Host "  odin router status       Ver providers configurados" -ForegroundColor Green
    Write-Host "  odin mimir --help        Gestionar memoria" -ForegroundColor Green
    Write-Host "  odin heimdall --help      Seguridad" -ForegroundColor Green
    Write-Host "  odin sync --help          Sincronización" -ForegroundColor Green
    Write-Host "  odin theme --help         Temas" -ForegroundColor Green
    Write-Host ""
    Write-Host "  Documentación: https://github.com/andragon31/ODIN-AI" -ForegroundColor Gray
    Write-Host ""
    Write-Prompt "  Presiona ENTER para continuar..."
    Read-Host
}

# ============================================================================
# INSTALL ODIN
# ============================================================================

function Install-ODIN {
    param(
        [string]$Version,
        [string]$InstallPath
    )

    Show-Banner
    Write-Info "Instalando ODIN AI v$Version..."
    Write-Info "Instalando en: $InstallPath"

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

        # Check if pre-built binary exists
        $shouldBuild = $true
        try {
            $response = Invoke-WebRequest -Uri $downloadUrl -UseBasicParsing -TimeoutSec 5
            if ($response.StatusCode -eq 200) {
                $shouldBuild = $false
            }
        } catch {
            Write-Warn "Pre-built binary not found"
        }

        if ($shouldBuild) {
            Write-Host ""
            Write-Warn "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
            Write-Warn "  ¡ATENCIÓN! No hay binario pre-compilado para esta versión" -ForegroundColor Yellow
            Write-Warn "════════════════════════════════════════════════════════════" -ForegroundColor Yellow
            Write-Host ""
            Write-Host "  Para instalar desde código fuente:" -ForegroundColor White
            Write-Host ""
            Write-Host "  1. git clone https://github.com/andragon31/ODIN-AI.git" -ForegroundColor Cyan
            Write-Host "  2. cd ODIN-AI" -ForegroundColor Cyan
            Write-Host "  3. go build -o odin.exe ./cmd/odin" -ForegroundColor Cyan
            Write-Host ""
            Write-Host "  O espera a que se publiquen los binarios en:" -ForegroundColor White
            Write-Host "  https://github.com/andragon31/ODIN-AI/releases" -ForegroundColor Cyan
            Write-Host ""
            return
        }

        # Download
        $archivePath = Join-Path $tempDir $archiveName
        Download-File -Url $downloadUrl -OutFile $archivePath

        # Create install directory
        New-Item -ItemType Directory -Force -Path $InstallPath | Out-Null

        # Extract
        Write-Info "Extrayendo..."
        tar -xzf $archivePath -C $InstallPath

        # Find executable
        $exePath = Get-OdinPath -InstallPath $InstallPath
        if ($exePath) {
            Write-Success "Binario instalado en: $exePath"
        } else {
            Write-Err "No se pudo encontrar el ejecutable"
            return
        }

        # Initialize directories
        Initialize-Directories

        # Add to PATH (automatically)
        Add-ToPath -Path $InstallPath

        Write-Success "ODIN AI v$Version instalado correctamente!"
        Write-Host ""

        # Run setup wizard unless skipped
        if (-not $SkipSetup) {
            Show-SetupWizard -InstallPath $InstallPath
        }

    } finally {
        # Cleanup
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# ============================================================================
# UNINSTALL
# ============================================================================

function Uninstall-ODIN {
    Show-Banner
    Write-Warn "Desinstalando ODIN AI..."

    $exePath = Join-Path $InstallPath "odin.exe"
    if (Test-Path $exePath) {
        Remove-Item -Path $InstallPath -Recurse -Force
        Write-Success "ODIN desinstalado de $InstallPath"
    } else {
        Write-Err "ODIN no encontrado en $InstallPath"
    }

    Write-Info "Archivos de configuración保留 en $env:USERPROFILE\.odin"
    Write-Info "Para eliminar todo: Remove-Item -Recurse $env:USERPROFILE\.odin"
}

# ============================================================================
# DRY RUN
# ============================================================================

function Dry-Run {
    Show-Banner
    $os = Get-OS
    $arch = Get-Architecture

    Write-Info "Dry run - instalación simulada:"
    Write-Host "  Versión:     $Version"
    Write-Host "  SO:          $os"
    Write-Host "  Arquitectura:$arch"
    Write-Host "  Instalar en: $InstallPath"
    Write-Host "  Config en:   $env:USERPROFILE\.odin"
    Write-Host ""
    Write-Info "URL de descarga:"
    Write-Host "  https://github.com/andragon31/ODIN-AI/releases/download/v$Version/odin-$os-$arch.tar.gz"
}

# ============================================================================
# MAIN
# ============================================================================

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
