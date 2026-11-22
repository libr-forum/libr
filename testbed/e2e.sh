#!/usr/bin/env bash
# LIBR v2 end-to-end test on the Docker testbed (compose.yml): post from one isolated
# network, read from the other, then take moderators, a DB node and a relay down.
# SKIP_BUILD=1 reuses existing images. KEEP=1 leaves the testbed running.
set -euo pipefail
cd "$(dirname "$0")"
ROOT=..
OUT=out
P="docker compose -p libr-testbed -f compose.yml --env-file $OUT/env"

if [ -z "${SKIP_BUILD:-}" ]; then
  docker build -q -t libr-ctl -f $ROOT/core/libr/cmd/librctl/Dockerfile $ROOT/core/libr >/dev/null
  docker build -q -t libr-db -f $ROOT/core/db/Dockerfile $ROOT >/dev/null
  docker build -q -t libr-relay -f $ROOT/services/relay/Dockerfile $ROOT >/dev/null
fi

# Fresh keys and genesis every run (nothing secret is committed).
rm -rf $OUT && mkdir -p $OUT/keys $OUT/communities $OUT/trace
ctl() { docker run --rm -u "$(id -u):$(id -g)" -v "$PWD/$OUT:/out" libr-ctl "$@"; }
for n in founder author relay-a relay-b db1 db2 db3 db4 db5 mod1 mod2 mod3 mod4; do
  ctl keygen -out /out/keys/$n.key -libp2p > $OUT/keys/$n.id
done
pid() { cut -d' ' -f1 $OUT/keys/$1.id; }
pub() { cut -d' ' -f2 $OUT/keys/$1.id; }
ctl genesis -community general -f 1 -founders /out/keys/founder.key \
  -mods "$(pub mod1),$(pub mod2),$(pub mod3),$(pub mod4)" > $OUT/communities/general.json
cat > $OUT/env <<ENV
HOST_USER=$(id -u):$(id -g)
RELAY_A_KEY=$(cut -d' ' -f3 $OUT/keys/relay-a.id)
RELAY_B_KEY=$(cut -d' ' -f3 $OUT/keys/relay-b.id)
RELAYS=/dns4/relay-a/tcp/4001/p2p/$(pid relay-a),/dns4/relay-b/tcp/4001/p2p/$(pid relay-b)
BOOT=/dns4/db1/tcp/4001/p2p/$(pid db1),/dns4/db2/tcp/4001/p2p/$(pid db2)
ENV

[ -n "${KEEP:-}" ] || trap '$P --profile client down -v --remove-orphans >/dev/null 2>&1' EXIT
$P up -d relay-a relay-b db1 db2
sleep 3
$P up -d
echo "waiting for the DHT to settle"
sleep 30

fail() { echo "FAIL: $*"; $P logs --tail 30 >&2 || true; exit 1; }
submit() { # client text [extra flags]
  $P run --rm -T "$1" client submit -community general -text "$2" -trace /out/trace/$1.jsonl "${@:3}"
}
hash_of() { sed -n 's/.*hash=\([0-9a-f]*\).*/\1/p' <<<"$1"; }
feed() { $P run --rm -T "$1" client feed -community general -since 1h; }
expect_post() { # client text
  local got; got=$(feed "$1") || fail "feed from $1: $got"
  grep -qF "$2" <<<"$got" || fail "$1 feed lacks '$2': $got"
}
expect_gone() { # client text
  local got; got=$(feed "$1") || fail "feed from $1: $got"
  if grep -qF "$2" <<<"$got"; then fail "$1 feed still shows '$2'"; fi
}

echo "== 1. post from nat1, read from nat2"
submit client1 "hello from nat1" || fail "submit"
expect_post client2 "hello from nat1"

echo "== 2. one mod rejects by policy; the other three still approve"
submit client2 "this is forbidden by mod4" || fail "3 of 4 approvals should pass"

echo "== 3. one mod down (f=1): still posts"
$P stop mod1
submit client1 "one mod down" || fail "submit with one mod down"

echo "== 4. two mods down: no quorum, clean error"
$P stop mod2
if out=$(submit client1 "two mods down" 2>&1); then fail "posted without quorum: $out"; fi
grep -q "not enough moderators\|rejected by moderators" <<<"$out" || fail "unexpected error: $out"
$P start mod1 mod2
sleep 15

echo "== 5. NAT-only DB node down: receipt quorum still holds"
$P stop db3
submit client2 "one db down" || fail "submit with db3 down"
expect_post client1 "one db down"

echo "== 6. relay-a down: everything moves to relay-b"
$P stop relay-a
sleep 20 # one reservation keeper tick withdraws relay-a circuit addresses
submit client1 "after relay failover" || fail "submit after relay-a down"
expect_post client2 "after relay failover"
expect_post client2 "hello from nat1"

echo "== 7. author deletes their post (from the other network)"
out=$(submit client1 "author will delete this" -key /out/keys/author.key) || fail "submit: $out"
h=$(hash_of "$out")
expect_post client2 "author will delete this"
$P run --rm -T client2 client delete -hash "$h" -key /out/keys/author.key || fail "delete"
expect_gone client1 "author will delete this"
if $P run --rm -T client1 client delete -hash "$(hash_of "$(submit client1 "not yours")")" 2>/dev/null; then
  fail "deleted a post with a different key"
fi

echo "== 8. report: mods remove by quorum, or keep"
out=$(submit client1 "please remove-me") || fail "submit: $out"
$P run --rm -T client2 client report -hash "$(hash_of "$out")" -reason "breaks rules" || fail "report"
expect_gone client1 "please remove-me"
out=$(submit client1 "a fine post") || fail "submit: $out"
if $P run --rm -T client2 client report -hash "$(hash_of "$out")" -reason "dislike" 2>/dev/null; then
  fail "fine post removed"
fi
expect_post client2 "a fine post"

echo "== trace"
for op in vote store index_put; do
  grep -q "\"op\":\"$op\"" $OUT/trace/client1.jsonl || fail "no $op in trace"
done
wc -l $OUT/trace/*.jsonl
echo "E2E PASS"
