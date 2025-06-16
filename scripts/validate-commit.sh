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
    echo "   ./scripts/validate-commit.sh \"fix: correct spelling in README\""
    echo "   ./scripts/validate-commit.sh \"docs: update installation guide\""
    echo ""
    echo "📚 Need help? Check out: docs/BEGINNER_GUIDE.md"
    exit 1
fi

# Simple pattern for beginners - just type: description
PATTERN="^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(.+\))?(!)?: .{1,100}"

echo "💬 Your message: \"$COMMIT_MSG\""
echo ""

if [[ $COMMIT_MSG =~ $PATTERN ]]; then
    echo "✅ Perfect! Your commit message follows the correct format!"
    echo ""
    
    # Extract and display parts
    TYPE=$(echo "$COMMIT_MSG" | sed -n 's/^\([^(: ]*\).*/\1/p')
    SCOPE=$(echo "$COMMIT_MSG" | sed -n 's/^[^(]*(\([^)]*\)).*/\1/p')
    
    echo "📋 Message breakdown:"
    echo "   Type: $TYPE"
    if [ -n "$SCOPE" ]; then
        echo "   Scope: $SCOPE"
    fi
    
    # Explain what this type means
    case $TYPE in
        "feat")
            echo "   📝 Meaning: You're adding a new feature!"
            ;;
        "fix")
            echo "   🐛 Meaning: You're fixing a bug!"
            ;;
        "docs")
            echo "   📚 Meaning: You're updating documentation!"
            ;;
        "style")
            echo "   💅 Meaning: You're improving the appearance!"
            ;;
        "test")
            echo "   🧪 Meaning: You're adding or fixing tests!"
            ;;
        *)
            echo "   🔧 Meaning: You're making other improvements!"
            ;;
    esac
    
    # Check for breaking change
    if [[ $COMMIT_MSG == *"!"* ]] || [[ $COMMIT_MSG == *"BREAKING CHANGE"* ]]; then
        echo ""
        echo "⚠️  Breaking change detected!"
        echo "   This means your change might break existing code."
        echo "   Make sure this is intentional!"
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
    echo "🏷️  Available types:"
    echo "   • feat:     Adding a new feature"
    echo "   • fix:      Fixing a bug"
    echo "   • docs:     Updating documentation"
    echo "   • style:    Changing colors, fonts, layout"
    echo "   • test:     Adding or fixing tests"
    echo "   • refactor: Improving code structure"
    echo ""
    echo "✅ Good examples:"
    echo "   feat: add dark mode toggle"
    echo "   fix: correct login button alignment"
    echo "   docs: update README installation steps"
    echo "   style: change header background color"
    echo ""
    echo "❌ What's wrong with your message:"
    echo "   \"$COMMIT_MSG\""
    echo ""
    echo "💡 Quick fixes:"
    echo "   • Make sure you start with a type (feat, fix, docs, etc.)"
    echo "   • Add a colon (:) after the type"
    echo "   • Add a space after the colon"
    echo "   • Keep it under 100 characters"
    echo ""
    echo "📚 Need more help? Check out: docs/BEGINNER_GUIDE.md"
    exit 1
fi
