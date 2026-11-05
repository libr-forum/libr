#!/bin/bash

# Script to create a simple APT repository for Libr packages
set -e

REPO_DIR=${REPO_DIR:-"./apt-repo"}
DIST=${DIST:-"stable"}
COMPONENT=${COMPONENT:-"main"}
ARCH=${ARCH:-"amd64"}

echo "Setting up APT repository..."
echo "Repository directory: $REPO_DIR"
echo "Distribution: $DIST"
echo "Component: $COMPONENT"
echo "Architecture: $ARCH"

# Create repository structure
mkdir -p "$REPO_DIR/dists/$DIST/$COMPONENT/binary-$ARCH"
mkdir -p "$REPO_DIR/pool/$COMPONENT"

# Copy packages to pool
echo "Copying packages to repository pool..."
cp dist/*.deb "$REPO_DIR/pool/$COMPONENT/" 2>/dev/null || echo "No .deb files found in dist/"

# Generate Packages file
echo "Generating Packages file..."
cd "$REPO_DIR"
dpkg-scanpackages pool/$COMPONENT /dev/null > "dists/$DIST/$COMPONENT/binary-$ARCH/Packages"

# Compress Packages file
gzip -k "dists/$DIST/$COMPONENT/binary-$ARCH/Packages"

# Generate Release file
echo "Generating Release file..."
cat > "dists/$DIST/Release" << EOF
Origin: Libr Repository
Label: Libr
Suite: $DIST
Codename: $DIST
Date: $(date -Ru)
Architectures: $ARCH
Components: $COMPONENT
Description: Libr - A Moderated, Censorship-Resilient Social Network Framework
EOF

# Calculate checksums for Release file
echo "MD5Sum:" >> "dists/$DIST/Release"
find "dists/$DIST" -name "Packages*" -exec md5sum {} \; | sed 's|dists/'$DIST'/| |' >> "dists/$DIST/Release"

echo "SHA1:" >> "dists/$DIST/Release"
find "dists/$DIST" -name "Packages*" -exec sha1sum {} \; | sed 's|dists/'$DIST'/| |' >> "dists/$DIST/Release"

echo "SHA256:" >> "dists/$DIST/Release"
find "dists/$DIST" -name "Packages*" -exec sha256sum {} \; | sed 's|dists/'$DIST'/| |' >> "dists/$DIST/Release"

cd - > /dev/null

echo ""
echo "APT repository created successfully in $REPO_DIR"
echo ""
echo "To use this repository, add the following to your sources.list:"
echo "deb [trusted=yes] file://$(realpath $REPO_DIR) $DIST $COMPONENT"
echo ""
echo "Or for HTTP hosting:"
echo "deb [trusted=yes] http://your-domain.com/path-to-repo $DIST $COMPONENT"
echo ""
echo "Repository structure:"
find "$REPO_DIR" -type f | head -20
