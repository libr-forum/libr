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

* `gofmt`; follow [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments).
* Lint: `golangci-lint run --config=.golangci.yml ./...` (v1.64.8, as in CI). It checks only code changed since `origin/staging`.
* Network handler packages (`core/libr/{app,dbnode,moderator,wire,forum,p2p,proto,seeds}`) must not print: stray prints can log request bodies.
* Anything signed goes through `proto.Encode` with its own domain string; never concatenate fields by hand.

#### TypeScript / React (desktop frontend, mobile app)

* ESLint as configured in each package (`npm run lint`, `npx expo lint`); TypeScript must typecheck.
* Call the Go core only through the typed wrappers (`frontend/src/services/api.ts`, `mobile-app/modules/LibrCore.ts`), never with `as any`.

## Setting Up Development Environment

### Project Structure

| Path | What | Stack |
|---|---|---|
| `core/libr` | Shared Go core: protocol objects, networking, DB node logic, moderator, client, `librctl` CLI | Go |
| `core/db` | DB (storage) node binary, SQLite store | Go |
| `core/mod_client` | Desktop app (Wails) | Go + React/Vite |
| `mobile-app` | Android app (Expo / React Native) and its gomobile bridge (`go-core/`) | TS + Go |
| `services/relay` | Circuit relay | Go |
| `services/librserver` | Bootstrap seed list | Node |
| `testbed/` | Docker end-to-end network, LAN dev network | shell + compose |
| `deploy/` | Production compose for relays, DB nodes, seed server | compose |

See [README.md](README.md#architecture) for how the pieces talk to each other.

### Toolchain

* Go 1.24.4 (the root `go.work` ties the Go modules together)
* Node 22
* Docker (testbed, images)
* Desktop: Wails v2 CLI; on Ubuntu 24.04 build with `-tags webkit2_41`
* Mobile: JDK 17, Android SDK + NDK 26.3.11579264. gomobile is pinned in `mobile-app/go-core/go.mod`; install `gobind` with `go install golang.org/x/mobile/cmd/gobind` from that directory and **do not** run `gomobile init` (it fetches `@latest`).

`./scripts/setup.sh` checks the toolchain and installs the npm dependencies.

## Testing

Run what CI runs for the parts you touched:

```bash
# Go: in each of core/libr, core/db, core/mod_client, mobile-app/go-core, services/relay
go vet ./... && go test -race ./...
# (core/mod_client embeds frontend/dist: `mkdir -p frontend/dist && touch frontend/dist/.keep` first)

# Desktop frontend
cd core/mod_client/frontend && npm ci && npm run lint && npm run typecheck && npm test

# Mobile
cd mobile-app && npm ci && npx expo lint && npx tsc --noEmit

# Seed server
cd services/librserver && npm ci && npm test

# Whole network in Docker (relays, NATed DB nodes, moderators, clients, faults)
testbed/e2e.sh
```

API pins that fail on purpose when you change an interface:
* `core/mod_client/app_api_test.go`: Go `App` methods vs `frontend/wailsjs/go/main/App.d.ts` (regenerate with `wails generate module -tags webkit2_41`).
* `mobile-app/go-core/bridge/api_test.go`: the bridge API golden file (`UPDATE_GOLDEN=1 go test ./bridge`) and every Kotlin `callBridgeString` call.

To try the apps against a real network on your LAN: `DEVNET_IP=<your ip> testbed/devnet.sh start` (see the script header).

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
3. Create releases by pushing a `v*` tag; follow [docs/RELEASE.md](docs/RELEASE.md)
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
