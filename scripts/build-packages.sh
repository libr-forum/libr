#!/bin/bash
set -e

# Build all package types for libr
# Usage: ./build-packages.sh [version]

VERSION=${1:-"1.0.0-beta"}
export VERSION

echo "📦 Building libr packages v$VERSION"

# Ensure dist directory exists
mkdir -p dist

# Ensure the binary is built
if [ ! -f "dist/libr-linux-amd64" ]; then
    echo "❌ Binary not found at dist/libr-linux-amd64"
    echo "Please build the binary first:"
    echo "  go build -o dist/libr-linux-amd64 ./core/mod_client"
    exit 1
fi

echo "🔨 Building Debian package..."
nfpm pkg --packager deb --config packaging/nfpm.yaml --target dist/libr_${VERSION}_amd64.deb

echo "🔨 Building RPM package..."
nfpm pkg --packager rpm --config packaging/nfpm-rpm.yaml --target dist/libr-${VERSION}-1.x86_64.rpm

echo "🔨 Building Arch Linux package..."
nfpm pkg --packager archlinux --config packaging/nfpm-arch.yaml --target dist/libr-${VERSION}-1-x86_64.pkg.tar.zst

echo "✅ All packages built successfully!"
echo ""
echo "📋 Generated packages:"
ls -la dist/*.{deb,rpm,zst} 2>/dev/null || echo "No packages found"

echo ""
echo "🚀 To test the packages:"
echo "  • Debian/Ubuntu: sudo dpkg -i dist/libr_${VERSION}_amd64.deb"
echo "  • Fedora/RHEL:   sudo dnf install dist/libr-${VERSION}-1.x86_64.rpm"
echo "  • Arch Linux:    sudo pacman -U dist/libr-${VERSION}-1-x86_64.pkg.tar.zst"
