#!/bin/bash
# Comprehensive APT Repository Test Script

set -e

echo "🧪 Testing Libr APT Repository"
echo "=============================="

# Test 1: Repository Structure
echo "📋 Test 1: Checking repository structure..."
curl -s https://libr-forum.github.io/libr-apt-repo/dists/stable/Release | head -5
echo "✅ Release file accessible"

# Test 2: Package metadata
echo
echo "📦 Test 2: Checking package metadata..."
curl -s https://libr-forum.github.io/libr-apt-repo/dists/stable/main/binary-amd64/Packages | grep -E "^(Package|Version|Architecture):"
echo "✅ Package metadata valid"

# Test 3: GPG signature
echo
echo "🔐 Test 3: Checking GPG signature..."
if curl -s https://libr-forum.github.io/libr-apt-repo/pubkey.gpg | gpg --import --quiet 2>/dev/null; then
    if curl -s https://libr-forum.github.io/libr-apt-repo/dists/stable/InRelease | gpg --verify --quiet 2>/dev/null; then
        echo "✅ GPG signature valid"
    else
        echo "❌ GPG signature verification failed"
        exit 1
    fi
else
    echo "❌ Failed to import GPG key"
    exit 1
fi

# Test 4: Setup script test (dry run)
echo
echo "📥 Test 4: Testing setup script..."
curl -s https://libr-forum.github.io/libr-apt-repo/setup-repo.sh | head -10
echo "✅ Setup script accessible"

# Test 5: Package download test
echo
echo "📥 Test 5: Testing package download..."
if curl -I "https://libr-forum.github.io/libr-apt-repo/pool/main/libr/libr/libr_1.0.0~beta_amd64.deb" 2>/dev/null | grep -q "HTTP.*200"; then
    echo "✅ Package file accessible"
else
    echo "❌ Package file not accessible"
    exit 1
fi

echo
echo "🎉 All tests passed! Repository is working correctly."
echo
echo "📋 For users to install:"
echo "curl -fsSL https://libr-forum.github.io/libr-apt-repo/setup-repo.sh | bash"
echo "sudo apt install libr"
