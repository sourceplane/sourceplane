#!/bin/sh
set -e

# Sourceplane CLI installer script
# Usage: curl -sSfL https://raw.githubusercontent.com/sourceplane/cli/main/install.sh | sh
# Or with BINARY env var: BINARY=thinci curl -sSfL ... | sh

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
REPO="sourceplane/cli"
BINARY_NAME="${BINARY:-sp}"

get_latest_release() {
  curl --silent "https://api.github.com/repos/$REPO/releases/latest" |
    grep '"tag_name":' |
    sed -E 's/.*"([^"]+)".*/\1/'
}

detect_os_arch() {
  OS=$(uname -s | tr '[:upper:]' '[:lower:]')
  ARCH=$(uname -m)
  
  case "$OS" in
    linux) OS="Linux" ;;
    darwin) OS="Darwin" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
  esac
  
  case "$ARCH" in
    x86_64|amd64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    armv7l) ARCH="armv7" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
  esac
}

main() {
  detect_os_arch
  VERSION=$(get_latest_release)
  
  echo "Installing $BINARY_NAME $VERSION for $OS/$ARCH..."
  
  DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/${BINARY_NAME}_${VERSION#v}_${OS}_${ARCH}.tar.gz"
  TEMP_DIR=$(mktemp -d)
  
  echo "Downloading from $DOWNLOAD_URL..."
  curl -sSfL "$DOWNLOAD_URL" | tar -xz -C "$TEMP_DIR"
  
  echo "Installing to $INSTALL_DIR..."
  sudo mv "$TEMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
  sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
  
  rm -rf "$TEMP_DIR"
  
  echo "✓ $BINARY_NAME installed successfully!"
  echo "Run '$BINARY_NAME version' to verify the installation."
}

main
