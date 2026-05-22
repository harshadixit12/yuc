#!/usr/bin/env bash

set -euo pipefail

REPO="harshadixit12/yuc"
VERSION="${VERSION:-v0.1.1}"

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Darwin)
    PLATFORM="darwin"
    ;;
  Linux)
    PLATFORM="linux"
    ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

case "$ARCH" in
  arm64|aarch64)
    ARCH="arm64"
    ;;
  x86_64|amd64)
    ARCH="amd64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

FILENAME="yuc_${PLATFORM}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

TMP_DIR="$(mktemp -d)"

echo "Downloading ${URL}..."
curl -fsSL "$URL" -o "${TMP_DIR}/${FILENAME}"

echo "Extracting..."
tar -xzf "${TMP_DIR}/${FILENAME}" -C "${TMP_DIR}"

chmod +x "${TMP_DIR}/yuc"

INSTALL_DIR="/usr/local/bin"

if [ ! -w "$INSTALL_DIR" ]; then
  echo "Installing to ${INSTALL_DIR} using sudo..."
  sudo mv "${TMP_DIR}/yuc" "${INSTALL_DIR}/yuc"
else
  mv "${TMP_DIR}/yuc" "${INSTALL_DIR}/yuc"
fi

rm -rf "${TMP_DIR}"

echo
echo "yuc installed successfully!"
echo
echo "Run:"
echo "  yuc --help"