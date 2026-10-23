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
need awk
need grep
need install
need mktemp
need sed
need tar
need tr
need uname

if command -v sha256sum >/dev/null 2>&1; then
  SHA256_TOOL=sha256sum
elif command -v shasum >/dev/null 2>&1; then
  SHA256_TOOL=shasum
else
  echo "Missing required command: sha256sum or shasum"
  exit 1
fi

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

ASSET="${APP}_${OS}_${ARCH}.${EXT}"
BASE_URL="https://github.com/$REPO/releases/download/${LATEST}"

echo "Downloading $APP $LATEST for $OS/$ARCH..."
TMP=$(mktemp -d)
ARCHIVE="$TMP/$ASSET"
CHECKSUMS="$TMP/checksums.txt"
curl -sSfL "$BASE_URL/$ASSET" -o "$ARCHIVE"
curl -sSfL "$BASE_URL/checksums.txt" -o "$CHECKSUMS"

if ! EXPECTED_CHECKSUM=$(awk -v asset="$ASSET" '
  NF == 2 && ($2 == asset || $2 == "*" asset) &&
    length($1) == 64 && $1 !~ /[^[:xdigit:]]/ {
      checksum = tolower($1)
      matches++
    }
  END {
    if (matches != 1) exit 1
    print checksum
  }
' "$CHECKSUMS"); then
  echo "Could not find a unique valid checksum for $ASSET."
  exit 1
fi

if [ "$SHA256_TOOL" = sha256sum ]; then
  ACTUAL_CHECKSUM=$(sha256sum "$ARCHIVE")
else
  ACTUAL_CHECKSUM=$(shasum -a 256 "$ARCHIVE")
fi
ACTUAL_CHECKSUM=${ACTUAL_CHECKSUM%% *}

if [ "$ACTUAL_CHECKSUM" != "$EXPECTED_CHECKSUM" ]; then
  echo "Checksum verification failed for $ASSET."
  exit 1
fi

tar -xzf "$ARCHIVE" -C "$TMP"

if [ ! -f "$TMP/LICENSE" ]; then
  echo "The release archive does not contain LICENSE."
  exit 1
fi

echo "Installing to $INSTALL_DIR/$APP ..."
if [ -w "$INSTALL_DIR" ]; then
  install -m 755 "$TMP/$APP" "$INSTALL_DIR/$APP"
  install -m 644 "$TMP/LICENSE" "$INSTALL_DIR/$APP-LICENSE.txt"
  if [ -f "$TMP/THIRD_PARTY_NOTICES" ]; then
    install -m 644 "$TMP/THIRD_PARTY_NOTICES" "$INSTALL_DIR/$APP-THIRD_PARTY_NOTICES.txt"
  fi
elif command -v sudo >/dev/null 2>&1; then
  sudo install -m 755 "$TMP/$APP" "$INSTALL_DIR/$APP"
  sudo install -m 644 "$TMP/LICENSE" "$INSTALL_DIR/$APP-LICENSE.txt"
  if [ -f "$TMP/THIRD_PARTY_NOTICES" ]; then
    sudo install -m 644 "$TMP/THIRD_PARTY_NOTICES" "$INSTALL_DIR/$APP-THIRD_PARTY_NOTICES.txt"
  fi
else
  echo "Install dir is not writable: $INSTALL_DIR"
  echo "Set INSTALL_DIR to a writable directory or run as a user with permission."
  exit 1
fi

echo "Done! Run: $APP --help"
