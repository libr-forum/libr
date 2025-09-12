#!/usr/bin/env bash

set -euo pipefail

# -------------------------------
# Configuration
# -------------------------------
VERSION="${1:-latest}"   # default = latest release
ARCH="$(uname -m)"
DEBUG="${DEBUG:-1}"       # set DEBUG=0 to silence
REPO="libr-forum/libr"

log() { echo -e "🔹 $*"; }
debug() { [[ "$DEBUG" -eq 1 ]] && echo -e "🐞 DEBUG: $*"; }
err() { echo -e "❌ $*" >&2; exit 1; }

# -------------------------------
# Detect distro
# -------------------------------
detect_distro() {
  if [ -f /etc/os-release ]; then
    . /etc/os-release
    DISTRO=$ID
    debug "Detected distro: $DISTRO"
  else
    err "Unable to detect distribution."
  fi
}

# -------------------------------
# Detect installed version
# -------------------------------
installed_version() {
  if command -v libr >/dev/null 2>&1; then
    libr --version 2>/dev/null | awk '{print $2}'
  else
    echo ""
  fi
}

# -------------------------------
# Get latest release from GitHub
# -------------------------------
latest_version() {
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name":' | cut -d'"' -f4
}

# -------------------------------
# Download + Install package
# -------------------------------
install_pkg() {
  local url="$1"
  local file="$2"

  log "⬇️ Downloading $url"
  if ! curl -fL "$url" -o "$file"; then
    err "Failed to download $url"
  fi

  case "$DISTRO" in
    ubuntu|debian)
      sudo dpkg -i "$file" || sudo apt-get install -f -y
      ;;
    fedora|rhel|centos|rocky|almalinux|opensuse*)
      sudo rpm -Uvh --force "$file"
      ;;
    arch|manjaro)
      sudo pacman -U --noconfirm "$file"
      ;;
    *)
      err "Unsupported distro: $DISTRO"
      ;;
  esac
}

# -------------------------------
# Main
# -------------------------------
main() {
  detect_distro

  if [[ "$VERSION" == "latest" ]]; then
    VERSION="$(latest_version)"
  fi
  debug "Target version: $VERSION"

  CURRENT="$(installed_version)"
  debug "Currently installed version: ${CURRENT:-none}"

  if [[ "$CURRENT" == "$VERSION" ]]; then
    log "✅ libr $VERSION is already installed."
    exit 0
  fi

  case "$DISTRO" in
    ubuntu|debian)
      FILE="libr_${VERSION}_${ARCH}.deb"
      ;;
    fedora|rhel|centos|rocky|almalinux|opensuse*)
      FILE="libr-${VERSION}.${ARCH}.rpm"
      ;;
    arch|manjaro)
      FILE="libr-${VERSION}-${ARCH}.pkg.tar.zst"
      ;;
    *)
      err "Unsupported distro: $DISTRO"
      ;;
  esac

  URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILE}"
  debug "Download URL: $URL"
  debug "Local filename: $FILE"

  install_pkg "$URL" "$FILE"

  log "🎉 libr $VERSION installed successfully!"
}

main "$@"
