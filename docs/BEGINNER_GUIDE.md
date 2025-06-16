# 🚀 New Contributor Quick Start Guide

Welcome to the LIBR project! This guide will help you get started contributing, even if you're new to open source development.

## 📋 Table of Contents
1. [First Time Setup](#first-time-setup)
2. [Understanding the Project Structure](#understanding-the-project-structure)
3. [How to Contribute](#how-to-contribute)
4. [Writing Good Commit Messages](#writing-good-commit-messages)
5. [Creating Issues](#creating-issues)
6. [Making Pull Requests](#making-pull-requests)
7. [Getting Help](#getting-help)

## 🔧 First Time Setup

### 1. Fork and Clone the Repository
```bash
# 1. Click "Fork" button on GitHub to create your copy
# 2. Clone your fork to your computer
git clone https://github.com/YOUR_USERNAME/libr.git
cd libr

# 3. Add the original repository as upstream
git remote add upstream https://github.com/devlup-labs/libr.git
```

### 2. Install Required Tools
Make sure you have these installed on your computer:

- **Git** - [Download here](https://git-scm.com/downloads)
- **Go 1.21+** - [Download here](https://golang.org/dl/)
- **Node.js 18+** - [Download here](https://nodejs.org/)
- **Flutter 3.16+** - [Download here](https://flutter.dev/docs/get-started/install)

### 3. Test Your Setup
```bash
# Check if everything is installed correctly
go version      # Should show Go 1.21 or higher
node --version  # Should show Node 18 or higher
flutter --version  # Should show Flutter 3.16 or higher
```

## 📁 Understanding the Project Structure

LIBR is organized into different folders for different programming languages:

```
src/
├── core/           # 🟢 Go code - Backend logic
├── web-client/     # 🔵 React/TypeScript - Website
├── mobile-client/  # 🟣 Flutter/Dart - Mobile app
├── contracts/      # 🟡 Solidity - Blockchain contracts
└── tests/          # 🔴 All tests
```

**Which folder should you work in?**
- **New to programming?** Start with `src/web-client/` (website)
- **Know JavaScript/TypeScript?** Work in `src/web-client/`
- **Know mobile development?** Work in `src/mobile-client/`
- **Know Go/backend?** Work in `src/core/`
- **Know blockchain?** Work in `src/contracts/`

## 🤝 How to Contribute

### Step 1: Pick an Issue
1. Go to the [Issues page](https://github.com/devlup-labs/libr/issues)
2. Look for issues labeled `good first issue` - these are perfect for beginners!
3. Comment "I'd like to work on this" on the issue you choose

### Step 2: Create a Branch
```bash
# Always create a new branch for your work
git checkout -b my-feature-name

# Examples of good branch names:
# git checkout -b add-user-login
# git checkout -b fix-button-color
# git checkout -b update-readme
```

### Step 3: Make Your Changes
- Work on your feature/fix in the appropriate `src/` folder
- Test your changes locally
- Make sure your code works before submitting

### Step 4: Commit Your Changes
Use our simple commit message format (more details below):
```bash
git add .
git commit -m "feat: add user login button"
```

### Step 5: Push and Create Pull Request
```bash
git push origin my-feature-name
# Then go to GitHub and click "Create Pull Request"
```

## ✍️ Writing Good Commit Messages

We use a simple format that helps us track what changes were made:

### Format: `type: description`

**Types you can use:**
- `feat:` - Adding a new feature
- `fix:` - Fixing a bug
- `docs:` - Updating documentation
- `style:` - Changing colors, fonts, layout
- `test:` - Adding tests

### ✅ Good Examples:
```bash
feat: add login button to homepage
fix: correct spelling mistake in README
docs: update installation instructions
style: change button color to blue
test: add test for user registration
```

### ❌ Bad Examples:
```bash
"fixed stuff"           # Not descriptive
"Added feature"         # Missing type
"feat added login"      # Wrong format
```

### 🔍 Check Your Commit Message
Before committing, you can test your message:
```bash
# Test if your commit message is correct
./scripts/validate-commit.sh "feat: add login button"
```

If it shows ✅, you're good to go! If it shows ❌, fix the format.

## 🐛 Creating Issues

Found a bug or have an idea? Create an issue!

### For Bugs:
1. Click "Issues" → "New Issue"
2. Choose "Bug report"
3. Fill out the template with:
   - What you expected to happen
   - What actually happened
   - Steps to reproduce the bug

### For New Features:
1. Click "Issues" → "New Issue"
2. Choose "Feature request"
3. Describe your idea and why it would be useful

### For Questions:
1. Click "Issues" → "New Issue"
2. Choose "Question"
3. Ask away! No question is too basic.

## 🔄 Making Pull Requests

When your code is ready:

1. **Push your branch** to your fork
2. **Go to GitHub** and click "Compare & pull request"
3. **Fill out the template:**
   - Describe what your PR does
   - Check the boxes that apply
   - Link to any related issues
4. **Wait for review** - maintainers will look at your code
5. **Make changes if requested** - it's normal to get feedback!

### PR Checklist (simplified):
- [ ] My code works locally
- [ ] I tested my changes
- [ ] My commit messages follow the format
- [ ] I filled out the PR template

## 🆘 Getting Help

### Stuck? Here's how to get help:

1. **Check existing issues** - someone might have the same problem
2. **Read the documentation** - look at README.md and CONTRIBUTING.md
3. **Ask questions** - create a new issue with the "Question" template
4. **Contact mentors:**
   - Aradhya Mahajan: +91 90581 38511
   - Lakshya Jain: +91 79761 23107

### Common Beginner Questions:

**Q: I made a mistake in my commit message. What do I do?**
A: You can fix the last commit message with:
```bash
git commit --amend -m "new correct message"
```

**Q: How do I update my fork with the latest changes?**
A: 
```bash
git checkout main
git fetch upstream
git merge upstream/main
git push origin main
```

**Q: My code isn't working. Where do I ask for help?**
A: Create a new issue with the "Question" template and describe your problem!

**Q: Can I work on multiple issues at the same time?**
A: It's better to finish one issue before starting another, especially when you're learning.

## 🎉 Your First Contribution

Ready to make your first contribution? Here's what to do:

1. **Find a "good first issue"** - these are designed for newcomers
2. **Comment on the issue** saying you want to work on it
3. **Follow the steps above** to make your changes
4. **Don't worry about making mistakes** - that's how we learn!
5. **Ask for help** if you get stuck - we're here to help you succeed

Remember: Every expert was once a beginner. We're excited to help you learn and contribute to LIBR! 🚀

---

**Need more help?** Check out:
- [CONTRIBUTING.md](CONTRIBUTING.md) - Detailed contribution guidelines
- [README.md](README.md) - Project overview and setup
- [Issues page](https://github.com/devlup-labs/libr/issues) - Find something to work on
