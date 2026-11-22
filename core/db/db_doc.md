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
| `/libr/2/index/get` | `{community, bucket}` | `{hashes}` (≤ 1000) |

Every cert is checked with `proto.VerifyMsgCert` against the community moderator sets in
`LIBR_COMMUNITIES` before it is stored or indexed; anything unknown is refused. Requests
are rate limited per peer and frames are capped at 2 MiB. Receipts are signed with the
node key, whose libp2p peer ID clients can check against the peer they called.

## Placement
Content key = `MsgCert.Hash()`; feed index key = `wire.IndexKey(community, bucket)` with
15-minute buckets. Clients store on the K=4 closest DB nodes found through the
`/libr` kad-dht (DB nodes are DHT servers, clients are DHT clients).

## SQLite schema
`certs(hash PK, msg_id, community, ts, body)` and `feed_index(community, bucket, hash)`.
Removal (tombstones) arrives with the report flow.
