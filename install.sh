#!/usr/bin/env bash
#===============================================================================
# ODIN AI - Installer Script
#===============================================================================
# Norse-themed local-first AI ecosystem installer
# Inspired by Gentle AI, powered for local-first
#
# Usage:
#   curl -fsSL https://get.odin.ai/install.sh | bash
#   curl -fsSL https://get.odin.ai/install.sh | bash -s -- --version 1.0.0
#
#===============================================================================

set -e

# Configuration
REPO_NAME="andragon31/ODIN-AI"
INSTALL_DIR="${HOME}/.local/bin"
CONFIG_DIR="${HOME}/.odin"
INSTALL_SCRIPT_VERSION="1.0.0"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Print banner
print_banner() {
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  ██╗    ██╗██╗██╗  ██╗██╗    ██╗ █████╗ ███████╗${NC}"
    echo -e "${GREEN}  ██║    ██║██║██║ ██╔╝██║    ██║██╔══██╗██╔════╝${NC}"
    echo -e "${GREEN}  ██║ █╗ ██║██║█████╔╝ ██║    ██║███████║███████╗${NC}"
    echo -e "${GREEN}  ██║███╗██║██║██╔══██╗ ██║    ██║██╔══██║╚════██║${NC}"
    echo -e "${GREEN}  ╚███╔███╔╝██║██║  ██║██║    ██║██║  ██║███████║${NC}"
    echo -e "${GREEN}   ╚══╝╚══╝ ╚═╝╚═╝  ╚═╝╚═╝    ╚═╝╚═╝  ╚═╝╚══════╝${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
    echo -e "${BLUE}  Local-First AI Ecosystem · 100% OSS · $0 Infra Cost${NC}"
    echo ""
}

# Print usage
print_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help           Show this help message"
    echo "  -v, --version VERSION   Install specific version"
    echo "  -p, --path PATH     Install to custom directory"
    echo "  -d, --dry-run       Show what would be installed"
    echo "  -u, --uninstall     Uninstall ODIN"
    echo "  --skip-checksum     Skip checksum verification"
    echo "  --no-modify-path    Don't add to PATH"
    echo ""
    echo "Examples:"
    echo "  curl -fsSL https://get.odin.ai/install.sh | bash"
    echo "  curl -fsSL https://get.odin.ai/install.sh | bash -s -- -v 1.0.0"
    echo "  ./install.sh -p /custom/path"
}

# Detect OS
detect_os() {
    OS=$(uname -s)
    case "$OS" in
        Linux*)     echo "linux" ;;
        Darwin*)    echo "darwin" ;;
        CYGWIN*)    echo "windows" ;;
        MINGW*)     echo "windows" ;;
        MSYS*)      echo "windows" ;;
        *)          echo "unsupported" ;;
    esac
}

# Detect architecture
detect_arch() {
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64)     echo "amd64" ;;
        aarch64)    echo "arm64" ;;
        arm64)      echo "arm64" ;;
        i386)       echo "386" ;;
        i686)       echo "386" ;;
        *)          echo "unsupported" ;;
    esac
}

# Detect if running in container
detect_container() {
    if [ -f /.dockerenv ] || grep -q docker /proc/1/cgroup 2>/dev/null; then
        echo "docker"
    elif [ -f /run/.containerenv ]; then
        echo "podman"
    elif [ -f /var/run/secrets/kubernetes.io/serviceaccount/token ]; then
        echo "kubernetes"
    else
        echo ""
    fi
}

# Get latest release version
get_latest_version() {
    # In production, this would hit GitHub API
    # For now, return the hardcoded version
    echo "$INSTALL_SCRIPT_VERSION"
}

# Download binary
download_binary() {
    local version=$1
    local os=$2
    local arch=$3
    local tmp_file="/tmp/odin-${version}-${os}-${arch}.tar.gz"

    log_info "Downloading ODIN ${version} for ${os}/${arch}..."

    local download_url="https://github.com/${REPO_NAME}/releases/download/v${version}/odin-${os}-${arch}.tar.gz"

    if command -v curl &> /dev/null; then
        curl -fsSL "$download_url" -o "$tmp_file" || {
            log_error "Failed to download from $download_url"
            return 1
        }
    elif command -v wget &> /dev/null; then
        wget -q "$download_url" -O "$tmp_file" || {
            log_error "Failed to download from $download_url"
            return 1
        }
    else
        log_error "Neither curl nor wget found. Please install one of them."
        return 1
    fi

    echo "$tmp_file"
}

# Extract binary
extract_binary() {
    local archive=$1
    local dest=$2

    log_info "Extracting binary..."

    if command -v tar &> /dev/null; then
        tar -xzf "$archive" -C "$dest" || {
            log_error "Failed to extract archive"
            return 1
        }
    else
        log_error "tar not found. Cannot extract archive."
        return 1
    fi

    log_success "Binary extracted to ${dest}"
}

# Verify checksum (if not skipped)
verify_checksum() {
    local file=$1
    local expected_checksum=$2

    if [ -z "$expected_checksum" ]; then
        log_warn "Checksum not provided, skipping verification"
        return 0
    fi

    log_info "Verifying checksum..."

    if command -v sha256sum &> /dev/null; then
        actual=$(sha256sum "$file" | cut -d' ' -f1)
    elif command -v shasum &> /dev/null; then
        actual=$(shasum -a 256 "$file" | cut -d' ' -f1)
    else
        log_warn "No checksum tool found, skipping verification"
        return 0
    fi

    if [ "$actual" = "$expected_checksum" ]; then
        log_success "Checksum verified!"
    else
        log_error "Checksum mismatch! Expected $expected_checksum, got $actual"
        return 1
    fi
}

