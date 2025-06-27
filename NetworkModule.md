# 🌐 Network Module Documentation

This module handles the **connection protocols** between different components of the system and manages **data flow logic**.

It comprises two protocol layers:

- **General Protocol:** Underlying peer-to-peer connection management
- **Application Protocol:** Defines how LIBR modules interact over the general network

---

## 📁 File Structure
```
Network/
├── General/
│ ├── Peer/
│ │ └── peer.go
│ └── Relay/
│ └── relay.go
└── Application/
└── main.go
```
---

# ⚙️ General Protocol

The General Protocol defines how nodes connect, communicate, and manage peer-to-peer connections using `libp2p`. It is reusable across any P2P system for connection establishment and data transmission.

---

## 📁 File Structure (General)
```
General
├── Peer
│ └── peer.go
└── Relay
  └── relay.go
```    

---

## 🧹 Peer.go Functions

---

### `NewChatPeer(relayAddr string) (*ChatPeer, error)`

**Purpose:**  
Initializes a new chat peer and sets up all required connections.

**Logic:**  
- Parses the relay address  
- Sets up a `libp2p` peer for relay connection and hole punching  
- Uses Identify Service to fetch peer's public address  
- Defines a StreamHandler to receive messages using the ChatProtocol  
- Returns a `ChatPeer` instance with a configured Host for NAT traversal  

---

### `Start(ctx context.Context) error`

**Purpose:**  
Connects the peer to the specified relay and displays circuit (relay) address info.

**Logic:**  
- Uses the provided Go `context.Context` for connection lifecycle  
- Returns an error if the connection to relay fails  

---

### `handleChatStream(stream network.Stream)`

**Purpose:**  
Handles incoming data streams between peers.

**Logic:**  
- Listens for incoming streams  
- Reads and processes messages from peers  

---

### `ConnectToPeer(ctx context.Context, peerAddress string) error`

**Purpose:**  
Establishes a connection between two peers using Host logic.

**Logic:**  
- First attempts hole punching via `libp2p`  
- If hole punching fails, falls back to relay-based communication  

---

### `SendMessage(peerID peer.ID, message string) error`

**Purpose:**  
Sends a message to the specified peer.

**Logic:**  
- Creates a new stream to the given `peerID`  
- Sends the message over the stream  
- Returns an error if transmission fails  

---

### `GetConnectedPeers() []peer.ID`

**Purpose:**  
Returns a list of all peer IDs the node is currently connected to.

---

## 🔄 Proposed Messaging Flow Changes

- On joining the network, a peer connects to a public relay and listens for incoming messages.  
- When `peerA` wants to connect to `peerB`:
  - Peer A fetches `peerB`'s `peerID` from Kademlia DHT (not direct address)
  - Peer A requests the relay for peer B's address
  - Relay provides addresses to both peers
  - They attempt direct connection via hole punching
  - If hole punching fails, fallback communication occurs via relay

---

# 💻 Application Protocol

This protocol layer describes how the General Protocol integrates with **LIBR's module communication**, ensuring reliable data flow.

---

## 📄 Functions in Application/main.go

---

### `StartNode(relayAddr string)`

**Purpose:**  
Initializes a peer in the LIBR network.

**Logic:**  
- Calls `NewChatPeer` from General Protocol  
- Uses `Start` to connect to the public relay  

---

### `SendToDB(msg Message)`

**Purpose:**  
-Sends a message certificate (MsgCert) from the client to all database nodes.

**Logic:**  
- Fetches database node `peerIDs` from Kademlia  
- Calls `ConnectToPeer` for each DB node  
- Sends the message to all connected DB nodes simultaneously  

---

### `SendToMods(message string, moderatorList []peer.ID)`

**Purpose:**  
-Sends a message to a set of moderator nodes.

**Logic:**  
- Establishes connections with moderators from the provided list 
- PINGS to see which moderators are active
- Sends the message to those moderators
- Collects and relays moderator responses to the client  

---

### `Ping(nodeAddresses []string)`

**Purpose:**  
Checks the liveness of specified nodes.

**Logic:**  
- Attempts to ping the provided addresses  
- Determines if nodes are active  

---

### `BootstrapRelayConnection()`
**Purpose:**  
Connects newly joined peers to default relay(s).
**Logic:**  
To be discussed

---

### `LookupDBNodes(ts Timestamp) []string`
**Purpose:**  
Find DB nodes from Kademlia
**Logic:**  
-Gives a call to Kademlia which returns IP address of required DB nodes as a slice

---

### ‘RegisterNodeToKademlia(peerID)`
**Purpose:**  
When a new node joins, it is used to tell kademlia that a new node is created
**Logic:**  
-Gives a call to Kademlia with  peerID of new node

---

## 🔄 Message Flow (High-Level Design)

1. Peer joins network → assigned a relay.
2. Peer listens for incoming messages.
3. Peer A wants to talk to Peer B:
   - Peer A queries Kademlia for B's peerID.
   - Relay resolves the peerID to address.
   - Hole punching is attempted; fallback to relay messaging if failed.


```text
Client → Relay → Connect
      ↓
     Wait for incoming messages
      ↓
Peer A wants to talk to Peer B:
    1. A looks up B's peerID from Kademlia
    2. Relay resolves address for peerID
    3. Try hole punching → fallback to relay
```
---

## 🧠 Responsibilities of the Network Module

| Role         | Task                                 |
|--------------|--------------------------------------|
| Client       | Connect to relay, send Msg/MsgCert   |
| Moderator    | Connect to relay, receive Msg, send ModCert |
| DB Node      | Accept and verify MsgCerts           |
| Network Mod  | Route all messages via send/relay    |

---

# 🏗 Future Improvements

✅ Integrate relay-assisted address discovery via Kademlia  
✅ Optimize hole punching fallback logic  
✅ Improve error handling in stream management  

---


