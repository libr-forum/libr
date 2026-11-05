#!/bin/bash
set -e

# Detect distribution
source /etc/os-release
DISTRO=$ID
DETECTED_ARCH=$(dpkg --print-architecture 2>/dev/null || uname -m)

# Convert architecture names for different package formats
case "$DETECTED_ARCH" in
  x86_64|amd64)
    DEB_ARCH="amd64"
    RPM_ARCH="x86_64"
    ARCH_ARCH="x86_64"
    ;;
  *)
    echo "❌ Unsupported architecture: $DETECTED_ARCH"
    exit 1
    ;;
esac

# Fetch latest or use provided version
LATEST_VERSION=$(curl -s https://api.github.com/repos/libr-forum/libr/releases/latest \
  | grep tag_name | cut -d '"' -f4)
VERSION=${1:-$LATEST_VERSION}

# Convert version for package names (remove 'v' prefix if present)
PACKAGE_VERSION=${VERSION#v}

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

echo "📦 Installing libr $VERSION for $DISTRO ($DETECTED_ARCH)..."

case "$DISTRO" in
  ubuntu|debian)
    URL="https://github.com/libr-forum/libr/releases/download/$VERSION/libr_${PACKAGE_VERSION}_${DEB_ARCH}.deb"
    wget -qO libr.deb "$URL"
    sudo apt install -y ./libr.deb
    rm libr.deb
    
    # Fix WebKit library issues for newer Ubuntu/Debian
    if ! ldconfig -p | grep -q libwebkit2gtk-4.0.so.37; then
      echo "🔧 Setting up WebKit compatibility..."
      if [ -f /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.1.so.0 ]; then
        sudo ln -sf /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.1.so.0 \
                    /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.0.so.37
        sudo ln -sf /usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.1.so.0 \
                    /usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.0.so.18
      fi
    fi
    ;;
  fedora|rhel|centos)
    URL="https://github.com/libr-forum/libr/releases/download/$VERSION/libr-${PACKAGE_VERSION}-1.${RPM_ARCH}.rpm"
    wget -qO libr.rpm "$URL"
    sudo dnf install -y ./libr.rpm || sudo yum install -y ./libr.rpm
    rm libr.rpm
    
    # Fix WebKit library issues for newer Fedora/RHEL
    if ! ldconfig -p | grep -q libwebkit2gtk-4.0.so.37; then
      echo "🔧 Setting up WebKit compatibility..."
      if [ -f /usr/lib64/libwebkit2gtk-4.1.so.0 ]; then
        sudo ln -sf /usr/lib64/libwebkit2gtk-4.1.so.0 \
                    /usr/lib64/libwebkit2gtk-4.0.so.37
        sudo ln -sf /usr/lib64/libjavascriptcoregtk-4.1.so.0 \
                    /usr/lib64/libjavascriptcoregtk-4.0.so.18
      fi
    fi
    ;;
  arch)
    URL="https://github.com/libr-forum/libr/releases/download/$VERSION/libr-${PACKAGE_VERSION}-1-${ARCH_ARCH}.pkg.tar.zst"
    wget -qO libr.pkg.tar.zst "$URL"
    sudo pacman -U --noconfirm libr.pkg.tar.zst
    rm libr.pkg.tar.zst
    
    # Fix WebKit library issues for Arch Linux
    if ! ldconfig -p | grep -q libwebkit2gtk-4.0.so.37; then
      echo "🔧 Setting up WebKit compatibility..."
      if [ -f /usr/lib/libwebkit2gtk-4.1.so.0 ]; then
        sudo ln -sf /usr/lib/libwebkit2gtk-4.1.so.0 \
                    /usr/lib/libwebkit2gtk-4.0.so.37
        sudo ln -sf /usr/lib/libjavascriptcoregtk-4.1.so.0 \
                    /usr/lib/libjavascriptcoregtk-4.0.so.18
      fi
    fi
    ;;
  *)
    echo "❌ Unsupported distribution: $DISTRO"
    exit 1
    ;;
esac

echo "✅ libr $VERSION installed successfully."
echo "🚀 You can now run 'libr' from anywhere or find it in your applications menu."
