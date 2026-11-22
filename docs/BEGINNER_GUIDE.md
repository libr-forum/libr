# 🚀 New Contributor Quick Start Guide

Welcome! This guide gets you from zero to your first pull request.

## 📋 Table of Contents

1. [First time setup](#-first-time-setup)
2. [Understanding the project](#-understanding-the-project)
3. [How to contribute](#-how-to-contribute)
4. [Getting help](#-getting-help)

## 🔧 First Time Setup

### 1. Fork and clone

```bash
# 1. Click "Fork" on GitHub
# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/Libr.git
cd Libr
# 3. Track the original repository
git remote add upstream https://github.com/libr-forum/Libr.git
```

### 2. Install the tools

You only need the ones for the part you work on:

| Working on | You need |
|---|---|
| Go code (`core/`, `services/relay`) | Go 1.24.4 |
| Desktop app (`core/mod_client`) | Go, Node 22, [Wails v2](https://wails.io/docs/gettingstarted/installation) |
| Android app (`mobile-app`) | Node 22, JDK 17, Android SDK; Go for the bridge |
| Seed server (`services/librserver`) | Node 22 |
| End-to-end tests (`testbed/`) | Docker |

### 3. Check your setup

```bash
./scripts/setup.sh
```

It reports what is missing and installs the npm dependencies.

## 📁 Understanding the Project

LIBR is a decentralized forum: posts live on many storage ("DB") nodes, and a post is
shown only if enough of the community's moderators signed it. The
[README](../README.md#architecture) has the full picture. Where things are:

```
core/libr/           shared Go core (start here for protocol work)
  proto/             signed objects and their verification
  forum/             posting, reading, reporting (the client)
  moderator/         how moderators vote; human review queue
  dbnode/            what storage nodes accept and serve
  app/               the node both apps run
  cmd/librctl/       command-line tool
core/db/             storage node program
core/mod_client/     desktop app (Go + React in frontend/)
mobile-app/          Android app (React Native; Go bridge in go-core/)
services/relay/      relay for peers behind NAT
services/librserver/ seed list server (Node)
testbed/             Docker test network and LAN dev network
```

## 🤝 How to Contribute

### Step 1: pick an issue

Look for `good first issue` or `documentation` labels and comment that you would
like to work on it.

### Step 2: create a branch with a good name

```bash
git checkout staging
git pull upstream staging
git checkout -b fix-feed-refresh      # short, describes the change
```

Good: `add-community-search`, `fix-report-dialog`, `update-readme`.
Avoid: `my-changes`, `fix`, `new-stuff`.

### Step 3: make the change and test it

Run the tests for what you touched (full list in
[CONTRIBUTING.md](../CONTRIBUTING.md#testing)), for example:

```bash
cd core/libr && go test ./...                         # Go
cd core/mod_client/frontend && npm test               # desktop UI
cd mobile-app && npx tsc --noEmit && npx expo lint    # mobile
```

Want to see it working? `DEVNET_IP=<your ip> testbed/devnet.sh start` starts a small
network on your machine that the apps can join.

### Step 4: commit with a conventional message

```bash
git commit -m "fix: refresh the feed after posting"
./scripts/validate-commit.sh "fix: refresh the feed after posting"   # check a message
```

Types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`.

### Step 5: open a pull request

```bash
git push origin fix-feed-refresh
```

Open a PR against `staging`, describe what changed and how you tested it. CI runs the
same checks as above; a maintainer will review.

## 💬 Getting Help

* Ask in the issue you are working on: there are no silly questions.
* Read [CONTRIBUTING.md](../CONTRIBUTING.md) for the details.
* Stuck on setup? Paste the output of `./scripts/setup.sh` in your question.

Thank you for contributing! 🎉
