# Installation Guide

This document explains how to install and run **LIBR** on Windows, Linux, and macOS.

---

## 🪟 Windows

1. Download the latest **Windows release** (`libr-win-amd64.exe`) from the [Releases](../../releases) page.  
2. Double-click to run it.  
   - If you face issues (e.g., the app doesn’t start), try **right-click → Run as administrator**.  

---

## 🐧 Linux

### Ubuntu 22.04 and below
1. Download the **Linux build** (`libr-linux-amd64`) from [Releases](../../releases).  
2. Make it executable:
   ```bash
   chmod +x ./libr-linux-amd64
   ```
3. Run it:
   ```bash
   ./libr-linux-amd64
   ```

---

### Ubuntu 24.04 (Noble) and newer
On newer Ubuntu versions, you may encounter missing library errors like:

```
./libr-linux-amd64: error while loading shared libraries: libwebkit2gtk-4.0.so.37: cannot open shared object file: No such file or directory
```

or

```
./libr-linux-amd64: error while loading shared libraries: libjavascriptcoregtk-4.0.so.18: cannot open shared object file: No such file or directory
```

👉 To fix this, install the updated WebKitGTK libraries and create symlinks:

```bash
# Update package index
sudo apt update

# Install newer WebKitGTK packages
sudo apt install -y libwebkit2gtk-4.1-0 libjavascriptcoregtk-4.1-0

# Create symlinks so the binary finds the expected names
sudo ln -sf /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.1.so.0 \
            /usr/lib/x86_64-linux-gnu/libwebkit2gtk-4.0.so.37

sudo ln -sf /usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.1.so.0 \
            /usr/lib/x86_64-linux-gnu/libjavascriptcoregtk-4.0.so.18

# Make the binary executable
chmod +x ./libr-linux-amd64

# Run it
./libr-linux-amd64
```

---

## 🍎 macOS

1. Download the **macOS release** (`libr-darwin-amd64.out`) from [Releases](../../releases).  
2. On first run, macOS may block the app. To fix this:
   - Go to **System Settings → Privacy & Security**.  
   - Allow the app under the “Security” section.  
3. Run the binary normally.  

👉 If the error still persists:
```bash
chmod +x ./libr-darwin-amd64.out
./libr-darwin-amd64.out
```

---

## 📩 Feedback & Support
If you encounter any issues during installation, please let us know here: [Feedback Form](https://docs.google.com/forms/d/e/1FAIpQLSdOnq6uPpLYEQIueuHtvydMI8q1CMHC_TJzDkUDUU8UCGo4ew/viewform)

---

✅ You should now have LIBR running on your system!
