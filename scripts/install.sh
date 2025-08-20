#!/bin/bash

# GHCP Memory Context Server Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/tr4d3r/ghcp-memory-context/main/scripts/install.sh | bash
# Or with version: curl -fsSL ... | bash -s v1.0.0

set -euo pipefail

# Configuration
GITHUB_REPO="tr4d3r/ghcp-memory-context"
BINARY_NAME="ghcp-memory-context"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect platform
detect_platform() {
    local os
    local arch

    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          log_error "Unsupported operating system: $(uname -s)" && exit 1 ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        armv6l)         arch="armv6" ;;
        armv7l)         arch="armv7" ;;
        *)              log_error "Unsupported architecture: $(uname -m)" && exit 1 ;;
    esac

    echo "${os}-${arch}"
}

# Get latest release version
get_latest_version() {
    local version
    version=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [[ -z "$version" ]]; then
        log_error "Failed to get latest version"
        exit 1
    fi
    echo "$version"
}

# Download and install binary
install_binary() {
    local version="${1:-$(get_latest_version)}"
    local platform
    platform=$(detect_platform)

    log_info "Installing ${BINARY_NAME} ${version} for ${platform}..."

    # Create temporary directory
    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap "rm -rf ${tmp_dir}" EXIT

    # Determine file extension
    local file_ext="tar.gz"
    if [[ "$platform" == *"windows"* ]]; then
        file_ext="zip"
    fi

    # Download URL
    local download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${BINARY_NAME}-${platform}.${file_ext}"
    local archive_file="${tmp_dir}/${BINARY_NAME}-${platform}.${file_ext}"

    log_info "Downloading from: ${download_url}"

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "${archive_file}" "${download_url}"
    elif command -v wget >/dev/null 2>&1; then
        wget -q -O "${archive_file}" "${download_url}"
    else
        log_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi

    # Extract archive
    log_info "Extracting archive..."
    cd "${tmp_dir}"

    if [[ "$file_ext" == "zip" ]]; then
        unzip -q "${archive_file}"
    else
        tar -xzf "${archive_file}"
    fi

    # Find binary
    local binary_path
    binary_path=$(find "${tmp_dir}" -name "${BINARY_NAME}*" -type f -executable | head -1)

    if [[ -z "$binary_path" ]]; then
        log_error "Binary not found in archive"
        exit 1
    fi

    # Install binary
    log_info "Installing to ${INSTALL_DIR}..."

    if [[ ! -w "$INSTALL_DIR" ]]; then
        log_info "Installing with sudo (${INSTALL_DIR} is not writable)"
        sudo mkdir -p "${INSTALL_DIR}"
        sudo cp "${binary_path}" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        mkdir -p "${INSTALL_DIR}"
        cp "${binary_path}" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    log_success "${BINARY_NAME} ${version} installed successfully!"
}

# Verify installation
verify_installation() {
    if command -v "${BINARY_NAME}" >/dev/null 2>&1; then
        local version
        version=$("${BINARY_NAME}" --version 2>/dev/null || echo "unknown")
        log_success "Verification: ${BINARY_NAME} is installed (${version})"

        # Check if it's in PATH
        if [[ "$(command -v "${BINARY_NAME}")" == "${INSTALL_DIR}/${BINARY_NAME}" ]]; then
            log_success "Binary is in PATH and ready to use"
        else
            log_warning "Binary installed but may not be in PATH"
            log_info "You may need to add ${INSTALL_DIR} to your PATH"
        fi

        return 0
    else
        log_error "Installation verification failed"
        log_info "Try adding ${INSTALL_DIR} to your PATH:"
        echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        return 1
    fi
}

# Show usage information
show_usage() {
    log_info "GHCP Memory Context Server is now installed!"
    echo ""
    echo "Quick Start:"
    echo "  ${BINARY_NAME}                              # Start server on port 8080"
    echo "  ${BINARY_NAME} --port 3000                  # Start on custom port"
    echo "  ${BINARY_NAME} --data-dir /path/to/data     # Use custom data directory"
    echo "  ${BINARY_NAME} --help                       # Show all options"
    echo ""
    echo "VS Code Integration:"
    echo "  1. Install GitHub Copilot extension"
    echo "  2. Add MCP configuration to .vscode/settings.json"
    echo "  3. Start the memory server"
    echo ""
    echo "Documentation: https://github.com/${GITHUB_REPO}#readme"
}

# Uninstall function
uninstall() {
    log_info "Uninstalling ${BINARY_NAME}..."

    if [[ -f "${INSTALL_DIR}/${BINARY_NAME}" ]]; then
        if [[ ! -w "$INSTALL_DIR" ]]; then
            sudo rm -f "${INSTALL_DIR}/${BINARY_NAME}"
        else
            rm -f "${INSTALL_DIR}/${BINARY_NAME}"
        fi
        log_success "${BINARY_NAME} uninstalled successfully"
    else
        log_warning "${BINARY_NAME} not found in ${INSTALL_DIR}"
    fi
}

# Main function
main() {
    local version=""
    local action="install"

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --version=*)
                version="${1#*=}"
                shift
                ;;
            --uninstall)
                action="uninstall"
                shift
                ;;
            --help|-h)
                echo "GHCP Memory Context Server Installer"
                echo ""
                echo "Usage: $0 [OPTIONS] [VERSION]"
                echo ""
                echo "OPTIONS:"
                echo "  --version=VERSION    Install specific version"
                echo "  --uninstall          Uninstall the binary"
                echo "  --help, -h           Show this help"
                echo ""
                echo "ENVIRONMENT VARIABLES:"
                echo "  INSTALL_DIR          Installation directory (default: /usr/local/bin)"
                echo ""
                echo "Examples:"
                echo "  $0                   # Install latest version"
                echo "  $0 v1.0.0            # Install specific version"
                echo "  $0 --version=v1.0.0  # Install specific version"
                echo "  $0 --uninstall       # Uninstall"
                exit 0
                ;;
            v*)
                version="$1"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    case "$action" in
        install)
            install_binary "$version"
            verify_installation
            show_usage
            ;;
        uninstall)
            uninstall
            ;;
    esac
}

# Check if running with bash
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
