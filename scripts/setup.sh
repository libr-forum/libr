#!/bin/bash

# 🚀 LIBR Quick Setup Script
# This script helps new contributors get started quickly!

set -e

echo "🚀 Welcome to LIBR!"
echo "=================="
echo ""
echo "This script will help you set up the project on your computer."
echo ""

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to print status
print_status() {
    if [ $? -eq 0 ]; then
        echo "✅ $1"
    else
        echo "❌ $1"
    fi
}

echo "🔍 Checking your system..."
echo ""

# Check Git
if command_exists git; then
    echo "✅ Git is installed"
    git_version=$(git --version)
    echo "   Version: $git_version"
else
    echo "❌ Git is not installed"
    echo "   Please install Git: https://git-scm.com/downloads"
    exit 1
fi

echo ""

# Check Node.js
if command_exists node; then
    echo "✅ Node.js is installed"
    node_version=$(node --version)
    echo "   Version: $node_version"
    
    # Check if version is 18 or higher
    major_version=$(echo $node_version | sed 's/v\([0-9]*\).*/\1/')
    if [ "$major_version" -ge 18 ]; then
        echo "   👍 Version is good (18+ required)"
    else
        echo "   ⚠️  Version might be too old (18+ recommended)"
        echo "   Consider updating: https://nodejs.org/"
    fi
else
    echo "❌ Node.js is not installed"
    echo "   Please install Node.js 18+: https://nodejs.org/"
    echo "   (Needed for web client development)"
fi

echo ""

# Check Go
if command_exists go; then
    echo "✅ Go is installed"
    go_version=$(go version)
    echo "   Version: $go_version"
else
    echo "❌ Go is not installed"
    echo "   Please install Go 1.21+: https://golang.org/dl/"
    echo "   (Needed for core protocol development)"
fi

echo ""

# Check Flutter
if command_exists flutter; then
    echo "✅ Flutter is installed"
    flutter_version=$(flutter --version | head -n 1)
    echo "   Version: $flutter_version"
else
    echo "❌ Flutter is not installed"
    echo "   Please install Flutter 3.16+: https://flutter.dev/docs/get-started/install"
    echo "   (Needed for mobile app development)"
fi

echo ""
echo "📁 Setting up the project..."
echo ""

# Check if we're in the right directory
if [ ! -f "README.md" ] || [ ! -d "src" ]; then
    echo "❌ This doesn't look like the LIBR project directory."
    echo "   Make sure you're running this script from the LIBR project root."
    exit 1
fi

echo "✅ Found LIBR project files"

# Set up each component that exists
echo ""
echo "🔧 Setting up project components..."
echo ""

# Core Protocol (Go)
if [ -d "src/core-protocol" ] && [ -f "src/core-protocol/go.mod" ]; then
    echo "📦 Setting up Core Protocol..."
    cd src/core-protocol
    go mod download
    print_status "Core Protocol dependencies installed"
    cd ../..
elif [ -d "src/core-protocol" ]; then
    echo "⚠️  Core Protocol directory exists but no go.mod found"
    echo "   This is normal if the Go project isn't set up yet"
else
    echo "ℹ️  No Core Protocol directory found (that's okay!)"
fi

# Network Layer (Go)
if [ -d "src/network" ] && [ -f "src/network/go.mod" ]; then
    echo "📦 Setting up Network Layer..."
    cd src/network
    go mod download
    print_status "Network Layer dependencies installed"
    cd ../..
elif [ -d "src/network" ]; then
    echo "⚠️  Network Layer directory exists but no go.mod found"
    echo "   This is normal if the Go project isn't set up yet"
else
    echo "ℹ️  No Network Layer directory found (that's okay!)"
fi

# Web client (Node.js)
if [ -d "src/web-client" ] && [ -f "src/web-client/package.json" ]; then
    echo "📦 Setting up web client..."
    cd src/web-client
    npm install
    print_status "Web client dependencies installed"
    cd ../..
elif [ -d "src/web-client" ]; then
    echo "⚠️  Web client directory exists but no package.json found"
    echo "   This is normal if the web project isn't set up yet"
else
    echo "ℹ️  No web client directory found (that's okay!)"
fi

# Mobile client (Flutter)
if [ -d "src/mobile-client" ] && [ -f "src/mobile-client/pubspec.yaml" ]; then
    echo "📦 Setting up mobile client..."
    cd src/mobile-client
    flutter pub get
    print_status "Mobile client dependencies installed"
    cd ../..
elif [ -d "src/mobile-client" ]; then
    echo "⚠️  Mobile client directory exists but no pubspec.yaml found"
    echo "   This is normal if the Flutter project isn't set up yet"
else
    echo "ℹ️  No mobile client directory found (that's okay!)"
fi

# Smart contracts (Node.js)
if [ -d "src/contracts" ] && [ -f "src/contracts/package.json" ]; then
    echo "📦 Setting up smart contracts..."
    cd src/contracts
    npm install
    print_status "Smart contract dependencies installed"
    cd ../..
elif [ -d "src/contracts" ]; then
    echo "⚠️  Contracts directory exists but no package.json found"
    echo "   This is normal if the contracts aren't set up yet"
else
    echo "ℹ️  No contracts directory found (that's okay!)"
fi

echo ""
echo "🎉 Setup complete!"
echo ""
echo "📚 What's next?"
echo "==============="
echo ""
echo "1. 📖 Read the Beginner Guide: docs/BEGINNER_GUIDE.md"
echo "2. 🐛 Find a 'good first issue': https://github.com/libr-forum/libr/labels/good%20first%20issue"
echo "3. 💬 Ask questions if you need help!"
echo ""
echo "🛠️  Useful commands:"
echo "   ./scripts/validate-commit.sh \"feat: your message\"  # Check commit messages"
echo "   git status                                          # See what you've changed"
echo "   git add .                                           # Stage your changes"
echo "   git commit -m \"feat: your message\"                 # Commit your changes"
echo ""
echo "❓ Need help?"
echo "   • Create an issue: https://github.com/libr-forum/libr/issues/new"
echo "   • Contact mentors: Check the README for contact info"
echo ""
echo "Happy coding! 🚀"
