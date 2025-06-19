# 💻 LIBR Client Module

## 📌 Overview

The **Client Module** is responsible for orchestrating the complete lifecycle of a user-submitted message in the LIBR protocol. It handles:

- Accepting user messages
- Interfacing with the Crypto Module to:
  - Sign messages
  - Build message certificates (`MsgCert`)
- Communicating with moderator nodes to collect `ModSign`s
- Selecting DB nodes for storage via PRNG
- Sending validated messages to DB nodes
- Querying DB nodes for previously stored messages

---

## 🗂️ File Structure

```
client/
│
├── main.go                     # Entry point
│
├── signer/
│   └── signer.go               # Wrapper over Crypto module for signing
│
├── certbuilder/
│   ├── cert_builder.go         # Handles ModSign collection and MsgCert construction
│   ├── mod_communicator.go     # Handles communication with moderators
│   └── types.go                # Structs: Message, ModSign, MsgCert
│
├── storage/
│   ├── prng_selector.go        # PRNG-based DB node selection
│   └── db_communicator.go      # MsgCert delivery to DB nodes
│
├── query/
│   └── fetcher.go              # Message querying logic by timestamp
│
├── utils/
│   └── state_reader.go         # Parses blockchain state (MOD_JOINED, DB_JOINED, etc.)
```

---

## 🌐 External Endpoints Used

### `POST /api/moderate` (Moderator Node)

**Purpose**: Submit signed message for moderation

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
  "sign": "signature"
}
```

---

## ⚙️ Core Functions

### 1. `SignMessage(message, timestamp) -> (signature, pubKey)` *(delegated to Crypto Module)*

Signs `message + timestamp` using Ed25519 private key.

---

### 2. `SendToModerators(message, timestamp, signature) -> []ModSign`

- Sends signed message to `2M+1` moderators.
- Collects at least `M+1` valid moderator signatures (`ModSign`).

---

### 3. `BuildMsgCert(message, timestamp, modSigns) -> MsgCert`  
*Delegated to Crypto Module*

Constructs a `MsgCert` containing:
- Sender's public key
- Message
- Timestamp
- Moderator signatures (`ModCert`)
- Signature over the full cert

---

### 4. `SelectDBNodes(timestamp) -> []DBNode`

- Uses `SHA256(timestamp)` → 8-byte PRNG seed  
- Selects `R` DB nodes from current active set (read via state)

---

## 🔄 Interactions

| Source          | Target           | Purpose                                      |
|-----------------|------------------|----------------------------------------------|
| Client          | Moderator Nodes  | Send signed message for moderation           |
| Client          | Crypto Module    | Sign messages, build MsgCerts                |
| Client          | State Layer      | Retrieve quorum configurations and node sets |

---

## 📝 Notes & Assumptions

- Ed25519 keypair generated and stored securely on client side
- Timestamp must be valid (non-repeating, monotonic, and recent)
- Retry and fallback logic required for mod node unavailability
- MsgCert must be constructed only after collecting `M+1` `ModSign`s
- DB node selection must be deterministic across all honest clients

---

## 🧠 Summary of Responsibilities

| Function             | Description                                       |
|----------------------|---------------------------------------------------|
| `SignMessage()`      | Signs message before moderation                   |
| `SendToModerators()` | Sends message to mods and collects signatures     |
| `BuildMsgCert()`     | Builds final certificate of moderation approval   |
| `SelectDBNodes()`    | Picks DB nodes using deterministic PRNG logic     |

---

## 🔐 Related Module

👉 See [Crypto Module Documentation](../crypto/README.md) for:
- Key generation
- Signature logic
- MsgCert construction & verification
