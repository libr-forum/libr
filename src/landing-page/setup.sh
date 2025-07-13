#!/bin/bash

# Libr Landing Page Setup Script
# This script installs dependencies and sets up the development environment

echo "🚀 Setting up Libr Landing Page..."

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 18+ first."
    echo "Visit: https://nodejs.org/"
    exit 1
fi

# Check Node.js version
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -lt 18 ]; then
    echo "❌ Node.js version 18+ is required. Current version: $(node -v)"
    exit 1
fi

echo "✅ Node.js $(node -v) detected"

# Navigate to landing page directory
cd "$(dirname "$0")" || exit 1

# Clean any existing node_modules and lock files to start fresh
echo "🧹 Cleaning existing dependencies..."
rm -rf node_modules package-lock.json yarn.lock

# Install dependencies
echo "📦 Installing dependencies..."
if command -v npm &> /dev/null; then
    npm install
elif command -v yarn &> /dev/null; then
    yarn install
else
    echo "❌ Neither npm nor yarn found. Please install one of them."
    exit 1
fi

echo "✅ Dependencies installed successfully!"

# Check if installation was successful
if [ -d "node_modules" ]; then
    echo ""
    echo "🎉 Setup complete! You can now run:"
    echo ""
    echo "  npm run dev     # Start development server"
    echo "  npm run build   # Build for production"
    echo "  npm run preview # Preview production build"
    echo ""
    echo "The development server will be available at http://localhost:5173"
    echo ""
    echo "📋 Features included:"
    echo "  • Modern React 18 with TypeScript"
    echo "  • Tailwind CSS for styling"
    echo "  • Framer Motion for animations"
    echo "  • Responsive design for all devices"
    echo "  • Dark/Light mode toggle"
    echo "  • Professional landing page sections"
else
    echo "❌ Installation failed. Please check the error messages above."
    exit 1
fi
