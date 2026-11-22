# LIBR — Claude notes

Decentralized, moderated forum: libp2p + Kademlia DHT storage + multi-moderator quorum signing (Ed25519).

## Working mode
- Use the **ponytail** skill (full) by default: smallest complete change, no speculative abstractions.
- Codebase questions: check `graphify-out/GRAPH_REPORT.md` / `graphify query` first (see graphify section below).

## Layout (each dir is its own module; no root build)
| Path | What | Stack |
|---|---|---|
| `services/librserver` | Discovery/registry + Ed25519 challenge auth | Node, Express, MongoDB (`npm start`) |
| `services/relay` | Stock libp2p circuit relay v2 (forwards encrypted bytes only) | Go (`relay.go`, uses `core/libr/p2p`) |
| `core/db` | v2 DB node binary + SQLite store (`sqlstore/`); logic in `core/libr/dbnode` (see `core/db/db_doc.md`) | Go, SQLite (`docker-compose.yml`, `core/db/Dockerfile`) |
| `core/libr` | **Shared Go core** (desktop + mobile). v2: `proto` (signed objects), `p2p` (hosts, NAT reservations, DHT), `wire` (framing, rate limits, protocol IDs), `dbnode` (storage + `/libr/2/remove` tombstones keyed by MsgID, content deleted), `moderator` (`/libr/2/vote`, `/libr/2/report`), `forum` (Submit/Feed/Get/Report/Delete client; Feed hides messages with a verified tombstone from the bucket's index nodes), `cmd/librctl` (headless CLI: keygen, genesis, mod serve, client submit/feed/report/delete). v1 still used by the apps until step 6: `client` (submit/fetch/report/cron), `peers`/`network` (libp2p), `mod/*` (moderator handlers, LLM pipeline, mod SQLite), `crypto`, `types`, `cache`, `identity`, `keycache` | Go |
| `core/mod_client` | Desktop app: thin Wails adapter (`app.go`) over `core/libr` | Go + Wails v2, React/Vite/shadcn frontend |
| `mobile-app` | Expo / React Native client | TS; native module `modules/LibrCore.ts` → `android/.../LibrCoreModule` |
| `mobile-app/go-core` | gomobile bridge only (`bridge/` = exported API) over `core/libr` | Go |
| `landing-page` | Marketing site | React/Vite |
| `packaging`, `Casks`, `scripts` | nfpm deb/rpm/arch, Homebrew, install/release scripts | |

## Gotchas
- Desktop and mobile share `core/libr`; there is no fork any more. UI hooks go through `client.SetEvents` (Wails emits on desktop, notifications on mobile) and `peers.HandlePost` (desktop registers moderator routes; mobile forwards unknown routes to native). Never import Wails or cgo SQLite from packages the bridge uses (`go list -deps ./bridge` in go-core to check).
- The bridge API is pinned: `mobile-app/go-core/bridge/testdata/bridge_api.golden` plus a test that every `callBridgeString` in `LibrCoreModule.kt` resolves (Kotlin calls by reflection; mismatches fail silently). Desktop: `app_api_test.go` checks `App` methods = `frontend/wailsjs/go/main/App.d.ts`.
- go-core needs a `replace` to `core/libr` because gomobile ignores go.work; build with `GOWORK=off` (`build-android.sh`).
- `.wailsignore` excludes everything but `core/mod_client` from Wails dev/build.
- Root `go.work` covers core/db, core/libr, core/mod_client, mobile-app/go-core, services/relay. Tests: `go test -race ./...` per module; frontend `npm test`/`npm run typecheck`/`npm run lint`; librserver `npm test`. CI = `.github/workflows/ci.yml`. Lint config `.golangci.yml` only flags code changed since `origin/staging`.
- `core/mod_client` embeds `frontend/dist`; create `frontend/dist/.keep` (or build the frontend) before `go vet`/`go test`.
- `CONTRIBUTING.md` / `docs/BEGINNER_GUIDE.md` still describe a stale `src/` layout; BEGINNER_GUIDE has unresolved merge-conflict markers (fixed in v2 step 10).
- v2 rebuild in progress on branch `v2`. New shared code lives in `core/libr`; `core/libr/internal/testnet` builds in-process relayed topologies (all NATed peers reachable only via `/p2p-circuit`). Streams over relays need `network.WithAllowLimitedConn`; LIBR hosts use `testnet.ResourceManager()` because libp2p's default conn-rate limit charges relayed conns to the relay IP. Docker e2e: `testbed/e2e.sh` (full v2 network: relays, public + NAT-only DB nodes, 4 mods, 2 clients on isolated networks) and `core/libr/internal/testnet/dhtspike/run.sh`; both in workflow `e2e.yml`, `SKIP_BUILD=1` reuses images. NATed hosts use `p2p.Config{Private: true}` (own reservations, circuit addrs only; AutoRelay ignores relays without public IPs).
- Branches: work on `staging`, PRs to `main`.

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
- A committed graphify-out/graph.json is rewritten whole on every build, so a large diff to it in a pull request is expected. Do not flag it as churn or suggest skipping the graph rebuild or its hooks.
