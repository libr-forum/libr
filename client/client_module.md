# 💻 LIBR Client Module

## 📌 Overview

The **Client Module** is responsible for initiating and managing the message certification process in the LIBR system. It performs:

- ✅ Validating user-submitted messages
- 🧾 Creating message structs with UNIX timestamp
- 🤝 Communicating with Moderator nodes (via the Network Module)
- 🔐 Verifying moderator responses using the Crypto Module
- 🧠 Creating a `MsgCert` once quorum is achieved
- 📤 Sending the `MsgCert` directly to DB nodes

> Note: The **Crypto Module**, **Network Module**, and **Kademlia Module** are external and not part of this directory.

---

## 🗂️ File Structure

```text
client/
├── main.go                     # Entry point
│
├── handler/
│   └── input_handler.go        # Accepts user input, calls SendToMods
│
├── core/
│   ├── mod_comm.go             # Implements SendToMods and goroutine logic
│   ├── cert_builder.go         # Implements CreateMsgCert
│   ├── send_to_db.go           # Implements SendToDB (inside Client Module)
│   └── validator.go            # Implements isValidMessage
│
├── utils/
│   └── state.go                # Reads current moderators and DB node sets
```

---

## 🧩 Structs (All JSON format)

### 🔸 `ModCert`
```json
{
  "Sign": "string",
  "Pub_key": "string",
  "Status": "approved" | "rejected"
}
```

### 🔸 `Msg`
```json
{
  "message": "string",
  "ts": 1234567890123
}
```

### 🔸 `MsgCert`
```json
{
  "SenderPub_key": "string",
  "Msg": {
    "message": "string",
    "ts": 1234567890123
  },
  "ts": 1234567890123,
  "Modcert": [ /* ModCert[] */ ],
  "sign": "string"
}
```

---

## ⚙️ Functions (Exact Logic Preserved)

### 🔹 `isValidMessage(msg)`

```
Function isValidMessage(msg):
    If msg is not a string:
        Return false
    Trim msg
    If msg is empty or too long:
        Return false
    Return true
```

---

### 🔹 `SendToMods(message)`

```
Function SendToMods(message):

    1. Retrieve current UNIX timestamp → ts

    2. Construct Msg object:
         Msg = Msg{
             message: message,
             ts: ts
         }

    3. Get list of currently online moderators → onlineMods
       Set totalMods = count of onlineMods

    4. Initialize:
        - modcertList = empty list of approved ModCerts
        - approvedCount = 0

    5. For each mod in onlineMods:
        - Send Msg to the mod using the network module
        - Wait for response with a fixed timeout
          (Handled using goroutines and channels)

    6. As responses arrive:
        For each response:
            a. Verify the mod’s signature using crypto module
            b. If valid and status == "approved":
                - Add ModCert to modcertList
                - approvedCount += 1
            c. If response not received or invalid:
                - Decrease totalMods by 1

    7. After processing all mods:
        If approvedCount > totalMods / 2:
            - cert = CreateMsgCert(message, ts, modcertList)
            - SendToDB(cert)

    8. Return modcertList
```

---

### 🔹 `CreateMsgCert(message, ts, modcertList)`

```
Function CreateMsgCert(message, ts, modcertList):

    1. Retrieve sender’s public key → senderPubKey

    2. Construct dataToSign = {
           "message": message,
           "timestamp": ts,
           "modcerts": modcertList
       }

    3. Canonically serialize dataToSign to a string
       (e.g., using json.Marshal with sorted ModCert list)

    4. sign = SignMessage(privateKey, serializedString)

    5. Construct MsgCert = {
           "SenderPub_key": senderPubKey,
           "Msg": {
               "message": message,
               "ts": ts
           },
           "ts": ts,
           "Modcert": modcertList,
           "sign": sign
       }

    6. Return MsgCert
```

> 🔐 `SignMessage(privateKey, message string)` uses:
> ```go
> ed25519.Sign(privateKey, []byte(message))
> ```

---

### 🔹 `SendToDB(cert)`

```
Function SendToDB(cert):

    1. Extract ts = cert.ts

    2. dbNodes = SelectDBNodes(ts)  // From the Kademlia module

    3. For each dbNode in dbNodes:
        - Send cert to dbNode using the Network Module
          (e.g. over custom UDP)

    4. Optionally log or retry failures
```

> 🧠 This function is implemented inside the **Client Module**.

---

## 🔄 Module Interactions

| From        | To              | Purpose                                      |
|-------------|------------------|----------------------------------------------|
| Client      | Moderator Nodes | Send Msg, collect ModCerts                   |
| Client      | Crypto Module   | Sign MsgCert, verify ModCerts                |
| Client      | Network Module  | Sends messages to Mods and DBs               |
| Client      | Kademlia Module | Selects DB nodes based on timestamp          |
| Client      | State Utility   | Loads MOD_JOINED and DB node lists           |

---

## 🧠 Summary of Responsibilities

| Function           | Description                              |
|--------------------|------------------------------------------|
| `isValidMessage()` | Validates message content                |
| `SendToMods()`     | Sends to Mods, verifies responses        |
| `CreateMsgCert()`  | Signs data deterministically             |
| `SendToDB()`       | Sends MsgCert to selected DBs            |

---
