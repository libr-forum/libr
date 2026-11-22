#!/bin/bash
# LIBR quick setup: checks the toolchain for each part of the repo and installs the
# npm dependencies of the parts whose tools are present. Safe to run repeatedly.
set -u
cd "$(dirname "$0")/.."

have() { command -v "$1" >/dev/null 2>&1; }
ok() { echo "✅ $*"; }
miss() { echo "❌ $*"; }

echo "🚀 LIBR setup"
echo

have git && ok "git" || { miss "git: https://git-scm.com/downloads"; exit 1; }

if have go; then
  ok "$(go version) (go.work wants 1.24.4; GOTOOLCHAIN=auto fetches it)"
else
  miss "Go 1.24.4: https://go.dev/dl/ (core, DB node, relay, desktop, mobile bridge)"
fi

if have node; then
  major=$(node --version | sed 's/v\([0-9]*\).*/\1/')
  [ "$major" -ge 22 ] && ok "node $(node --version)" || miss "node $(node --version): Node 22 is used in CI"
else
  miss "Node 22: https://nodejs.org/ (desktop frontend, mobile app, seed server)"
fi

have docker && ok "docker (testbed/e2e.sh)" || miss "Docker: https://docs.docker.com/get-docker/ (only for testbed/)"
have wails && ok "wails (desktop app)" || echo "ℹ️  wails not installed: only needed for the desktop app (https://wails.io)"
[ -n "${ANDROID_HOME:-}" ] && ok "Android SDK at $ANDROID_HOME" || echo "ℹ️  ANDROID_HOME not set: only needed for the Android app"

echo
if have go; then
  echo "📦 Go modules"
  for m in core/libr core/db core/mod_client mobile-app/go-core services/relay; do
    (cd "$m" && go mod download) && ok "$m" || miss "$m: go mod download failed"
  done
fi
if have npm; then
  echo "📦 npm packages"
  for d in core/mod_client/frontend mobile-app services/librserver; do
    (cd "$d" && npm ci --silent) && ok "$d" || miss "$d: npm ci failed"
  done
fi

cat <<'EOF'

🎉 Done. Next:
  • Read docs/BEGINNER_GUIDE.md
  • Run the tests for your area (CONTRIBUTING.md#testing)
  • Try a local network: DEVNET_IP=<your ip> testbed/devnet.sh start
  • Check a commit message: ./scripts/validate-commit.sh "feat: your message"
EOF
