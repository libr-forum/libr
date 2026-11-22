# LIBR v2: abuse limits and known gaps

Audit of every v2 network entry point (step 8). Numbers are the constants in the code;
change them there and here together.

## Transport (all peers)

| Limit | Value | Where |
|---|---|---|
| Frame size, checked before the body is read | 2 MiB | `wire.MaxFrame` |
| Whole request/response exchange | 30 s | `wire.Timeout` |
| Rate limit, checked before reading the request | per peer token bucket (below) | `wire.Handle` |
| Limiter memory | 10 000 keys, then reset | `wire.maxTracked` |
| New connections per IP (relayed conns count against the relay IP) | 20/s, burst 200 | `p2p.ResourceManager` |
| Streams/conns/memory per peer | libp2p default resource manager (autoscaled) | `p2p.ResourceManager` |
| Encryption and peer authentication | Noise; a dialer pins the peer ID it expects | `p2p.BaseOptions` |

## Relays

| Limit | Value | Where |
|---|---|---|
| Circuit lifetime / data | 10 min / 8 MiB, then reset | `p2p.RelayResources` |
| Concurrent circuits | 64 | same |
| Reservations per IP / ASN | libp2p defaults | same |
| Application data | none: relays forward encrypted bytes only, no `/chat/1.0.0` | `services/relay` |

## DB nodes (`/libr/2/store`, `get`, `index/*`, `remove`)

| Check | Where |
|---|---|
| 20 req/s per peer, burst 100 | `dbnode.ratePerSecond/rateBurst` |
| Store and index only certificates that verify against a known ModSet (fail closed) | `dbnode.admit` |
| Content 1..1 MiB, timestamp at most 5 min in the future | `proto.VerifyMsgCert` |
| Removed messages: new certificates of them are refused | `dbnode.admit` |
| Opted-out communities refused (`LIBR_OPTOUT`), clients move to the next node | `dbnode.Node.OptOut` |
| index/get returns at most 1000 hashes and 1000 tombstones | `wire.MaxIndexHashes` |
| Removal needs a valid RepCert quorum of the message's own community, or the author's signature | `proto.VerifyTombstone` |
| Request bodies are never logged | `wire.Handle`, lint rule below |

## Moderators (`/libr/2/vote`, `/libr/2/report`)

| Check | Where |
|---|---|
| 5 req/s per peer, burst 20 | `moderator.peerPerSecond/peerBurst` |
| 0.2 req/s per author key and per reporter key, burst 5; charged only after the signature checks | `moderator.authorPerSecond` |
| At most 16 policy (LLM) decisions at once; more get `busy` | `moderator.maxDeciding` |
| Membership, epoch, timestamp within ±5 min, author signature: all checked before the policy runs | `moderator.Decide` |
| Reports: only verifying certificates, reporter signature over community/epoch/msg/reason, reason ≤ 1 KiB | `moderator.DecideReport` |
| Policy errors: a post is rejected, a reported post is kept | `moderator.Decide`, `DecideReport` |

## Clients (`forum`)

| Check | Where |
|---|---|
| Every certificate is re-hashed and verified; tampered copies are skipped | `forum.Get` |
| Receipts and removal acks must be signed by the peer actually contacted | `forum.checkReceipt`, `checkAck` |
| Tombstones hide a post only after verifying against that post | `forum.Feed` |
| Feed range ≤ 1 week (672 buckets), ≤ 2000 certificate fetches per call | `forum.MaxFeedBuckets`, `MaxFeedFetches` |

## Seed server (`services/librserver`)

| Check | Where |
|---|---|
| 1 req/s per IP, burst 30 (`TRUST_PROXY=1` behind Render) | `src/ratelimit.js` |
| JSON body ≤ 4 KiB (413 otherwise) | `src/app.js` |
| peerId must derive from the public key; signature covers challenge, peerId and addresses; challenge single-use, 60 s | `src/app.js` |
| ≤ 8 addresses of ≤ 256 chars; ≤ 10 000 registrations; ≤ 10 000 pending challenges | `src/app.js`, `src/utils/nonce.js` |
| `/peers` returns ≤ 32 random entries; entries expire (`ENTRY_TTL_SECONDS`, default 600) | `src/registry.js` |
| No database: nothing to leak, nothing persistent | |

## Lint

`forbidigo` bans `fmt.Print*`/`print` in `dbnode`, `moderator`, `wire`, `forum`, `p2p`,
`proto`, `seeds` (`.golangci.yml`).

## Known gaps (accepted for now)

- **Junk author deletes.** A DB node accepts an author-signed tombstone for any message of
  a known community, even one it never saw; anyone can mint keys and messages. Bounded
  only by per-peer rate limits, and peer IDs are free. Fix with contribution-based
  admission (PoSe) or by requiring the node to know the message.
- **Free identities.** Per-peer and per-author limits slow one identity, not a Sybil
  swarm. Per-IP connection limits and the moderator's global decision cap are the
  backstop.
- **Slow requests.** A peer can hold streams open for up to 30 s each; libp2p's per-peer
  stream limits bound how many.
- **Feed truncation.** Buckets with more than 1000 posts, and feeds with more than 2000,
  are cut arbitrarily; junk index entries can crowd out real ones within those caps.
- **Relay metadata.** Relays see who talks to whom, and how much (no content).
- **Seed servers can lie about addresses.** That costs a failed dial, not an
  impersonation: Noise authenticates the peer ID.
