#!/bin/bash
set -e

# Detect distribution
source /etc/os-release
DISTRO=$ID
ARCH=$(dpkg --print-architecture 2>/dev/null || uname -m)

# Fetch latest or use provided version
LATEST_VERSION=$(curl -s https://api.github.com/repos/libr-forum/libr/releases/latest \
  | grep tag_name | cut -d '"' -f4)
VERSION=${1:-$LATEST_VERSION}

# Check installed version
if command -v libr >/dev/null 2>&1; then
  INSTALLED_VERSION=$(libr --version | awk '{print $2}')
else
  INSTALLED_VERSION="none"
fi

if [ "$INSTALLED_VERSION" = "$VERSION" ]; then
  echo "✅ libr $VERSION already installed."
  exit 0
fi

echo "📦 Installing libr $VERSION for $DISTRO ($ARCH)..."

case "$DISTRO" in
  ubuntu|debian)
    URL="https://github.com/libr-forum/libr/releases/download/$VERSION/libr_${VERSION}_${ARCH}.deb"
    echo "⬇️ Downloading $URL..."
    wget -O libr.deb "$URL" || { echo "❌ Failed to download $URL"; exit 1; }
    echo "⚙️ Installing .deb package..."
    sudo dpkg -i libr.deb || { 
      echo "⚠️ dpkg failed, trying apt-get -f install..."; 
      sudo apt-get install -f -y; 
    }
    echo "🧹 Cleaning up..."
    rm libr.deb
    ;;
  fedora|rhel|centos)
    URL="https://github.com/libr-forum/libr/releases/download/$VERSION/libr-${VERSION}.${ARCH}.rpm"
    echo "⬇️ Downloading $URL..."
    wget -O libr.rpm "$URL" || { echo "❌ Failed to download $URL"; exit 1; }
    echo "⚙️ Installing .rpm package..."
    if command -v dnf >/dev/null 2>&1; then
      sudo dnf install -y ./libr.rpm
    else
      sudo yum install -y ./libr.rpm
    fi
    echo "🧹 Cleaning up..."
    rm libr.rpm
    ;;
  arch)
    URL="https://github.com/libr-forum/libr/releases/download/$VERSION/libr_${VERSION}-${ARCH}.pkg.tar.zst"
    echo "⬇️ Downloading $URL..."
    wget -O libr.pkg.tar.zst "$URL" || { echo "❌ Failed to download $URL"; exit 1; }
    echo "⚙️ Installing Arch package..."
    sudo pacman -U --noconfirm libr.pkg.tar.zst
    echo "🧹 Cleaning up..."
    rm libr.pkg.tar.zst
    ;;
  *)
    echo "❌ Unsupported distribution: $DISTRO"
    exit 1
    ;;
esac

echo "✅ libr $VERSION installed successfully."
