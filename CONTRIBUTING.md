# Contributing to LIBR

First off, thank you for considering contributing to LIBR! It's people like you that make LIBR such a great tool.

## Code of Conduct

By participating in this project, you are expected to uphold our [Code of Conduct](CODE_OF_CONDUCT.md). Please read it before contributing.

## How Can I Contribute?

### Reporting Bugs

This section guides you through submitting a bug report for LIBR. Following these guidelines helps maintainers and the community understand your report, reproduce the behavior, and find related reports.

- **Use a clear and descriptive title** for the issue to identify the problem.
- **Describe the exact steps which reproduce the problem** in as many details as possible.
- **Provide specific examples to demonstrate the steps**. Include links to files or GitHub projects, or copy/pasteable snippets, which you use in those examples.
- **Describe the behavior you observed after following the steps** and point out what exactly is the problem with that behavior.
- **Explain which behavior you expected to see instead and why.**
- **Include screenshots and animated GIFs** which show you following the described steps and clearly demonstrate the problem.
- **If the problem wasn't triggered by a specific action**, describe what you were doing before the problem happened.

### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion for LIBR, including completely new features and minor improvements to existing functionality.

- **Use a clear and descriptive title** for the issue to identify the suggestion.
- **Provide a step-by-step description of the suggested enhancement** in as many details as possible.
- **Provide specific examples to demonstrate the steps**. Include copy/pasteable snippets which you use in those examples.
- **Describe the current behavior** and **explain which behavior you expected to see instead** and why.
- **Include screenshots and animated GIFs** which help you demonstrate the steps or point out the part of LIBR which the suggestion is related to.
- **Explain why this enhancement would be useful** to most LIBR users.
- **List some other applications where this enhancement exists.**
- **Specify which version of LIBR you're using.**

### Pull Requests

The process described here has several goals:

- Maintain LIBR's quality
- Fix problems that are important to users
- Engage the community in working toward the best possible LIBR
- Enable a sustainable system for LIBR's maintainers to review contributions

Please follow these steps to have your contribution considered by the maintainers:

