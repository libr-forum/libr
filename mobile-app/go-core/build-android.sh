#!/bin/bash

# Configuration
PACKAGE_PATH="./bridge"
OUTPUT_NAME="libr-core.aar"
ANDROID_API=21

# Setup
export GO111MODULE=on

# Build
echo "🚀 Building Go mobile bridge for Android..."
gomobile bind -v -target=android -androidapi $ANDROID_API -o ../android/app/libs/$OUTPUT_NAME $PACKAGE_PATH

if [ $? -eq 0 ]; then
    echo "✅ Build successful! Output: ../android/app/libs/$OUTPUT_NAME"
else
    echo "❌ Build failed!"
    exit 1
fi
