# 💻 LIBR Client Module

## 📌 Overview

The **Client Module** is responsible for orchestrating the partial lifecycle of a user-submitted message in the LIBR protocol. It handles:

- Accepting user messages
- Performing user-side validation and clock sync checks
- Interfacing with the Crypto Module to:
  - Sign messages
  - Build message certificates (`MsgCert`)
- Communicating with moderator nodes to collect `ModSign`s
- Passing `MsgCert` to the Database Module for DB node selection and storage

While the client uses functions like `SelectDBNodes()` and `SendToDBs()`, they **belong to the Database Module** and are imported from there.

---

## 🗂️ File Structure
client/
│
├── main.go # Entry point
│
├── signer/
│ └── signer.go # Wrapper over Crypto module for signing
│
├── certbuilder/
│ ├── cert_builder.go # Handles ModSign collection and MsgCert construction
│ ├── mod_communicator.go # Handles communication with moderators
│ └── types.go # Structs: Message, ModSign, MsgCert
│
├── validator/
│ └── message_validator.go # Validates message content and timestamp sanity
│
├── utils/
│ ├── timesync.go # Clock sync helpers
│ └── state_reader.go # Parses blockchain state (MOD_JOINED, etc.)

---

## 🌐 External Endpoints Used

### `POST /api/moderate` (Moderator Node)

**Purpose:** Submit signed message for moderation

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
##⚙️ Core Functions
###1. ValidateMessage(message, timestamp) -> error (validator)
Ensures:

Message is not empty and within allowed size

Timestamp is recent (within drift margin of system clock)

###2. IsClockSynced(remoteTimestamps) -> bool (utils/timesync.go)
Optionally checks time drift by comparing remote mod-provided timestamps with local time

###3. SignMessage(message, timestamp) -> (signature, pubKey) (Crypto Module)
Signs the message using Ed25519 private key.

###4. SendToModerators(message, timestamp, signature) -> []ModSign
Sends signed message to 2M+1 moderators

Collects at least M+1 valid ModSigns

###5. BuildMsgCert(message, timestamp, modSigns) -> MsgCert (Crypto Module)
Builds the message certificate using:

Moderator approvals

Client signature over final payload

###6. PassToDBModule(msgCert) (Database Module)
Calls SelectDBNodes(timestamp)

Sends msgCert to selected DB nodes via SendToDBs()

##🔄 Interactions
Source	Target	Purpose
Client	Moderator Nodes	Send signed message for moderation
Client	Crypto Module	Sign messages, build MsgCerts
Client	Validator/Time	Check for validity and clock correctness
Client	Database Module	Select DBs, store MsgCert (used, not owned)
Client	State Layer	Retrieve MOD_JOINED quorum info

##📝 Notes & Assumptions
Ed25519 keypair is securely generated (via Crypto Module)

Message content is validated before sending

Clock drift is managed at the client side before timestamp use

MsgCert creation is allowed only after receiving sufficient ModSigns

DB node interaction is abstracted out and delegated

##🧠 Summary of Responsibilities
Function	Description
ValidateMessage()	Ensures user message is safe and timestamp is sane
IsClockSynced()	Detects major clock drift from remote timestamps
SignMessage()	Signs the message before moderation
SendToModerators()	Sends message to moderators
BuildMsgCert()	Builds quorum-signed MsgCert
PassToDBModule()	Passes control to DB Module for further handling

##🔐 Related Modules
👉 See Crypto Module for:

Key management

Signing logic

MsgCert construction

##👉 See Database Module for:

DB node selection

MsgCert storage

Querying functionality

