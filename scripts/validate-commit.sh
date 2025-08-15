#!/bin/bash

# 🔍 Commit Message Validator for LIBR
# This script helps you write good commit messages!
# 
# Usage: ./scripts/validate-commit.sh "your commit message"
# Example: ./scripts/validate-commit.sh "feat: add login button"

set -e

COMMIT_MSG="$1"

echo "🔍 LIBR Commit Message Validator"
echo "================================="

if [ -z "$COMMIT_MSG" ]; then
    echo ""
    echo "❌ No commit message provided!"
    echo ""
    echo "📖 How to use this script:"
    echo "   ./scripts/validate-commit.sh \"your commit message\""
    echo ""
    echo "✅ Examples of good commit messages:"
    echo "   ./scripts/validate-commit.sh \"feat: add user login button\""
    echo "   ./scripts/validate-commit.sh \"Fix: correct spelling in README\""
    echo "   ./scripts/validate-commit.sh \"docs: update installation guide\""
    echo ""
    echo "📚 Need help? Check out: docs/BEGINNER_GUIDE.md"
    exit 1
fi

# Pattern for beginners - case insensitive type matching
PATTERN="^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(.+\))?(!)?: .{1,100}"

echo "💬 Your message: \"$COMMIT_MSG\""
echo ""

# Convert message to lowercase for pattern matching
COMMIT_MSG_LOWER=$(echo "$COMMIT_MSG" | tr '[:upper:]' '[:lower:]')

if [[ $COMMIT_MSG_LOWER =~ $PATTERN ]]; then
    echo "✅ Perfect! Your commit message follows the correct format!"
    echo ""
    
    # Extract and display parts (normalize type to lowercase)
    TYPE=$(echo "$COMMIT_MSG" | sed -n 's/^\([^(: ]*\).*/\1/p' | tr '[:upper:]' '[:lower:]')
    SCOPE=$(echo "$COMMIT_MSG" | sed -n 's/^[^(]*(\([^)]*\)).*/\1/p')
    
    echo "📋 Message breakdown:"
    echo "   Type: $TYPE"
    if [ -n "$SCOPE" ]; then
        echo "   Scope: $SCOPE"
    fi
    
    # Explain what this type means
    case $TYPE in
        "feat")
            echo "   ✨ You're adding a new feature - awesome!"
            ;;
        "fix")
            echo "   🐛 You're fixing a bug - great work!"
            ;;
        "docs")
            echo "   📚 You're improving documentation - very helpful!"
            ;;
        "style")
            echo "   💅 You're updating styles/formatting - looks good!"
            ;;
        "refactor")
            echo "   🔧 You're improving code structure - nice cleanup!"
            ;;
        "test")
            echo "   🧪 You're adding tests - excellent for quality!"
            ;;
        "chore")
            echo "   🏠 You're doing maintenance work - much appreciated!"
            ;;
        "perf")
            echo "   ⚡ You're improving performance - fantastic!"
            ;;
        "build")
            echo "   🔨 You're updating build configuration - great!"
            ;;
        "ci")
            echo "   🚀 You're improving CI/CD - excellent!"
            ;;
        "revert")
            echo "   ↩️ You're reverting changes - sometimes necessary!"
            ;;
    esac
    
    # Check for breaking change
    if [[ $COMMIT_MSG == *"!"* ]] || [[ $COMMIT_MSG == *"BREAKING CHANGE"* ]]; then
        echo "   ⚠️  Breaking change detected - make sure this is intentional!"
    fi
    
    echo ""
    echo "🚀 You're ready to commit! Your message will be included in our changelog."
    
    exit 0
else
    echo "❌ Oops! Your commit message doesn't follow our format."
    echo ""
    echo "😅 Don't worry - this is easy to fix!"
    echo ""
    echo "📋 The correct format is: type: description"
    echo ""
    echo "🏷️  Available types (case insensitive):"
    echo "   • feat/Feat:     Adding a new feature"
    echo "   • fix/Fix:       Fixing a bug"
    echo "   • docs/Docs:     Updating documentation"
    echo "   • style/Style:   Changing colors, fonts, layout"
    echo "   • test/Test:     Adding or fixing tests"
    echo "   • refactor/Refactor: Improving code structure"
    echo "   • perf/Perf:     Performance improvements"
    echo "   • build/Build:   Build system changes"
    echo "   • ci/CI:         CI/CD configuration"
    echo "   • chore/Chore:   Maintenance tasks"
    echo ""
    echo "✅ Good examples:"
    echo "   feat: add dark mode toggle"
    echo "   Fix: correct login button alignment"
    echo "   DOCS: update README installation steps"
    echo "   style: change header background color"
    echo ""
    echo "❌ What's wrong with your message:"
    echo "   \"$COMMIT_MSG\""
    echo ""
    echo "💡 Quick fixes:"
    echo "   • Make sure you start with a type (feat, fix, docs, etc.)"
    echo "   • Types can be lowercase or capitalized (feat or Feat)"
    echo "   • Add a colon (:) after the type"
    echo "   • Add a space after the colon"
    echo "   • Keep it under 100 characters"
    echo ""
    echo "📚 Need more help? Check out: docs/BEGINNER_GUIDE.md"
    
    exit 1
fi
