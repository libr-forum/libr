# LIBR

[![CI](https://github.com/devlup-labs/libr/actions/workflows/ci.yml/badge.svg)](https://github.com/devlup-labs/libr/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

A Moderated, Censorship-Resilient Social Network Framework

## Overview

LIBR is a protocol for building digital public forums and social networks that are both provably censorship-resilient and safeguarded against harmful or illegal content.

Traditional centralized platforms (Facebook, Reddit, Twitter, etc.) control their own databases, which lets them remove or block content at will—undermining free expression. Fully decentralized networks avoid that single point of control, but without any moderation they may become overrun by offensive or malicious posts, turning communities chaotic rather than constructive.

LIBR strikes a balance by using a replicated DHT (Distributed Hash Table) setting for partial immutability—cheaper and faster than storing every message on a full blockchain—while storing necessary global configuration on a public Blockchain (eg., Ethereum). At the same time, content, for each community, is vetted (or approved) by a decentralized moderation quorum (a majority of moderators), so that no single moderator can decide the fate of a message. Only when a majority of moderators approve does a message get stored and shared, keeping the forum both open and safe.

## 🚀 New Contributors Welcome!

**First time contributing to open source?** We're here to help! 

👉 **Start with our [Beginner Guide](docs/BEGINNER_GUIDE.md)** - it has everything you need to get started, explained in simple terms.

🏷️ **Look for `good first issue` labels** in our [Issues page](https://github.com/devlup-labs/libr/issues) - these are perfect for newcomers!

💬 **Questions?** Don't hesitate to ask! Create a new issue or contact our mentors.

## Architecture

LIBR is built with the following components:

1. **Protocol and Networking Layer (Go)**: The backbone of the system, implementing the DHT, cryptographic operations, moderation quorum mechanisms, and peer-to-peer communication.
2. **Blockchain Layer (Solidity)**: Smart contracts that manage global state, moderator registry, and community governance and incentivization.
3. **Web Client (React)**: User-friendly interface for interacting with LIBR communities.
4. **Mobile Client (Flutter)**: Native mobile experience for broader accessibility.

## Tech Stack

- **Smart Contracts**: Solidity
- **Blockchain Interface**: Go Ethereum
- **Protocol and Networking Logic**: Go Lang
- **Web Client**: React
- **Mobile Client**: Flutter

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- Flutter 3.16+
- Solidity compiler
- Ethereum development environment (Hardhat, Truffle, etc.)


## Project Structure

All source code is organized under the `src/` directory:

```
src/
├── core/           # Go - Core protocol implementation
├── web-client/     # React/TypeScript - Web interface
├── mobile-client/  # Flutter/Dart - Mobile application
├── contracts/      # Solidity - Smart contracts
└── tests/          # Integration and end-to-end tests
```

### Language Guidelines by Directory

- **`src/core/`**: Go (1.21+) - Implements the core LIBR protocol, DHT operations, cryptographic functions, and peer-to-peer networking
- **`src/web-client/`**: React with TypeScript - User-facing web application with modern UI/UX
- **`src/mobile-client/`**: Flutter/Dart - Cross-platform mobile application
- **`src/contracts/`**: Solidity - Ethereum smart contracts for global state management
- **`src/tests/`**: Mixed (Go/JS/Dart) - Integration tests and test utilities

### Running the Components

```bash
# Core protocol
cd src/core
go run main.go

# Web client
cd src/web-client
npm start

# Mobile client
cd src/mobile-client
flutter run

# Smart contracts (local development)
cd src/contracts
npx hardhat node
```

## 🛠️ Helpful Tools for Contributors

We've created some tools to make contributing easier:

```bash
# 🚀 Quick project setup
./scripts/setup.sh

# 🔍 Check if your commit message is correct
./scripts/validate-commit.sh "feat: add new feature"

# Examples:
./scripts/validate-commit.sh "feat: add dark mode"        # ✅ Good
./scripts/validate-commit.sh "fix: button not working"    # ✅ Good  
./scripts/validate-commit.sh "added new stuff"            # ❌ Bad format
```

## Development Roadmap

- [x] Prototype implementation
- [ ] Blockchain integration with Ethereum
- [ ] Complete web client implementation
- [ ] Mobile client development
- [ ] Governance model implementation
- [ ] Core protocol optimization
- [ ] Comprehensive testing and security audits
- [ ] Public beta launch

## Contributing

We welcome contributions from the community! Please check out our [Contributing Guidelines](CONTRIBUTING.md) for details on how to get involved.

## Documentation

For more detailed information about the LIBR protocol and its implementation, check out:

- [Concept Document](https://github.com/devlup-labs/libr/blob/main/docs/libr_soc_concept_doc.pdf)
- [Project Documentation](https://github.com/devlup-labs/libr/blob/main/docs/libr_project_doc.pdf)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

