#!/usr/bin/env bash
# LIBR cross-distro installer with verbose debugging
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/libr-forum/Libr/main/scripts/install-libr.sh | bash
#   curl -fsSL .../install-libr.sh | DEBUG=1 bash -s -- v1.0.0-beta   # install specific version with debug
set -Eeuo pipefail

DEBUG="${DEBUG:-1}"   # 1/true = verbose with xtrace; 0/false = quieter
if [[ "$DEBUG" =~ ^(1|true|yes)$ ]]; then set -x; fi

info()  { echo -e "[INFO]  $*"; }
debug() { [[ "$DEBUG" =~ ^(1|true|yes)$ ]] && echo -e "[DEBUG] $*"; }
warn()  { echo -e "[WARN]  $*" >&2; }
err()   { echo -e "[ERROR] $*" >&2; }

# --- Detect distribution & architecture ---
if [[ -f /etc/os-release ]]; then
  # shellcheck disable=SC1091
  source /etc/os-release
  DISTRO="${ID:-unknown}"
  DISTRO_VER="${VERSION_ID:-unknown}"
  CODENAME="${VERSION_CODENAME:-}"
else
  err "Cannot detect distribution (missing /etc/os-release)."
  exit 1
fi

# Debian/Ubuntu: dpkg prints 'amd64', else fall back to uname -m
ARCH="$(dpkg --print-architecture 2>/dev/null || uname -m)"
DEB_ARCH="$ARCH"
# Map to RPM arch naming
case "$ARCH" in
  amd64) RPM_ARCH="x86_64" ;;
  arm64|aarch64) RPM_ARCH="aarch64" ;;
  *) RPM_ARCH="$ARCH" ;;
esac
# Arch Linux typically uses uname -m like x86_64/aarch64
PAC_ARCH="$(uname -m)"

info "Detected: DISTRO=$DISTRO $DISTRO_VER $CODENAME  ARCH=$ARCH  RPM_ARCH=$RPM_ARCH  PAC_ARCH=$PAC_ARCH"

# --- Fetch latest or use provided version (tag from GitHub) ---
REPO="libr-forum/libr"
LATEST_VERSION="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | awk -F'"' '/tag_name/{print $4; exit}')"
INPUT_VERSION="${1:-}"
VERSION="${INPUT_VERSION:-$LATEST_VERSION}"

if [[ -z "${VERSION:-}" ]]; then
  err "Could not determine version (GitHub API returned no tag)."
  exit 1
fi

# Useful normalized variants for logging/diagnostics
VERSION_NO_V="${VERSION#v}"                            # v1.0.0-beta -> 1.0.0-beta
DEB_VERSION="$(printf '%s' "$VERSION_NO_V" | sed 's/-/~/g')"   # 1.0.0-beta -> 1.0.0~beta
debug "Version forms: RAW_TAG=$VERSION  NO_V=$VERSION_NO_V  DEB_VERSION=$DEB_VERSION"

# --- Determine installed version via package manager FIRST (avoids launching UI) ---
INSTALLED_VERSION="none"
case "$DISTRO" in
  ubuntu|debian)
    INSTALLED_VERSION="$(dpkg-query -W -f='${Version}\n' libr 2>/dev/null || true)"
    ;;
  fedora|rhel|centos)
    # %EVR is epoch:version-release; we’ll log it; comparison is just a best-effort here
    INSTALLED_VERSION="$(rpm -q --qf '%{EVR}\n' libr 2>/dev/null || true)"
    ;;
  arch)
    INSTALLED_VERSION="$(pacman -Q libr 2>/dev/null | awk '{print $2}' || true)"
    ;;
  *)
    warn "Unsupported distro for package-query; will try 'libr --version' fallback."
    ;;
esac

# Fallback: try not to launch UI; if 'libr --version' prints, capture it
if [[ -z "$INSTALLED_VERSION" || "$INSTALLED_VERSION" == "none" ]]; then
  if command -v libr >/dev/null 2>&1; then
    # Expect lines like: "libr 1.0.0" or "libr v1.0.0-beta"
    INSTALLED_VERSION="$(libr --version 2>/dev/null | awk '{print $2}')"
  fi
fi

INSTALLED_VERSION="${INSTALLED_VERSION:-none}"
info "Installed version detected: '$INSTALLED_VERSION'"

# --- Decide if we need to install ---
# We consider a match if any of these match (helps across packaging/tag formats):
match_versions() {
  local inst="$1" tag="$2" nov="$3" debv="$4"
  [[ -z "$inst" || "$inst" == "none" ]] && return 1
  [[ "$inst" == "$tag" ]]  && return 0
  [[ "$inst" == "$nov"  ]] && return 0
  [[ "$inst" == "$debv" ]] && return 0
  return 1
}

if match_versions "$INSTALLED_VERSION" "$VERSION" "$VERSION_NO_V" "$DEB_VERSION"; then
  info "libr is already at the requested version (match found). Skipping reinstall."
  exit 0
fi

if [[ "$INSTALLED_VERSION" != "none" ]]; then
  info "Removing existing libr ($INSTALLED_VERSION) before install/update…"
  case "$DISTRO" in
    ubuntu|debian) sudo apt-get remove -y libr || true ;;
    fedora|rhel|centos) sudo dnf remove -y libr || sudo yum remove -y libr || true ;;
    arch) sudo pacman -Rns --noconfirm libr || true ;;
    *) warn "Skip removal on unsupported distro '$DISTRO'."; ;;
  esac