1. **Follow all instructions in [the template](/.github/PULL_REQUEST_TEMPLATE.md)**
2. **Follow the [styleguides](#styleguides)**
3. **After you submit your pull request, verify that all [status checks](https://help.github.com/articles/about-status-checks/) are passing**

While the prerequisites above must be satisfied prior to having your pull request reviewed, the reviewer(s) may ask you to complete additional design work, tests, or other changes before your pull request can be ultimately accepted.

## Styleguides

### Git Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters or less
* Reference issues and pull requests liberally after the first line
* Consider starting the commit message with an applicable emoji:
    * 🎨 `:art:` when improving the format/structure of the code
    * ⚡️ `:zap:` when improving performance
    * 🔒 `:lock:` when dealing with security
    * 📝 `:memo:` when writing docs
    * 🐛 `:bug:` when fixing a bug
    * 🔥 `:fire:` when removing code or files
    * 💚 `:green_heart:` when fixing the CI build
    * ✅ `:white_check_mark:` when adding tests
    * 🚀 `:rocket:` when deploying stuff
    * ⬆️ `:arrow_up:` when upgrading dependencies
    * ⬇️ `:arrow_down:` when downgrading dependencies
    * 👕 `:shirt:` when removing linter warnings

### Changelog Management

We maintain a [CHANGELOG.md](CHANGELOG.md) following the [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format. 

#### For Contributors:
- **DO NOT** directly edit the changelog
- Use conventional commit messages that can be automatically parsed
- Your changes will be automatically added to the changelog during release

#### Conventional Commit Format:
```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: A new feature (correlates with MINOR in Semantic Versioning)
- `fix`: A bug fix (correlates with PATCH in Semantic Versioning)
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `build`: Changes that affect the build system or external dependencies
- `ci`: Changes to our CI configuration files and scripts
- `chore`: Other changes that don't modify src or test files
- `revert`: Reverts a previous commit

**Examples:**
```bash
feat(core): add DHT node discovery mechanism
fix(web-client): resolve connection timeout issues
docs(readme): update installation instructions
perf(core): optimize message validation algorithm
```

#### Breaking Changes:
For breaking changes, add `!` after the type/scope:
```bash
feat(core)!: redesign moderation quorum API
```

Or include `BREAKING CHANGE:` in the footer:
```bash
feat(core): redesign moderation quorum API

BREAKING CHANGE: The moderationQuorum.validate() method now requires
an additional parameter for consensus threshold.
```

#### Commit Message Validation

We provide a script to validate your commit messages locally:

```bash
# Validate a commit message
./scripts/validate-commit.sh "feat(core): add DHT node discovery mechanism"

# Example of invalid message
./scripts/validate-commit.sh "added new feature"  # ❌ Will fail

# Example of valid messages
./scripts/validate-commit.sh "feat(core): add DHT node discovery"     # ✅ Valid
./scripts/validate-commit.sh "fix(web-client): resolve timeout"       # ✅ Valid
./scripts/validate-commit.sh "docs: update installation guide"        # ✅ Valid
```

This helps ensure your commits will be properly included in the automated changelog generation.

### Code Styleguide

#### Go

* Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
* Follow the [Go Style Guide](https://golang.org/doc/effective_go.html)
* Document all functions and packages

#### JavaScript/React

* Use semicolons
* 2 spaces for indentation
* Prefer `'` over `"`
* Use [Prettier](https://prettier.io/) for formatting

#### Solidity

* Follow the [Solidity Style Guide](https://docs.soliditylang.org/en/latest/style-guide.html)
* Document all functions and contracts with NatSpec

#### Flutter/Dart

* Follow the [Dart Style Guide](https://dart.dev/guides/language/effective-dart/style)
* Use the Dart formatter (`dart format`)

## Setting Up Development Environment

### Project Structure

All source code is organized under the `src/` directory with specific language requirements:

```
src/
├── core/           # Go (1.21+) - Core protocol implementation
├── web-client/     # React/TypeScript - Web interface  
├── mobile-client/  # Flutter/Dart - Mobile application
├── contracts/      # Solidity - Smart contracts
└── tests/          # Integration and end-to-end tests
```

### Language Guidelines by Directory

- **`src/core/`**: 
  - **Language**: Go 1.21+
  - **Purpose**: Core LIBR protocol, DHT operations, cryptographic functions, peer-to-peer networking
  - **Standards**: Follow Go official style guide and effective Go practices

- **`src/web-client/`**: 
  - **Language**: React with TypeScript
  - **Purpose**: User-facing web application with modern UI/UX
  - **Standards**: Use TypeScript strict mode, React functional components with hooks

- **`src/mobile-client/`**: 
  - **Language**: Flutter/Dart
  - **Purpose**: Cross-platform mobile application
  - **Standards**: Follow Flutter/Dart style guide, use provider for state management

- **`src/contracts/`**: 
  - **Language**: Solidity 0.8.20+
  - **Purpose**: Ethereum smart contracts for global state management
  - **Standards**: Follow Solidity style guide, comprehensive NatSpec documentation

- **`src/tests/`**: 
  - **Languages**: Mixed (Go/JS/Dart based on component being tested)
  - **Purpose**: Integration tests and test utilities
  - **Standards**: Follow testing best practices for each respective language

### Core Protocol

```bash
cd src/core
go mod download
```

### Web Client

```bash
cd src/web-client
npm install
```

### Mobile Client

```bash
cd src/mobile-client
flutter pub get
```

### Smart Contracts

```bash
cd src/contracts
npm install
```

## Testing

### Core Protocol

```bash
cd src/core
go test -v ./...
```

### Web Client

```bash
cd src/web-client
npm test
```

### Mobile Client

```bash
cd src/mobile-client
flutter test
```

### Smart Contracts

```bash
cd src/contracts
npx hardhat test
```

## Additional Notes

### Release Process

LIBR follows [Semantic Versioning](https://semver.org/) and automated changelog generation:

1. **Automated Changelog**: The changelog is automatically updated during releases based on conventional commit messages
2. **Version Bumping**: Versions are determined automatically:
   - `fix:` commits trigger PATCH releases (0.1.0 → 0.1.1)
   - `feat:` commits trigger MINOR releases (0.1.0 → 0.2.0)
   - `BREAKING CHANGE:` triggers MAJOR releases (0.1.0 → 1.0.0)
3. **Release Notes**: Generated automatically from commit messages and PR descriptions

### Maintainer Responsibilities

For maintainers creating releases:
1. Ensure all PRs use conventional commit messages
2. Review and merge PRs following the pull request guidelines
3. Create releases through GitHub's release interface
4. The changelog will be automatically updated with properly categorized changes

**📚 For detailed maintainer guidance:** See [docs/MAINTAINER_GUIDE.md](docs/MAINTAINER_GUIDE.md) for best practices on working with new contributors.

### Issue and Pull Request Labels

This section lists the labels we use to help us track and manage issues and pull requests.

* `bug` - Issues with the code
* `documentation` - Issues with the documentation
* `enhancement` - Feature requests
* `good first issue` - Good for newcomers
* `help wanted` - Extra attention is needed
* `question` - Further information is requested
* `wontfix` - This will not be worked on

## 🌿 Branch Naming & Workflow Guidelines

To keep our project organized, please follow these simple guidelines for naming your branches:

### 📝 Simple Branch Naming

**Keep it simple and descriptive!** Use names that clearly explain what you're working on:

```bash
# ✅ Good examples - simple and clear
add-login-button
fix-mobile-crash
update-readme
improve-dark-mode
add-search-feature
fix-typo-in-docs

# ❌ Avoid these - too vague or complex  
my-changes
stuff
fix
new-feature
type/scope/very-long-description-that-is-hard-to-read
```

### 💡 Branch Naming Tips

**For new contributors:**
- **Be descriptive** but keep it short
- **Use dashes** instead of spaces (`add-login` not `add login`)
- **Start with action words** like `add`, `fix`, `update`, `improve`
- **Don't worry about perfect formatting** - clarity is more important!

**Examples by type of work:**
```bash
# Adding something new
add-user-profile
add-search-bar
add-mobile-support

# Fixing bugs
fix-login-error
fix-broken-link
fix-mobile-layout

# Updating documentation
update-installation-guide
fix-readme-typos
add-api-docs

# Improving existing features
improve-loading-speed
enhance-user-interface
optimize-database-queries
```

### 🔄 Workflow Steps

1. **Fork and Clone**
   ```bash
   # Fork the repository on GitHub, then:
   git clone https://github.com/YOUR_USERNAME/libr.git
   cd libr
   git remote add upstream https://github.com/devlup-labs/libr.git
   ```

2. **Create a Feature Branch**
   ```bash
   # Always start from the latest main branch
   git checkout main
   git pull upstream main
   
   # Create your feature branch with a simple, descriptive name
   git checkout -b add-user-profile
   ```

3. **Make Your Changes**
   ```bash
   # Make your changes, then stage them
   git add .
   
   # Use conventional commit messages
   git commit -m "feat: add user profile page with avatar upload"
   ```

4. **Keep Your Branch Updated**
   ```bash
   # Regularly sync with main to avoid conflicts
   git fetch upstream
   git rebase upstream/main
   ```

5. **Push and Create PR**
   ```bash
   # Push your branch
   git push origin feat/web/user-profile-page
   
   # Then create a Pull Request on GitHub
   ```

### ✅ Branch Best Practices

- **Keep branches focused** - One feature/fix per branch
- **Use descriptive names** - Others should understand what you're working on
- **Keep branches short-lived** - Aim to merge within a few days
- **Delete merged branches** - Clean up after your PR is merged
- **Rebase instead of merge** - Keep history clean

### 🚫 Branch Names to Avoid

```bash
# Too vague - what are you fixing/updating?
fix
update
my-changes
stuff

# Too long - keep it concise
create-a-new-user-authentication-system-with-oauth-and-jwt-tokens

# Hard to read - use dashes, not underscores or capitals
fix_login_bug
NEW-FEATURE
userauth
```

### 🔧 Branch Management Commands

```bash
# List all branches
git branch -a

# Delete local branch after merge
git branch -d add-user-profile

# Delete remote branch  
git push origin --delete add-user-profile

# Clean up merged branches
git branch --merged main | grep -v main | xargs -n 1 git branch -d
```

Thank you for your contributions to LIBR!
