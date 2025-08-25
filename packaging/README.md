# Debian Package Building for Libr

This directory contains the configuration and scripts for building Debian packages of the Libr client application.

## Files

- `nfpm.yaml` - Main nFPM configuration for amd64 packages
- `nfpm-arm64.yaml` - nFPM configuration for arm64 packages
- `../scripts/build-deb.sh` - Script to build Debian packages
- `../scripts/create-apt-repo.sh` - Script to create a local APT repository

## Prerequisites

1. Install nFPM:
   ```bash
   # Using Go
   go install github.com/goreleaser/nfpm/v2/cmd/nfpm@latest
   
   # Or using package manager (if available)
   # snap install nfpm
   ```

2. Build the Wails client executable:
   ```bash
   cd core/mod_client
   wails build
   ```

## Building Packages

### Quick Build (amd64)
```bash
./scripts/build-deb.sh
```

### Custom Version/Architecture
```bash
VERSION=1.0.1 ARCH=amd64 ./scripts/build-deb.sh
```

### Manual Build
```bash
# For amd64
nfpm pkg --packager deb --config packaging/nfpm.yaml --target dist/libr_1.0.0~beta_amd64.deb

# For arm64 (if you have the arm64 binary)
nfpm pkg --packager deb --config packaging/nfpm-arm64.yaml --target dist/libr_1.0.0~beta_arm64.deb
```

## Package Information

The generated package includes:

- **Executable**: `/usr/bin/libr` - The main Libr client application
- **Config directory**: `/etc/libr/` - System configuration directory
- **Data directory**: `/var/lib/libr/` - Application data directory
- **Log directory**: `/var/log/libr/` - Log files directory

### Dependencies
- `ca-certificates` - Required for TLS connections
- `libc6` - Standard C library

### Recommended packages
- `curl` - For HTTP operations
- `wget` - For file downloads

## Creating an APT Repository

To create your own APT repository for hosting the packages:

```bash
./scripts/create-apt-repo.sh
```

This will create a repository structure in `./apt-repo/` that you can host on any web server.

### Using the Repository

Add to your `/etc/apt/sources.list` or create `/etc/apt/sources.list.d/libr.list`:

```
deb [trusted=yes] http://your-domain.com/path-to-repo stable main
```

Then install:
```bash
sudo apt update
sudo apt install libr
```

## Package Installation

### Direct installation
```bash
sudo dpkg -i dist/libr_1.0.0~beta_amd64.deb
sudo apt-get install -f  # Fix any dependency issues
```

### Verify installation
```bash
dpkg -l | grep libr
libr --version  # Should show the application version
```

## Configuration Structure

The nFPM configuration follows the official specification and includes:

- **Package metadata**: Name, version, maintainer, description
- **Dependencies**: Required and recommended packages
- **File mappings**: Binary placement and directory creation
- **Debian-specific settings**: Architecture, compression, bug tracking

## Troubleshooting

### Binary not found
Ensure the Wails client is built and the binary exists at `./dist/libr-linux-amd64`:
```bash
ls -la dist/libr-linux-*
```

### Permission errors
Make sure scripts are executable:
```bash
chmod +x scripts/*.sh
```

### nFPM errors
Validate the configuration:
```bash
nfpm validate packaging/nfpm.yaml
```

## Advanced Usage

### Environment Variables
- `VERSION`: Package version (default: 1.0.0~beta)
- `ARCH`: Target architecture (default: amd64)
- `BUILD_DIR`: Output directory (default: dist)

### Multiple Architectures
To build for multiple architectures, you need the corresponding binaries:
1. Build for each architecture using Wails
2. Use the appropriate nFPM config file
3. Run the build script with the correct ARCH variable

Example for cross-compilation workflow:
```bash
# Build for different architectures
GOOS=linux GOARCH=amd64 wails build
GOOS=linux GOARCH=arm64 wails build

# Package for each architecture
ARCH=amd64 ./scripts/build-deb.sh
ARCH=arm64 ./scripts/build-deb.sh
```
