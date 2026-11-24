#!/usr/bin/env bash
# LIBR v2 dev network on one machine's host network, reachable by desktop and mobile
# apps on the LAN: a relay, DB nodes (one public, the rest NAT-style behind the relay)
# and librctl moderators for community "general".
#
#   DEVNET_IP=192.168.1.20 testbed/devnet.sh start   # prints the env for apps
#   testbed/devnet.sh stop
#
# EXTRA_MODS=<base64 pubkey>,... adds app identities to the moderator set (fewer
# librctl mods are started so there are always 4 moderators, f=1).
set -euo pipefail
ROOT=$(cd "$(dirname "$0")/.." && pwd)
DIR=${DEVNET_DIR:-$HOME/libr-devnet}
cmd=${1:-start}

stop() {
  [ -d "$DIR/pids" ] || return 0
  for f in "$DIR"/pids/*; do [ -e "$f" ] && kill "$(cat "$f")" 2>/dev/null || true; done
  rm -rf "$DIR/pids"
}
if [ "$cmd" = stop ]; then stop; echo stopped; exit 0; fi

IP=${DEVNET_IP:?set DEVNET_IP to this host address as apps see it}
stop
rm -rf "$DIR" && mkdir -p "$DIR"/{bin,keys,communities,logs,pids,data}
(cd "$ROOT/core/libr" && go build -o "$DIR/bin/librctl" ./cmd/librctl)
(cd "$ROOT/core/db" && go build -o "$DIR/bin/dbnode" .)
(cd "$ROOT/services/relay" && go build -o "$DIR/bin/relay" .)
ctl="$DIR/bin/librctl"

for n in founder relay db1 db2 db3 db4 mod1 mod2 mod3 mod4; do
  "$ctl" keygen -out "$DIR/keys/$n.key" -libp2p > "$DIR/keys/$n.id"
done
nid() { cut -d' ' -f1 "$DIR/keys/$1.id"; }
pub() { cut -d' ' -f2 "$DIR/keys/$1.id"; }

IFS=, read -r -a extra <<<"${EXTRA_MODS:-}"
mods=("${extra[@]}")
nctl=$((4 - ${#extra[@]}))
for i in $(seq 1 $nctl); do mods+=("$(pub mod$i)"); done
"$ctl" genesis -community general -f 1 -founders "$DIR/keys/founder.key" \
  -mods "$(IFS=,; echo "${mods[*]}")" > "$DIR/communities/general.json"

RELAY=/ip4/$IP/tcp/4301/p2p/$(nid relay)
BOOT=/ip4/$IP/tcp/4101/p2p/$(nid db1)
run() { local name=$1; shift; nohup "$@" > "$DIR/logs/$name.log" 2>&1 & echo $! > "$DIR/pids/$name"; }

LIBP2P_PRIV_KEY=$(cut -d' ' -f3 "$DIR/keys/relay.id") LIBR_LISTEN=/ip4/0.0.0.0/tcp/4301 \
  RELAY_ANNOUNCE=/ip4/$IP/tcp/4301 HEALTH_PORT=4080 run relay "$DIR/bin/relay"
sleep 1
for i in 1 2 3 4; do
  mkdir -p "$DIR/data/db$i" && cp "$DIR/keys/db$i.key" "$DIR/data/db$i/node.key"
  private=1; [ $i = 1 ] && private=0 # db1 is directly reachable and is the bootstrap peer
  LIBR_DATA_DIR=$DIR/data/db$i LIBR_COMMUNITIES=$DIR/communities LIBR_LISTEN=/ip4/0.0.0.0/tcp/410$i \
    LIBR_RELAYS=$RELAY LIBR_BOOTSTRAP=$BOOT LIBR_PRIVATE=$private run db$i "$DIR/bin/dbnode"
done
sleep 3
for i in $(seq 1 $nctl); do
  run mod$i "$ctl" mod serve -key "$DIR/keys/mod$i.key" -votes "$DIR/mod$i.votes" -communities "$DIR/communities" \
    -relays "$RELAY" -bootstrap "$BOOT" -listen /ip4/0.0.0.0/tcp/420$i -report-policy keyword:remove-me
done
sleep 5

cat > "$DIR/env" <<ENV
LIBR_RELAYS=$RELAY
LIBR_BOOTSTRAP=$BOOT
LIBR_COMMUNITIES=$DIR/communities
ENV
echo "devnet up in $DIR (logs in $DIR/logs). App environment:"
cat "$DIR/env"
echo "Copy $DIR/communities/general.json to the app's communities dir (or LIBR_COMMUNITIES)."
