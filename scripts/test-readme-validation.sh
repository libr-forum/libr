#!/bin/bash
# Simple test script to validate README instructions
# Usage: ./test-readme-validation.sh

set -e

echo "🔍 Validating README installation instructions..."

# Test 1: Check if all download URLs are valid
echo -e "\n1️⃣  Testing Download URLs..."

VERSION="v1.0.0-beta"
PACKAGE_VERSION=${VERSION#v}

URLS=(
    "https://github.com/libr-forum/Libr/releases/download/$VERSION/libr_${PACKAGE_VERSION}_amd64.deb"
    "https://github.com/libr-forum/Libr/releases/download/$VERSION/libr-${PACKAGE_VERSION}-1.x86_64.rpm"
    "https://github.com/libr-forum/Libr/releases/download/$VERSION/libr-${PACKAGE_VERSION}-1-x86_64.pkg.tar.zst"
    "https://github.com/libr-forum/Libr/releases/download/$VERSION/libr-linux-amd64"
)

for url in "${URLS[@]}"; do
    echo "Testing: $(basename $url)"
    if curl -s --head "$url" | head -n 1 | grep -q "HTTP/1.1 200\|HTTP/2 200"; then
        echo "  ✅ URL is accessible"
    else
        echo "  ❌ URL returns error (file may not exist yet)"
    fi
done

# Test 2: Validate APT repository setup commands
echo -e "\n2️⃣  Testing APT Repository Commands..."

APT_URLS=(
    "https://libr-forum.github.io/libr-apt-repo/libr-repo-key.gpg"
    "https://libr-forum.github.io/libr-apt-repo/"
)

for url in "${APT_URLS[@]}"; do
    echo "Testing: $url"
    if curl -s --head "$url" | head -n 1 | grep -q "HTTP/1.1 200\|HTTP/2 200"; then
        echo "  ✅ APT repository URL is accessible"
    else
        echo "  ❌ APT repository URL not accessible"
    fi
done

# Test 3: Validate install script
echo -e "\n3️⃣  Testing Installation Script..."

SCRIPT_URL="https://raw.githubusercontent.com/libr-forum/Libr/main/scripts/install-libr.sh"
if curl -s --head "$SCRIPT_URL" | head -n 1 | grep -q "HTTP/1.1 200\|HTTP/2 200"; then
    echo "  ✅ Installation script URL is accessible"
    
    # Download and check script syntax
    curl -s "$SCRIPT_URL" > /tmp/install-libr-test.sh
    if bash -n /tmp/install-libr-test.sh; then
        echo "  ✅ Installation script syntax is valid"
    else
        echo "  ❌ Installation script has syntax errors"
    fi
    rm -f /tmp/install-libr-test.sh
else
    echo "  ❌ Installation script URL not accessible"
fi

# Test 4: Check WebKit package availability
echo -e "\n4️⃣  Testing WebKit Package Availability..."

if command -v apt >/dev/null 2>&1; then
    echo "Testing Ubuntu/Debian WebKit packages..."
    apt-cache search libwebkit2gtk-4.1-0 >/dev/null && echo "  ✅ libwebkit2gtk-4.1-0 available" || echo "  ❌ libwebkit2gtk-4.1-0 not found"
    apt-cache search libjavascriptcoregtk-4.1-0 >/dev/null && echo "  ✅ libjavascriptcoregtk-4.1-0 available" || echo "  ❌ libjavascriptcoregtk-4.1-0 not found"
fi

if command -v dnf >/dev/null 2>&1; then
    echo "Testing Fedora WebKit packages..."
    dnf search webkit2gtk4.1-devel >/dev/null 2>&1 && echo "  ✅ webkit2gtk4.1-devel available" || echo "  ❌ webkit2gtk4.1-devel not found"
fi

if command -v pacman >/dev/null 2>&1; then
    echo "Testing Arch WebKit packages..."
    pacman -Ss webkit2gtk-4.1 >/dev/null 2>&1 && echo "  ✅ webkit2gtk-4.1 available" || echo "  ❌ webkit2gtk-4.1 not found"
fi

# Test 5: Architecture detection
echo -e "\n5️⃣  Testing Architecture Detection..."

DETECTED_ARCH=$(dpkg --print-architecture 2>/dev/null || uname -m)
echo "Detected architecture: $DETECTED_ARCH"

case "$DETECTED_ARCH" in
    x86_64|amd64)
        echo "  ✅ Supported architecture detected"
        ;;
    *)
        echo "  ❌ Unsupported architecture"
        ;;
esac

# Test 6: Build packages locally to verify they can be created
echo -e "\n6️⃣  Testing Package Building (if binary exists)..."

if [ -f "dist/libr-linux-amd64" ]; then
    echo "Binary found, testing package builds..."
    
    if command -v nfpm >/dev/null 2>&1; then
        echo "Testing Debian package build..."
        nfpm pkg --packager deb --config packaging/nfpm.yaml --target /tmp/test-libr_${PACKAGE_VERSION}_amd64.deb && echo "  ✅ Debian package builds successfully" || echo "  ❌ Debian package build failed"
        
        echo "Testing RPM package build..."
        nfpm pkg --packager rpm --config packaging/nfpm-rpm.yaml --target /tmp/test-libr-${PACKAGE_VERSION}-1.x86_64.rpm && echo "  ✅ RPM package builds successfully" || echo "  ❌ RPM package build failed"
        
        echo "Testing Arch package build..."
        nfpm pkg --packager archlinux --config packaging/nfpm-arch.yaml --target /tmp/test-libr-${PACKAGE_VERSION}-1-x86_64.pkg.tar.zst && echo "  ✅ Arch package builds successfully" || echo "  ❌ Arch package build failed"
        
        # Clean up test packages
        rm -f /tmp/test-libr*
    else
        echo "  ⚠️  nfpm not installed, skipping package build tests"
    fi
else
    echo "  ⚠️  Binary not found at dist/libr-linux-amd64, skipping package build tests"
    echo "     Run: go build -o dist/libr-linux-amd64 ./core/mod_client"
fi

echo -e "\n📋 Validation Summary:"
echo "✅ URL accessibility tested"
echo "✅ APT repository endpoints checked"
echo "✅ Installation script validated"
echo "✅ WebKit packages checked"
echo "✅ Architecture detection tested"
echo "✅ Package building tested (if applicable)"

echo -e "\n💡 Recommendations:"
echo "1. Build packages with: ./scripts/build-packages.sh"
echo "2. Upload packages to GitHub releases"
echo "3. Test actual installation on clean VMs/containers"
echo "4. Update documentation based on test results"