fi

info "Proceeding to install libr $VERSION on $DISTRO ($ARCH)…"

# --- Helpers: probe URL & download with status ---
probe_url() {
  local url="$1"
  local code
  code="$(curl -s -o /dev/null -w '%{http_code}' "$url" || echo 000)"
  debug "HTTP probe $url -> $code"
  printf '%s' "$code"
}

download_file() {
  local url="$1" out="$2"
  info "Downloading: $url"
  local code; code="$(probe_url "$url")"
  if [[ "$code" != "200" ]]; then
    err "URL not reachable (HTTP $code): $url"
    return 1
  fi
  curl -fSL "$url" -o "$out"
}

# --- Optional: Only apply WebKitGTK shim if the old SONAME is missing (Ubuntu/Debian only) ---
maybe_fix_webkit_ubuntu() {
  # Only on Debian/Ubuntu & only on amd64 path below
  [[ "$DISTRO" != "ubuntu" && "$DISTRO" != "debian" ]] && { debug "WebKit fix: not Debian/Ubuntu, skip."; return 0; }
  local libdir="/usr/lib/x86_64-linux-gnu"
  if [[ "$DEB_ARCH" != "amd64" || ! -d "$libdir" ]]; then
    debug "WebKit fix: non-amd64 or libdir missing; skip."
    return 0
  fi

  local need_fix=0
  [[ ! -e "$libdir/libwebkit2gtk-4.0.so.37" ]] && need_fix=1
  [[ ! -e "$libdir/libjavascriptcoregtk-4.0.so.18" ]] && need_fix=1

  if [[ "$need_fix" -eq 0 ]]; then
    debug "WebKit fix: 4.0 SONAMEs already present; nothing to do."
    return 0
  fi

  info "Applying WebKitGTK compatibility shim (quiet)…"
  sudo apt-get update -y >/dev/null 2>&1 || true
  sudo apt-get install -y libwebkit2gtk-4.1-0 libjavascriptcoregtk-4.1-0 >/dev/null 2>&1 || true
  sudo ln -sf "$libdir/libwebkit2gtk-4.1.so.0"        "$libdir/libwebkit2gtk-4.0.so.37" || true
  sudo ln -sf "$libdir/libjavascriptcoregtk-4.1.so.0" "$libdir/libjavascriptcoregtk-4.0.so.18" || true
  info "WebKitGTK shim applied."
}

# --- Install per distro (with DEBUG-friendly logging) ---
case "$DISTRO" in
  ubuntu|debian)
    # NOTE: Your published .deb name convention in prior messages: libr_<DEB_VERSION>_<DEB_ARCH>.deb
    DEB_NAME="libr_${DEB_VERSION}_${DEB_ARCH}.deb"
    URL="https://github.com/${REPO}/releases/download/${VERSION}/${DEB_NAME}"
    info  "DEB expected filename: $DEB_NAME"
    debug "DEB URL: $URL"
    download_file "$URL" "libr.deb" || { err "Failed to fetch $URL"; exit 1; }

    info "Installing .deb…"
    if ! sudo dpkg -i libr.deb; then
      warn "dpkg reported missing deps; attempting 'apt-get -f install'…"
      sudo apt-get install -f -y
    fi
    rm -f libr.deb

    maybe_fix_webkit_ubuntu
    ;;

  fedora|rhel|centos)
    # Your script expects: libr-${VERSION}.${RPM_ARCH}.rpm  (includes leading 'v' if present)
    RPM_NAME="libr-${VERSION}.${RPM_ARCH}.rpm"
    URL="https://github.com/${REPO}/releases/download/${VERSION}/${RPM_NAME}"
    info  "RPM expected filename: $RPM_NAME"
    debug "RPM URL: $URL"
    download_file "$URL" "libr.rpm" || { err "Failed to fetch $URL"; exit 1; }

    info "Installing .rpm…"
    if command -v dnf >/dev/null 2>&1; then
      sudo dnf install -y ./libr.rpm
    else
      sudo yum install -y ./libr.rpm
    fi
    rm -f libr.rpm
    ;;

  arch)
    # Your script expects: libr-${VERSION}-${PAC_ARCH}.pkg.tar.zst
    PAC_NAME="libr-${VERSION}-${PAC_ARCH}.pkg.tar.zst"
    URL="https://github.com/${REPO}/releases/download/${VERSION}/${PAC_NAME}"
    info  "Arch expected filename: $PAC_NAME"
    debug "Arch URL: $URL"
    download_file "$URL" "libr.pkg.tar.zst" || { err "Failed to fetch $URL"; exit 1; }

    info "Installing Arch package…"
    sudo pacman -U --noconfirm libr.pkg.tar.zst
    rm -f libr.pkg.tar.zst
    ;;

  *)
    err "Unsupported distribution: $DISTRO"
    exit 1
    ;;
esac

# --- Final verification ---
POST_VERSION="unknown"
case "$DISTRO" in
  ubuntu|debian) POST_VERSION="$(dpkg-query -W -f='${Version}\n' libr 2>/dev/null || true)" ;;
  fedora|rhel|centos) POST_VERSION="$(rpm -q --qf '%{EVR}\n' libr 2>/dev/null || true)" ;;
  arch) POST_VERSION="$(pacman -Q libr 2>/dev/null | awk '{print $2}' || true)" ;;
esac
info "Installed (post-check) version: '${POST_VERSION:-unknown}'"
echo "✅ Installation finished."
