#!/bin/bash

# Configuration
PACKAGE_PATH="./bridge"
OUTPUT_NAME="libr-core.aar"
ANDROID_API=21

# Setup
export GO111MODULE=on

# Build
echo "🚀 Building Go mobile bridge for Android..."
GOWORK=off go install golang.org/x/mobile/cmd/gobind # pinned by go.mod
export PATH="$(go env GOPATH)/bin:$PATH"
GOWORK=off go tool gomobile bind -v -target=android -androidapi $ANDROID_API -ldflags "-checklinkname=0" -o ../android/app/libs/$OUTPUT_NAME $PACKAGE_PATH

if [ $? -eq 0 ]; then
    echo "✅ Build successful! Output: ../android/app/libs/$OUTPUT_NAME"
else
    echo "❌ Build failed!"
    exit 1
fi
