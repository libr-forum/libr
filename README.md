# LIBR

A Moderated, Censorship-Resilient Social Network Framework

## Overview

LIBR is a protocol for building digital public forums and social networks that are both provably censorship-resilient and safeguarded against harmful or illegal content.

Traditional centralized platforms (Facebook, Reddit, Twitter, etc.) control their own databases, which lets them remove or block content at will—undermining free expression. Fully decentralized networks avoid that single point of control, but without any moderation they may become overrun by offensive or malicious posts, turning communities chaotic rather than constructive.

LIBR strikes a balance by using a replicated DHT (Distributed Hash Table) setting for partial immutability—cheaper and faster than storing every message on a full blockchain—while storing necessary global configuration on a public Blockchain (eg., Ethereum). At the same time, content, for each community, is vetted (or approved) by a decentralized moderation quorum (a majority of moderators), so that no single moderator can decide the fate of a message. Only when a majority of moderators approve does a message get stored and shared, keeping the forum both open and safe.

## 🚀 New Contributors Welcome!

**First time contributing to open source?** We're here to help! 

👉 **Start with our [Beginner Guide](docs/BEGINNER_GUIDE.md)** - it has everything you need to get started, explained in simple terms.

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

### Installation

Choose your installation method based on your operating system:

#### 🐧 Linux (Recommended - Easy Installation)

**Option 1: Automated Installation Script (Recommended)**
```bash
# Download and run the installation script
curl -fsSL https://raw.githubusercontent.com/libr-forum/Libr/main/scripts/install-libr.sh | bash

# Or download and inspect the script first (more secure)
wget https://raw.githubusercontent.com/libr-forum/Libr/main/scripts/install-libr.sh
chmod +x install-libr.sh
./install-libr.sh
```

**What the script does:**
- 🔍 Auto-detects your Linux distribution (Ubuntu, Debian, Fedora, RHEL, CentOS, Arch Linux)
- 📦 Downloads the appropriate package format (.deb, .rpm, .pkg.tar.zst)
- ⚙️ Installs dependencies and sets up desktop integration
- ✅ Configures everything so you can run `libr` from anywhere
- 🔄 Checks for existing installations and updates if needed

**Option 2: APT Repository (Ubuntu/Debian)**
```bash
# Add Libr APT repository
curl -fsSL https://libr-forum.github.io/libr-apt-repo/setup-repo.sh | bash

# Install via package manager
sudo apt install libr
```

**Option 3: Manual Installation**

*For Ubuntu 22.04 and below:*
1. Download the **Linux build** (`libr-linux-amd64`) from [Releases](../../releases)
2. Make it executable: `chmod +x ./libr-linux-amd64`
3. Run it: `./libr-linux-amd64`

*For Ubuntu 24.04 (Noble) and newer:*
If you encounter WebKitGTK library errors, install the updated libraries:
```bash
# Update package index
sudo apt update

# Install newer WebKitGTK packages
sudo apt install -y libwebkit2gtk-4.1-0 libjavascriptcoregtk-4.1-0

# Create symlinks for compatibility
sudo ln -sf /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.1.so.0 \
            /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.0.so.37

sudo ln -sf /usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.1.so.0 \
            /usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.0.so.18

# Make executable and run
chmod +x ./libr-linux-amd64
./libr-linux-amd64
```

#### 🪟 Windows

1. Download the latest **Windows release** (`libr-win-amd64.exe`) from the [Releases](../../releases) page
2. Double-click to run it
   - If the app doesn't start, try **right-click → Run as administrator**

#### 🍎 macOS

1. Download the **macOS release** (`libr-darwin-amd64.out`) from [Releases](../../releases)
2. On first run, macOS may block the app. To fix this:
   - Go to **System Settings → Privacy & Security**
   - Allow the app under the "Security" section
3. Make executable and run:
   ```bash
   chmod +x ./libr-darwin-amd64.out
   ./libr-darwin-amd64.out
   ```

#### 📋 Verification

After installation, verify that LIBR is working:
```bash
# Check if libr is installed
libr --version

# Launch the application
libr
```

**After installation, you can:**
- Find Libr in your applications menu under "Network" → "Libr"
- Run `libr` from any terminal
- The application includes desktop integration with proper icons and shortcuts

The application should appear in your applications menu under "Network" → "Libr".

#### 🔧 Troubleshooting

**Common Issues:**

- **"Command not found" error**: Make sure the binary is in your PATH or use the full path to the executable
- **Permission denied**: Run `chmod +x` on the downloaded binary
- **Library errors on Linux**: Install the required WebKitGTK libraries as shown above
- **macOS security warnings**: Allow the app in System Settings → Privacy & Security

**Need help?** 
- 📋 [Submit feedback](https://docs.google.com/forms/d/e/1FAIpQLSdOnq6uPpLYEQIueuHtvydMI8q1CMHC_TJzDkUDUU8UCGo4ew/viewform)
- 🐛 [Report issues](https://github.com/libr-forum/Libr/issues)

---



## Project Structure

All source code is organized under the `src/` directory:

```
src/
├── core-protocol/  # Go - Core LIBR protocol and moderation logic
├── network/        # Go - P2P networking and DHT operations
├── web-client/     # React/TypeScript - Web interface
├── mobile-client/  # Flutter/Dart - Mobile application
├── contracts/      # Solidity - Smart contracts
└── tests/          # Integration and end-to-end tests
```

### Language Guidelines by Directory

- **`src/core-protocol/`**: Go (1.21+) - Core LIBR protocol implementation, moderation logic, and data structures
- **`src/network/`**: Go (1.21+) - Peer-to-peer networking, DHT operations, and node discovery
- **`src/web-client/`**: React with TypeScript - User-facing web application with modern UI/UX
- **`src/mobile-client/`**: Flutter/Dart - Cross-platform mobile application
- **`src/contracts/`**: Solidity - Ethereum smart contracts for global state management
- **`src/tests/`**: Mixed (Go/JS/Dart) - Integration tests and test utilities

### Running the Components

```bash
# Core protocol
cd src/core-protocol
go run main.go

# Network layer
cd src/network
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

# 📦 Build Debian package (Linux)
./scripts/build-deb.sh

# 🗃️ Create APT repository
./scripts/create-apt-repo.sh

# 🧪 Test APT repository
./scripts/test-repository.sh

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

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

