#!/bin/bash

# Script to build Debian package for Libr client
set -e

# Default values
VERSION=${VERSION:-"1.0.0~beta"}
ARCH=${ARCH:-"amd64"}
BUILD_DIR="dist"

echo "Building Libr Debian package..."
echo "Version: $VERSION"
echo "Architecture: $ARCH"

# Check if the binary exists
BINARY_PATH="./dist/libr-linux-$ARCH"
if [ ! -f "$BINARY_PATH" ]; then
    echo "Error: Binary not found at $BINARY_PATH"
    echo "Please build the Wails client first with: wails build"
    exit 1
fi

# Create build directory if it doesn't exist
mkdir -p "$BUILD_DIR"

# Build the package
echo "Creating Debian package..."
nfpm pkg --packager deb \
    --config packaging/nfpm.yaml \
    --target "$BUILD_DIR/libr_${VERSION}_${ARCH}.deb"

echo "Package created successfully: $BUILD_DIR/libr_${VERSION}_${ARCH}.deb"

# Display package info
echo ""
echo "Package Information:"
dpkg --info "$BUILD_DIR/libr_${VERSION}_${ARCH}.deb"

echo ""
echo "Package Contents:"
dpkg --contents "$BUILD_DIR/libr_${VERSION}_${ARCH}.deb"

echo ""
echo "Package build completed successfully!"
echo "You can install it with: sudo dpkg -i $BUILD_DIR/libr_${VERSION}_${ARCH}.deb"
