#!/bin/bash
set -euo pipefail

# ============================================================
#   LIBR Installer Script
#   Supports: Ubuntu/Debian, Fedora/RHEL/CentOS, Arch
# ============================================================

# --- Debug logging helper ---
log() { echo -e "👉 $1"; }
success() { echo -e "✅ $1"; }
error() { echo -e "❌ $1" >&2; }

# --- Detect distribution & architecture ---
if [ -f /etc/os-release ]; then
  source /etc/os-release
  DISTRO=$ID
else
  error "Cannot detect distribution!"
  exit 1
fi
ARCH=$(dpkg --print-architecture 2>/dev/null || uname -m)

# --- Fetch latest or use provided version ---
LATEST_VERSION=$(curl -s https://api.github.com/repos/libr-forum/libr/releases/latest \
  | grep tag_name | cut -d '"' -f4)
VERSION=${1:-$LATEST_VERSION}

# --- Check installed version ---
if command -v libr >/dev/null 2>&1; then
  INSTALLED_VERSION=$(libr --version 2>/dev/null | awk '{print $2}')
else
  INSTALLED_VERSION="none"
fi

if [ "$INSTALLED_VERSION" = "$VERSION" ]; then
  success "libr $VERSION is already installed."
  exit 0
fi

log "📦 Installing libr $VERSION for $DISTRO ($ARCH)..."

# --- GTK check & install (only if missing) ---
install_gtk_if_missing() {
  case "$DISTRO" in
    ubuntu|debian)
      if ! dpkg -l | grep -q libgtk-3-0; then
        log "Installing GTK dependencies..."
        sudo apt-get update -y
        sudo apt-get install -y libgtk-3-0 || log "GTK install skipped"
      else
        log "GTK already present, skipping."
      fi
      ;;
    fedora|rhel|centos)
      if ! rpm -q gtk3 >/dev/null 2>&1; then
        log "Installing GTK dependencies..."
        sudo dnf install -y gtk3 || sudo yum install -y gtk3 || log "GTK install skipped"
      else
        log "GTK already present, skipping."
      fi
      ;;
    arch)
      if ! pacman -Q gtk3 >/dev/null 2>&1; then
        log "Installing GTK dependencies..."
        sudo pacman -Sy --noconfirm gtk3 || log "GTK install skipped"
      else
        log "GTK already present, skipping."
      fi
      ;;
  esac
}

# --- Main installation per distro ---
case "$DISTRO" in
  ubuntu|debian)
    URL="https://github.com/libr-forum/libr/releases/download/$VERSION/libr_${VERSION}_${ARCH}.deb"
    log "⬇️ Downloading $URL"
    wget -qO libr.deb "$URL" || { error "Failed to download $URL"; exit 1; }
    sudo dpkg -i libr.deb || sudo apt-get install -f -y
    rm -f libr.deb
    install_gtk_if_missing
    ;;
  fedora|rhel|centos)
    URL="https://github.com/libr-forum/libr/releases/download/$VERSION/libr-${VERSION}.${ARCH}.rpm"
    log "⬇️ Downloading $URL"
    wget -qO libr.rpm "$URL" || { error "Failed to download $URL"; exit 1; }
    sudo dnf install -y ./libr.rpm || sudo yum install -y ./libr.rpm
    rm -f libr.rpm
    install_gtk_if_missing
    ;;
  arch)
    URL="https://github.com/libr-forum/libr/releases/download/$VERSION/libr-${VERSION}-${ARCH}.pkg.tar.zst"
    log "⬇️ Downloading $URL"
    wget -qO libr.pkg.tar.zst "$URL" || { error "Failed to download $URL"; exit 1; }
    sudo pacman -U --noconfirm libr.pkg.tar.zst
    rm -f libr.pkg.tar.zst
    install_gtk_if_missing
    ;;
  *)
    error "Unsupported distribution: $DISTRO"
    exit 1
    ;;
esac

success "libr $VERSION installed successfully."
