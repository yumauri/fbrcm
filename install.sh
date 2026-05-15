#!/bin/sh

# one-liner installer for fbrcm
# Users run: curl -sSfL https://raw.githubusercontent.com/yumauri/fbrcm/main/install.sh | sh

set -eu

REPO="yumauri/fbrcm"
APP="fbrcm"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
TMP=""

cleanup() {
  if [ -n "$TMP" ] && [ -d "$TMP" ]; then
    rm -rf "$TMP"
  fi
}
trap cleanup EXIT

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1"
    exit 1
  fi
}

need curl
need grep
need install
need mktemp
need sed
need tar
need tr
need uname

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64)  ARCH="x86_64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

case "$OS" in
  linux)  EXT="tar.gz" ;;
  darwin) EXT="tar.gz" ;;
  *) echo "Unsupported OS: $OS - please download manually from https://github.com/$REPO/releases" && exit 1 ;;
esac

# Fetch latest release tag
LATEST=$(curl -sSf "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Could not determine the latest release. Check https://github.com/$REPO/releases"
  exit 1
fi

URL="https://github.com/$REPO/releases/download/${LATEST}/${APP}_${OS}_${ARCH}.${EXT}"

echo "Downloading $APP $LATEST for $OS/$ARCH..."
TMP=$(mktemp -d)
curl -sSfL "$URL" | tar -xz -C "$TMP"

echo "Installing to $INSTALL_DIR/$APP ..."
if [ -w "$INSTALL_DIR" ]; then
  install -m 755 "$TMP/$APP" "$INSTALL_DIR/$APP"
elif command -v sudo >/dev/null 2>&1; then
  sudo install -m 755 "$TMP/$APP" "$INSTALL_DIR/$APP"
else
  echo "Install dir is not writable: $INSTALL_DIR"
  echo "Set INSTALL_DIR to a writable directory or run as a user with permission."
  exit 1
fi

echo "Done! Run: $APP --help"
