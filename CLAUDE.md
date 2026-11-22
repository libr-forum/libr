# LIBR — Claude notes

Decentralized, moderated forum: libp2p + Kademlia DHT storage + multi-moderator quorum signing (Ed25519).

## Working mode
- Use the **ponytail** skill (full) by default: smallest complete change, no speculative abstractions.
- Codebase questions: check `graphify-out/GRAPH_REPORT.md` / `graphify query` first (see graphify section below).

## Layout (each dir is its own module; no root build)
| Path | What | Stack |
|---|---|---|
| `services/librserver` | Discovery/registry + Ed25519 challenge auth | Node, Express, MongoDB (`npm start`) |
| `services/relay` | libp2p Circuit Relay v2 | Go (`relay.go`) |
| `core/db` | DHT storage node (MsgCert/ReportCert, Kademlia routing) | Go, SQLite (`docker-compose.yml`, `core/db/Dockerfile`) |
| `core/crypto` | Key gen / sign / verify | Go |
| `core/mod_client` | Desktop user + moderator app, LLM auto-mod | Go + Wails v2, React/Vite/shadcn frontend |
| `mobile-app` | Expo / React Native client | TS; native module `modules/LibrCore.ts` → `android/.../LibrCoreModule` |
| `mobile-app/go-core` | gomobile build of the client core (`bridge/` = exported API) | Go |
| `landing-page` | Marketing site | React/Vite |
| `packaging`, `Casks`, `scripts` | nfpm deb/rpm/arch, Homebrew, install/release scripts | |

## Gotchas
- **`mobile-app/go-core` is a forked copy of `core/mod_client`** (same module path `github.com/libr-forum/Libr/core/mod_client`) and `go-core/crypto` copies `core/crypto`. ~18 files have drifted. A protocol fix in one usually needs porting to the other — diff before editing.
- Message signing depends on canonicalization (`util` / `CanonicalizeMsg`); desktop and mobile must stay byte-identical or signatures fail across clients.
- `.wailsignore` excludes everything but `core/mod_client` from Wails dev/build.
- Root `go.work` covers core/crypto, core/db, core/libr, core/mod_client, services/relay (`mobile-app/go-core` excluded until v2 step 3; use `GOWORK=off` there). Tests: `go test -race ./...` per module; frontend `npm test`/`npm run typecheck`/`npm run lint`; librserver `npm test`. CI = `.github/workflows/ci.yml`. Lint config `.golangci.yml` only flags code changed since `origin/staging`.
- `core/mod_client` embeds `frontend/dist`; create `frontend/dist/.keep` (or build the frontend) before `go vet`/`go test`.
- `CONTRIBUTING.md` / `docs/BEGINNER_GUIDE.md` still describe a stale `src/` layout; BEGINNER_GUIDE has unresolved merge-conflict markers (fixed in v2 step 10).
- v2 rebuild in progress on branch `v2`. New shared code lives in `core/libr`; `core/libr/internal/testnet` builds in-process relayed topologies (all NATed peers reachable only via `/p2p-circuit`). Streams over relays need `network.WithAllowLimitedConn`; LIBR hosts use `testnet.ResourceManager()` because libp2p's default conn-rate limit charges relayed conns to the relay IP. Docker e2e: `core/libr/internal/testnet/dhtspike/run.sh` (workflow `e2e.yml`).
- Branches: work on `staging`, PRs to `main`.

## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).
- A committed graphify-out/graph.json is rewritten whole on every build, so a large diff to it in a pull request is expected. Do not flag it as churn or suggest skipping the graph rebuild or its hooks.
