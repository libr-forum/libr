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

#### 🐧 Linux Installation

**Option 1: APT Repository (Ubuntu/Debian) - Recommended**
```bash
# Add GPG key for repository verification
wget -qO- https://libr-forum.github.io/libr-apt-repo/libr-repo-key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/libr-repo-key.gpg

# Add APT repository to sources
echo "deb [signed-by=/usr/share/keyrings/libr-repo-key.gpg] https://libr-forum.github.io/libr-apt-repo/ ./" | sudo tee /etc/apt/sources.list.d/libr.list

# Update package index and install
sudo apt update
sudo apt install libr
```

**Option 2: Direct Download - All Distributions**

*Ubuntu/Debian (.deb package):*
```bash
# Download the latest Debian package
wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr_1.0.0-beta_amd64.deb

# Install the package
sudo dpkg -i libr_1.0.0-beta_amd64.deb

# Fix any missing dependencies
sudo apt-get install -f
```

*Fedora/RHEL/CentOS (.rpm package):*
```bash
# Download the latest RPM package
wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr-1.0.0-beta-1.x86_64.rpm

# Install the package
sudo dnf install ./libr-1.0.0-beta-1.x86_64.rpm
# or on older systems: sudo yum install ./libr-1.0.0-beta-1.x86_64.rpm
```

*Arch Linux (.pkg.tar.zst package):*
```bash
# Download the latest Arch package
wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr-1.0.0-beta-1-x86_64.pkg.tar.zst

# Install the package
sudo pacman -U libr-1.0.0-beta-1-x86_64.pkg.tar.zst
```

**Option 3: Binary Installation**

If packages aren't available for your distribution:
```bash
# Download the binary
wget https://github.com/libr-forum/Libr/releases/download/v1.0.0-beta/libr-linux-amd64

# Make executable and install
chmod +x libr-linux-amd64
sudo mv libr-linux-amd64 /usr/local/bin/libr
```

**🔧 Solving WebKit Library Issues**

Libr uses WebKitGTK for its UI, which may require specific library versions on different distributions:

*Ubuntu 24.04 (Noble) and newer:*
```bash
# Install newer WebKitGTK packages
sudo apt update
sudo apt install -y libwebkit2gtk-4.1-0 libjavascriptcoregtk-4.1-0

# Create compatibility symlinks
sudo ln -sf /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.1.so.0 \
            /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.0.so.37

sudo ln -sf /usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.1.so.0 \
            /usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.0.so.18
```

*Fedora 35+ and newer RHEL/CentOS:*
```bash
# Install WebKitGTK packages
sudo dnf install webkit2gtk4.1-devel

# Create compatibility symlinks if needed
sudo ln -sf /usr/lib64/libwebkit2gtk-4.1.so.0 \
            /usr/lib64/libwebkit2gtk-4.0.so.37

sudo ln -sf /usr/lib64/libjavascriptcoregtk-4.1.so.0 \
            /usr/lib64/libjavascriptcoregtk-4.0.so.18
```

*Arch Linux:*
```bash
# Install WebKitGTK package
sudo pacman -S webkit2gtk-4.1

# Create compatibility symlinks
sudo ln -sf /usr/lib/libwebkit2gtk-4.1.so.0 \
            /usr/lib/libwebkit2gtk-4.0.so.37

sudo ln -sf /usr/lib/libjavascriptcoregtk-4.1.so.0 \
            /usr/lib/libjavascriptcoregtk-4.0.so.18
```

*Generic Linux (if above don't work):*
```bash
# Try installing WebKit development packages
# Ubuntu/Debian:
sudo apt install libwebkit2gtk-4.0-dev

# Fedora/RHEL/CentOS:
sudo dnf install webkit2gtk3-devel

# OpenSUSE:
sudo zypper install webkit2gtk3-devel

# Arch Linux:
sudo pacman -S webkit2gtk
```

**Alternative: Automated Installation Script**

If you prefer automatic detection and installation:
```bash
# Download and run the installation script
curl -fsSL https://raw.githubusercontent.com/libr-forum/Libr/main/scripts/install-libr.sh | bash

# Or inspect the script first (recommended for security)
wget https://raw.githubusercontent.com/libr-forum/Libr/main/scripts/install-libr.sh
chmod +x install-libr.sh
./install-libr.sh
```

The script automatically detects your distribution and handles package installation and library dependencies.

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
- **WebKit library errors on Linux**: 
  - Install the WebKitGTK packages for your distribution as shown above
  - The error typically looks like: `libwebkit2gtk-4.0.so.37: cannot open shared object file`
  - This affects Ubuntu 24.04+, modern Fedora, and Arch Linux due to library version changes
  - The symlink solutions above resolve compatibility issues
- **Package installation fails**: 
  - On Debian/Ubuntu: Run `sudo apt-get install -f` to fix dependencies
  - On Fedora/RHEL: Ensure EPEL repository is enabled for additional packages
  - On Arch Linux: Update system with `sudo pacman -Syu` before installing
- **macOS security warnings**: Allow the app in System Settings → Privacy & Security

**Distribution-specific Notes:**

- **Ubuntu 24.04+ (Noble)**: Requires WebKit 4.1 packages and compatibility symlinks
- **Fedora 35+**: May need `webkit2gtk4.1-devel` package
- **Arch Linux**: Install `webkit2gtk-4.1` or `webkit2gtk` packages
- **RHEL/CentOS 8+**: Enable PowerTools/CRB repository for WebKit packages

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

# 📦 Build all package types (Debian, RPM, Arch)
./scripts/build-packages.sh

# 📦 Build specific package type
./scripts/build-deb.sh        # Debian package only
nfpm pkg --packager rpm --config packaging/nfpm-rpm.yaml --target dist/  # RPM
nfpm pkg --packager archlinux --config packaging/nfpm-arch.yaml --target dist/  # Arch

# 🧪 Test installation instructions
./scripts/test-readme-validation.sh    # Validate README instructions
./scripts/test-installation.sh         # Test with Docker containers

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

