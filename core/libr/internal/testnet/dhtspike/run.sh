#!/usr/bin/env bash
# DHT spike in Docker. Phase 1: full topology. Phase 2: stop relay-a and check that
# a fresh peer joining via relay-b still routes and finds mods.
set -euo pipefail
cd "$(dirname "$0")"
ROOT=../../..   # core/libr
P="docker compose -p libr-dhtspike -f compose.yml"
trap '$P --profile probe down -v --remove-orphans >/dev/null 2>&1' EXIT

# SKIP_BUILD=1 reuses an existing image (e.g. when the registry is unreachable)
[ -n "${SKIP_BUILD:-}" ] || docker build -q -t libr-dhtspike -f Dockerfile "$ROOT" >/dev/null
$P up -d relay-a relay-b
sleep 2
$P up -d db1
sleep 2
$P up -d
sleep 40

echo "== phase 1: two relays"
$P --profile probe run --rm probe

echo "== phase 2: relay-a stopped"
$P stop relay-a
sleep 5
PROBE_RELAYS=relay-b $P --profile probe run --rm probe
