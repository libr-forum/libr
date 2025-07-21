# 🛡️ Database Module Documentation

## 📌 Module Overview

1. **Receives `MsgCert` JSON** when the node is one of the *k* nearest.
2. **Parses JSON** into a `MsgCert` Go struct.
3. **Checks required fields**: sender, msg, ts, and at least one mod_cert.
5. **Validates signature over MsgCert** by the client node
4. **Ensures node allocation**: confirms current node is among designated *k* storers.
6. **Uses PostgreSQL for storage** 
7. **Provides retrieval functions** to fetch
---

## 📁 File Structure
```
db/
├── main.go            
├── internal/
│   └── msgcert/                # Core logic for MsgCert handling
│       ├── handler.go          # HTTP handlers (/store, /fetch)
│       ├── store.go            # MsgCertStore interface and implementation
│       └── validator.go        # signature, schema, and allocation checks
├── db/
│   └── postgres/
│       ├── migrations/         # SQL migration scripts
│       ├── schema.sql          # table definitions
│       └── pgstore.go          # PostgreSQL implementation of MsgCertStore
├── models/
│   └── msgcert.go              # `MsgCert`, `CertEntry` structs
├── config/
│   └── config.go               # DB config loader (env variables)
├── go.mod
└── README.md
```

---


## 📥 Data Formats

### Client Submission (`MsgCert` JSON):
```json
{
  "sender":   "sender_public_key",
  "msg":      "the message text",
  "ts":       "timestamp_or_unique_id",
  "mod_cert": [
    {
      "public_key": "moderator_public_key",
      "sign":       "signature_string",
      "status":     "status"
    }
  ]
}
```

**Stored in Database(PostgreSQL)**:
```json
{
  "key":   "generated_by_kademlia",
  "value": { /* MsgCert JSON as above */ }
}
```

## 🌐 HTTP Handlers

### 1. **PingHandler**
**Response**
```json
{
  "message": "pong",
  "node_id": "<your_node_id>"
}   
```
### 2. **FindNodeHandler/FindValueHandler**
**Response**\ 
Returns array of **Node** up to k closest nodes to the queried ID.

Structure of Node:
```json
    "ID": "node_id",
    "IP": "ip_address",
    "Port": 30303
    "LastSeen": 0
```
---

## 🗄️ PostgreSQL Storage Integration

 ### 3a. Configuration (config/config.go): 
      Reads env variables, and store it in a DNS string
      Contains func LoadDBConfig() (*DBConfig, error)

### 3b. Migration Scripts (db/postgres/migrations/)
      Creates new database if not already present and make SQL files.
      Defines a msgcerts table.

### 3c. PostgreSQL Store Implementation (db/postgres/pgstore.go)
    NewPgStore(cfg *DBConfig) (*PgStore, error):
    Opens a *sql.DB using the postgres driver and verifies connection with Ping().

    Store(key string, cert MsgCert) error:
    Serializes cert to JSON and inserts into msgcerts using:

    Fetch(key string) (*MsgCert, error):
    Scans results into a MsgCert struct, decoding mod_cert from JSONB. Returns nil on sql.ErrNoRows.

    Close() error:
    Gracefully shuts down the database connection pool via db.Close().
 

## 🔄 Interactions

### Kademlia
- **find_node <key>**
→ Checks if current node is among k closest to key before storing a MsgCert.

- **find_value <key>**
→ Fetches a MsgCert. Queries network if not stored locally.

- **store <key> <val>**
→ Stores a MsgCert if the node is one of the k closest. Returns ack.

- **ping**
→ Used to check if other nodes are alive (e.g., during routing table updates).

- **bootstrap <[]nodes>**
→ Initializes the node with a list of known peers to join the DHT.


