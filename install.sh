#!/bin/bash
# Cortex Installer Script
# Usage: curl -fsSL https://raw.githubusercontent.com/adityaraj/cortex/main/install.sh | bash

set -e

REPO="adityaraj/cortex"
BINARY_NAME="cortex"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_banner() {
    echo -e "${CYAN}"
    echo "   ██████╗ ██████╗ ██████╗ ████████╗███████╗██╗  ██╗"
    echo "  ██╔════╝██╔═══██╗██╔══██╗╚══██╔══╝██╔════╝╚██╗██╔╝"
    echo "  ██║     ██║   ██║██████╔╝   ██║   █████╗   ╚███╔╝ "
    echo "  ██║     ██║   ██║██╔══██╗   ██║   ██╔══╝   ██╔██╗ "
    echo "  ╚██████╗╚██████╔╝██║  ██║   ██║   ███████╗██╔╝ ██╗"
    echo "   ╚═════╝ ╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚══════╝╚═╝  ╚═╝"
    echo -e "${NC}"
    echo "  AI Agent Orchestrator Installer"
    echo ""
}

info() {
    echo -e "${CYAN}ℹ ${NC}$1"
}

success() {
    echo -e "${GREEN}✓ ${NC}$1"
}

error() {
    echo -e "${RED}✗ ${NC}$1"
    exit 1
}

warn() {
    echo -e "${YELLOW}⚠ ${NC}$1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            error "Unsupported architecture: $ARCH"
            ;;
    esac

    case "$OS" in
        darwin)
            OS="darwin"
            ;;
        linux)
            OS="linux"
            ;;
        mingw*|msys*|cygwin*)
            OS="windows"
            ;;
        *)
            error "Unsupported OS: $OS"
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    info "Detected platform: $PLATFORM"
}

# Get latest release version
get_latest_version() {
    info "Fetching latest version..."
    VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$VERSION" ]; then
        error "Failed to fetch latest version"
    fi

    info "Latest version: $VERSION"
}

# Download and install
install_cortex() {
    local TMP_DIR=$(mktemp -d)
    local ARCHIVE_NAME="${BINARY_NAME}-${OS}-${ARCH}"
    local DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}.tar.gz"

    if [ "$OS" = "windows" ]; then
        DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}.zip"
    fi

    info "Downloading ${BINARY_NAME}..."

    if command -v curl &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "${TMP_DIR}/archive"
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "${TMP_DIR}/archive"
    else
        error "Neither curl nor wget found. Please install one of them."
    fi

    info "Extracting..."
    cd "$TMP_DIR"

    if [ "$OS" = "windows" ]; then
        unzip -q archive
    else
        tar -xzf archive
    fi

    info "Installing to ${INSTALL_DIR}..."

    if [ -w "$INSTALL_DIR" ]; then
        mv "${BINARY_NAME}" "${INSTALL_DIR}/"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        warn "Requires sudo to install to ${INSTALL_DIR}"
        sudo mv "${BINARY_NAME}" "${INSTALL_DIR}/"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    # Cleanup
    rm -rf "$TMP_DIR"
}

# Verify installation
verify_installation() {
    if command -v cortex &> /dev/null; then
        INSTALLED_VERSION=$(cortex --version 2>&1 | head -n1)
        success "Cortex installed successfully!"
        echo ""
        echo "  Version: $INSTALLED_VERSION"
        echo "  Location: $(which cortex)"
        echo ""
        echo -e "  ${CYAN}Quick Start:${NC}"
        echo "    cortex validate    # Validate your Cortexfile"
        echo "    cortex run         # Run your workflow"
        echo "    cortex sessions    # View past sessions"
        echo ""
    else
        error "Installation failed. Please try again or install manually."
    fi
}

# Main
main() {
    print_banner
    detect_platform
    get_latest_version
    install_cortex
    verify_installation
}

main "$@"
