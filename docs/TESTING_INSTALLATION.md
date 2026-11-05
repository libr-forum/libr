# Testing Installation Instructions Across Different OS

This guide helps you test the README installation instructions across different Linux distributions.

## 🚀 Quick Testing (Recommended)

### 1. Run Validation Script
```bash
# Test all URLs, commands, and package builds
./scripts/test-readme-validation.sh
```

### 2. Run Installation Tests
```bash
# Test with Docker containers (requires Docker)
./scripts/test-installation.sh
```

## 🐳 Docker Testing (Most Comprehensive)

### Ubuntu/Debian Testing
```bash
# Test Ubuntu 22.04
docker run -it --rm ubuntu:22.04 bash
apt update && apt install -y wget curl gpg

# Follow README APT instructions
wget -qO- https://libr-forum.github.io/libr-apt-repo/libr-repo-key.gpg | gpg --dearmor -o /usr/share/keyrings/libr-repo-key.gpg
echo "deb [signed-by=/usr/share/keyrings/libr-repo-key.gpg] https://libr-forum.github.io/libr-apt-repo/ ./" > /etc/apt/sources.list.d/libr.list
apt update

# Test WebKit fix for Ubuntu 24.04
docker run -it --rm ubuntu:24.04 bash
apt update && apt install -y libwebkit2gtk-4.1-0 libjavascriptcoregtk-4.1-0
ln -sf /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.1.so.0 /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.0.so.37
```

### Fedora/RHEL Testing
```bash
# Test Fedora
docker run -it --rm fedora:38 bash
dnf update -y && dnf install -y wget curl

# Test WebKit packages
dnf install -y webkit2gtk4.1-devel
```

### Arch Linux Testing
```bash
# Test Arch Linux
docker run -it --rm archlinux:latest bash
pacman -Sy --noconfirm wget curl

# Test WebKit packages
pacman -S --noconfirm webkit2gtk-4.1
```

## 🖥️ Virtual Machine Testing

### Using VirtualBox/VMware
1. **Create VMs for each distribution:**
   - Ubuntu 22.04 LTS
   - Ubuntu 24.04 LTS  
   - Fedora 38
   - Arch Linux (latest)

2. **Test installation methods:**
   - APT repository method
   - Direct package download
   - Binary installation
   - Automated script

### Using Cloud Instances
```bash
# AWS EC2, Google Cloud, or DigitalOcean
# Launch instances with different OS:
# - Ubuntu 22.04
# - Ubuntu 24.04
# - Fedora 38
# - Arch Linux

# SSH into each and test installation
ssh user@instance-ip
# Follow README instructions
```

## 🧪 Manual Testing Checklist

### For Each Distribution:
- [ ] **APT Repository Setup** (Ubuntu/Debian only)
  ```bash
  wget -qO- https://libr-forum.github.io/libr-apt-repo/libr-repo-key.gpg | sudo gpg --dearmor -o /usr/share/keyrings/libr-repo-key.gpg
  echo "deb [signed-by=/usr/share/keyrings/libr-repo-key.gpg] https://libr-forum.github.io/libr-apt-repo/ ./" | sudo tee /etc/apt/sources.list.d/libr.list
  sudo apt update
  sudo apt install libr
  ```

- [ ] **Direct Package Download**
  ```bash
  # Test appropriate package format
  wget <package-url>
  # Install with package manager
  # Verify installation
  ```

- [ ] **WebKit Library Fix**
  ```bash
  # Follow distribution-specific WebKit instructions
  # Test library linking
  ldconfig -p | grep webkit
  ```

- [ ] **Installation Script**
  ```bash
  curl -fsSL https://raw.githubusercontent.com/libr-forum/Libr/main/scripts/install-libr.sh | bash
  ```

- [ ] **Verification**
  ```bash
  which libr
  libr --version
  # Test GUI launch if possible
  ```

## 🔄 Automated Testing Pipeline

### GitHub Actions (Recommended)
Create `.github/workflows/test-installation.yml`:

```yaml
name: Test Installation

on: [push, pull_request]

jobs:
  test-installation:
    strategy:
      matrix:
        os: 
          - ubuntu:22.04
          - ubuntu:24.04
          - fedora:38
          - archlinux:latest
    
    runs-on: ubuntu-latest
    container: ${{ matrix.os }}
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Install dependencies
        run: |
          if command -v apt >/dev/null; then
            apt update && apt install -y wget curl gpg
          elif command -v dnf >/dev/null; then
            dnf update -y && dnf install -y wget curl gnupg2
          elif command -v pacman >/dev/null; then
            pacman -Sy --noconfirm wget curl gnupg
          fi
      
      - name: Test README instructions
        run: ./scripts/test-readme-validation.sh
```

## 📊 Test Results Documentation

### Create Test Report Template:
```markdown
# Installation Test Results

## Test Environment
- **Date**: YYYY-MM-DD
- **Libr Version**: v1.0.0-beta
- **Tested By**: Your Name

## Results Summary

| Distribution | APT Repo | Direct Package | WebKit Fix | Install Script | Status |
|-------------|----------|----------------|------------|----------------|---------|
| Ubuntu 22.04 | ✅ | ✅ | N/A | ✅ | PASS |
| Ubuntu 24.04 | ✅ | ✅ | ✅ | ✅ | PASS |
| Fedora 38 | N/A | ✅ | ✅ | ✅ | PASS |
| Arch Linux | N/A | ✅ | ✅ | ✅ | PASS |

## Issues Found
- List any problems discovered
- Include error messages
- Suggest fixes

## Recommendations
- Documentation updates needed
- Script improvements
- Package fixes required
```

## 🚨 Common Issues to Test

1. **Package Dependencies**: Ensure all required libraries are installed
2. **Architecture Naming**: Verify amd64 vs x86_64 consistency
3. **WebKit Compatibility**: Test library linking on newer distributions  
4. **Path Issues**: Check if binary is accessible after installation
5. **Permission Problems**: Test sudo requirements
6. **Network Issues**: Handle download failures gracefully

## ✅ Final Validation

Before marking as tested:
- [ ] All URLs return 200 status
- [ ] Packages install without errors
- [ ] Application launches successfully
- [ ] Desktop integration works
- [ ] Uninstallation works cleanly
- [ ] Documentation is accurate
