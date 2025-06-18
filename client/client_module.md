# 💻 Client Module Documentation

## 📌 Module Overview

The **Client Module** in LIBR is responsible for managing the full lifecycle of a user-submitted message. Its tasks include:

* Accepting user input
* Signing the message with the user's ECDSA private key
* Communicating with moderator nodes to gather `ModSign`s
* Building a `MsgCert` once quorum is reached
* Selecting DB nodes via PRNG logic
* Sending the `MsgCert` to selected DB nodes for storage
* Querying DB nodes to retrieve messages based on timestamps

---

## 📁 File Structure

```
client/
│
├── main.go                     # Entry point: initializes flow
│
├── signer/                    
│   └── signer.go              # Handles message signing using ECDSA
│
├── certbuilder/
│   ├── cert_builder.go        # Builds MsgCert from ModSigns
│   ├── mod_communicator.go    # Sends signed messages to moderators and collects ModSigns
│   └── types.go               # Structs: Message, ModSign, MsgCert
│
├── storage/
│   ├── prng_selector.go       # Selects DB nodes using PRNG(seed = timestamp)
│   └── db_communicator.go     # Sends MsgCert to selected DBs for storage
│
├── query/
│   └── fetcher.go             # Fetches messages from DB nodes based on timestamp
│
├── utils/
│   └── state_reader.go        # Parses state transactions to get active mod/DB node lists
```

---

## 🌐 Endpoints (Consumed)

### POST `/api/moderate` (Moderator Node)

**Purpose**: To send signed message for validation and receive `ModSign`

**Request:**

```json
{
  "message": "This is a user message.",
  "timestamp": 1718609422,
  "user_signature": "hex-string",
  "user_public_key": "hex-string"
}
```

**Response:**

```json
{
  "public_key": "pubkey",
  "sign": "sign"
}
```

---

## ⚙️ Core Functions

### 1. `SignMessage(message, timestamp) -> (signature, pubKey)`

**Purpose**: Signs a message + timestamp using the user's private key.

**Logic:**

```
1. Serialize (message + timestamp)
2. Compute SHA-256 hash
3. Sign hash using ECDSA private key
4. Return signature and public key
```

---

### 2. `SendToModerators(message, timestamp, signature) -> []ModSign`

**Purpose**: Sends signed message to `2M+1` moderators and collects `M+1` valid signatures.

**Logic:**

```
1. Prepare JSON payload with message, timestamp, signature, and pubkey
2. Send POST requests to moderators in parallel or sequentially
3. Validate each ModSign returned
4. Stop after M+1 valid ModSigns are collected
```

---

### 3. `BuildMsgCert(message, timestamp, modSigns) -> MsgCert`

**Purpose**: Constructs a message certificate with M+1 moderator approvals.

**Logic:**

```
1. Validate all mod signatures
2. Construct MsgCert object:
   - sender
   - message
   - timestamp
   - mod_signatures[]
```

---

### 4. `SelectDBNodes(timestamp) -> []DBNode`

**Purpose**: Deterministically selects `R` DB nodes using PRNG based on timestamp.

**Logic:**

```
1. Hash timestamp (SHA256)
2. Extract first 8 bytes as PRNG seed
3. Load list of active DB nodes from state
4. Use PRNG(seed) to select R DB nodes
```

---

### 5. `SendToDBs(msgCert) -> status`

**Purpose**: Sends MsgCert to selected DB nodes for storage.

**Logic:**

```
1. Retrieve DB node IPs and ports from state
2. Send MsgCert to each selected DB node using POST /store
3. Handle retry/failure if any DB node is down
```

---

### 6. `QueryMessage(timestamp) -> message`

**Purpose**: Fetches a previously stored message from the DB nodes.

**Logic:**

```
1. Use same PRNG logic to select DB nodes
2. Send GET /query to selected DB nodes
3. Return first successful response
```

---

## 🔄 Interactions

### 1. Client → Moderator Nodes

* Sends signed messages
* Collects `ModSign`s (moderator approvals)

### 2. Client (Internal)

* Aggregates moderator signatures to build `MsgCert`

### 3. Client → DB Nodes

* Sends `MsgCert` for storage
* Queries timestamp-indexed messages

### 4. Client → State Layer (Hashchain)

* Reads `MOD_JOINED`, `DB_JOINED`, etc. to determine quorum participants

---

## 📝 Notes & Assumptions

* Assumes user has a valid ECDSA keypair on device
* All messages must include timestamp and signature before moderation
* Message moderation is synchronous for now (can be improved)
* DB node selection must be deterministic and reproducible
* Retry logic must handle slow or unreachable nodes

---

## 🧠 Summary of Responsibilities

| Function             | Role                                           |
| -------------------- | ---------------------------------------------- |
| `SignMessage()`      | Authenticates user's message                   |
| `SendToModerators()` | Collects quorum of moderator approvals         |
| `BuildMsgCert()`     | Aggregates `ModSign`s into a valid certificate |
| `SelectDBNodes()`    | Chooses where to store the message             |
| `SendToDBs()`        | Forwards approved message to storage           |
| `QueryMessage()`     | Fetches messages by timestamp from DB nodes    |

---

```
```
