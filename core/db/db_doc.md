# LIBR v2 DB node

Stores moderator-approved posts. Code: protocol logic in `core/libr/dbnode`, wire format
in `core/libr/wire`, certificate rules in `core/libr/proto`; this module adds the SQLite
store (`sqlstore/`) and the binary (`main.go`, configured by environment, see its doc
comment).

## Protocols (one request/response per libp2p stream, length-prefixed JSON)
| ID | Request | Response |
|---|---|---|
| `/libr/2/store` | `proto.MsgCert` | `ReceiptResp` (signed `StoreReceipt` or `err`) |
| `/libr/2/get` | `{hash}` | `{cert}` or `err: notfound` |
| `/libr/2/index/put` | `proto.MsgCert` | `ReceiptResp` |
| `/libr/2/index/get` | `{community, bucket}` | `{hashes, tombs}` (≤ 1000 each) |
| `/libr/2/remove` | `{msg, tomb}` | `{ack}` (signed `RemoveAck`) or `err` |

Every cert is checked with `proto.VerifyMsgCert` against the community moderator sets in
`LIBR_COMMUNITIES` before it is stored or indexed; anything unknown is refused. Requests
are rate limited per peer and frames are capped at 2 MiB. Receipts are signed with the
node key, whose libp2p peer ID clients can check against the peer they called.

A removal (`proto.Tombstone`: a moderator `RepCert` quorum of the message's own
community, or the author's signed delete) is verified against the message, recorded,
and deletes every certificate and index entry of that message ID. Store and index/put
then answer `removed` for any new certificate of it. Communities in `LIBR_OPTOUT` are
answered `optout`; clients move on to the next-closest node.

## Placement
Content key = `MsgCert.Hash()`; feed index key = `wire.IndexKey(community, bucket)` with
15-minute buckets. Clients store on the K=4 closest DB nodes found through the
`/libr` kad-dht (DB nodes are DHT servers, clients are DHT clients).

## Joining the network
Bootstrap peers come from `LIBR_BOOTSTRAP`, the seed list built into the binary, the peer
cache (`peers.json` in the data dir, rewritten after each join) and `LIBR_SEED_URLS`.
With seed URLs set the node also registers its own addresses there (key-bound, renewed
at half the TTL). `LIBR_PRIVATE=1` is for nodes behind NAT: they reserve on
`LIBR_RELAYS` and advertise only circuit addresses. `LIBR_PUBLIC=1` makes a reachable
node a relay too. Production setup: `deploy/compose.yml` and `docs/RELEASE.md`.

## SQLite schema
`certs(hash PK, msg_id, community, ts, body)`, `feed_index(community, bucket, hash,
msg_id)` and `tombstones(msg_id PK, community, bucket, body)`.