# Add to PATH
add_to_path() {
    local bin_path=$1

    # Check if already in PATH
    if [[ ":$PATH:" == *":${bin_path}:"* ]]; then
        log_info "ODIN bin directory already in PATH"
        return 0
    fi

    # Detect shell profile
    local shell_profile=""
    if [ -n "$BASH_VERSION" ]; then
        if [ -f "$HOME/.bashrc" ]; then
            shell_profile="$HOME/.bashrc"
        elif [ -f "$HOME/.bash_profile" ]; then
            shell_profile="$HOME/.bash_profile"
        fi
    elif [ -n "$ZSH_VERSION" ]; then
        if [ -f "$HOME/.zshrc" ]; then
            shell_profile="$HOME/.zshrc"
        fi
    fi

    if [ -n "$shell_profile" ]; then
        echo "" >> "$shell_profile"
        echo "# Added by ODIN AI installer" >> "$shell_profile"
        echo "export PATH=\"\${PATH}:${bin_path}\"" >> "$shell_profile"
        log_success "Added ${bin_path} to PATH in ${shell_profile}"
        log_info "Please restart your shell or run: source ${shell_profile}"
    else
        log_warn "Could not detect shell profile. Manually add ${bin_path} to your PATH."
    fi
}

# Create config directories
create_config_dirs() {
    log_info "Creating configuration directories..."

    mkdir -p "${CONFIG_DIR}"
    mkdir -p "${CONFIG_DIR}/rules"
    mkdir -p "${CONFIG_DIR}/themes"
    mkdir -p "${CONFIG_DIR}/plugins"
    mkdir -p "${CONFIG_DIR}/sessions"
    mkdir -p "${CONFIG_DIR}/logs"
    mkdir -p "${CONFIG_DIR}/backups"

    log_success "Configuration directory created at ${CONFIG_DIR}"
}

# Run post-install setup
post_install() {
    log_info "Running post-install setup..."

    # Initialize ODIN
    if [ -x "${INSTALL_DIR}/odin" ]; then
        "${INSTALL_DIR}/odin" init || log_warn "ODIN init failed, will try again when running"
    fi

    log_success "Post-install complete!"
}

# Install ODIN
do_install() {
    local version="${1:-$INSTALL_SCRIPT_VERSION}"
    local install_path="${2:-$INSTALL_DIR}"
    local skip_checksum="${3:-false}"
    local modify_path="${4:-true}"

    print_banner

    log_info "Installing ODIN AI v${version}"

    # Detect environment
    local os=$(detect_os)
    local arch=$(detect_arch)
    local container=$(detect_container)

    if [ "$os" = "unsupported" ]; then
        log_error "Unsupported operating system: $(uname -s)"
        return 1
    fi

    if [ "$arch" = "unsupported" ]; then
        log_error "Unsupported architecture: $(uname -m)"
        return 1
    fi

    if [ -n "$container" ]; then
        log_warn "Running in ${container} container"
    fi

    # Create install directory
    mkdir -p "$install_path"

    # Download binary
    local archive_file=$(download_binary "$version" "$os" "$arch") || return 1

    # Extract binary
    extract_binary "$archive_file" "$install_path" || return 1

    # Make executable
    chmod +x "${install_path}/odin"

    # Cleanup
    rm -f "$archive_file"

    # Create config directories
    create_config_dirs

    # Add to PATH if requested
    if [ "$modify_path" = "true" ]; then
        add_to_path "$install_path"
    fi

    # Post-install
    post_install

    log_success "ODIN AI v${version} installed successfully!"
    log_info "Run 'odin --help' to get started"
}

# Uninstall ODIN
do_uninstall() {
    print_banner
    log_warn "Uninstalling ODIN AI..."

    read -p "Are you sure? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Uninstall cancelled"
        return 0
    fi

    rm -f "${INSTALL_DIR}/odin"
    log_success "ODIN uninstalled from ${INSTALL_DIR}"
    log_info "Configuration files kept at ${CONFIG_DIR}"
    log_info "To remove all data: rm -rf ${CONFIG_DIR}"
}

# Dry run
do_dry_run() {
    print_banner

    local os=$(detect_os)
    local arch=$(detect_arch)
    local version="${1:-$INSTALL_SCRIPT_VERSION}"

    log_info "Dry run - would install:"
    echo "  Version:    ${version}"
    echo "  OS:         ${os}"
    echo "  Arch:       ${arch}"
    echo "  Install to:  ${INSTALL_DIR}"
    echo "  Config at:  ${CONFIG_DIR}"
    echo ""
    log_info "Download URL would be:"
    echo "  https://github.com/${REPO_NAME}/releases/download/v${version}/odin-${os}-${arch}.tar.gz"
}

# Main entry point
main() {
    local version=""
    local install_path=""
    local dry_run=false
    local uninstall=false
    local skip_checksum=false
    local modify_path=true

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                print_usage
                exit 0
                ;;
            -v|--version)
                version="$2"
                shift 2
                ;;
            -p|--path)
                install_path="$2"
                shift 2
                ;;
            -d|--dry-run)
                dry_run=true
                shift
                ;;
            -u|--uninstall)
                uninstall=true
                shift
                ;;
            --skip-checksum)
                skip_checksum=true
                shift
                ;;
            --no-modify-path)
                modify_path=false
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                print_usage
                exit 1
                ;;
        esac
    done

    if [ "$uninstall" = true ]; then
        do_uninstall
    elif [ "$dry_run" = true ]; then
        do_dry_run "$version"
    else
        do_install "$version" "$install_path" "$skip_checksum" "$modify_path"
    fi
}

# Run main
main "$@"
