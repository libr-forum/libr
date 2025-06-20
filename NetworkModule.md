# P2P Communication Module

## Module Objectives

This module is responsible for:
- 📡 Communication between two peers in a P2P network.
- 🔐 Defining the protocol based on **hole punching** and **UDP communication**.
- 🛠️ Checking if a connection is established.
- 📨 Sending data between the peers.

---

# Network Module File Structure

```
Network/
├── Establish Connection/
│   ├── Relay.go             # Uses Relay to share information between Nodes (have functions EstablishConnection and FetchFromRelay
│   └── Holepunching.go      # Creates State Table in respective NAT (have function Holepunch)
│
├── Checks/
│   └── Check.go             # Checks if hole punching succeeded(have function check)
│
└── Send Data/
    └── sendData.go          # Sends data after connection is established( have function SendData)
```

## Functions

### 1. `EstablishConnection(source IP, source Port, Relay Address)`

**Purpose**:  
📡 Used by each peer to start listening and receiving messages at a port. Sends the relay its own IP and port.

**Logic**:
1. `source IP` and `source Port` are used to act as a server.
2. Send to Relay:
   - Self IP
   - Self Port
   - Device ID

```
example data to relay:
{
IP: (source devie IP)
PORT: (source device port)
ID: (device ID for identification)
}
```
---

### 2. `FetchFromRelay(relay address, target device ID)`

**Purpose**:  
📥 Helps each peer get the IP and port of the other peer they are trying to connect to via relay.

**Logic**:  
Returns:
- 🎯 Target IP  
- 🎯 Target Port

```
example request to relay:
{
targetID: (identification ID of target node)
}

example response:
{
targetIP: (IP of target device)
targetPort: (port number of target device)
}
```

---

### 3. `Holepunch(source IP, source PORT, fetched IP, fetched PORT)`

**Purpose**:  
🔓 Creates a mapping in the NAT table to allow the other peer to communicate.

**Logic**:
- 🔁 Both peers send a packet to one another.
- ✅ If successful, a hole is punched in the NAT table of both peers.

```
example request to target Node:
{
message: This is a connection request from {Source IP}
}
```
---

### 4. `Check(source IP, source PORT, fetched IP, fetched PORT)`

**Purpose**:  
📶 Checks whether hole punching was successful via acknowledgements.

**Logic**:
1. Send a packet and wait for an acknowledgement.
2. Listen for acknowledgement.
3. If received, send back an acknowledgement.

```
example package sent for acknowledge:
{
IP: {SourceIP:PORT}
}

Example Acknowledgement from target Node:
{
Acknowledgement: Verifed IP
}

Example Acknowledgement to target Node:
{
Status: Acknowledgement Received
}
```

✅ **If this step succeeds, both peers are considered connected.**

---

### 5. `SendData(source IP, source PORT, target IP, target PORT, data)`

**Purpose**:  
📤 Sends data after hole punching is complete.  
This function will be used uniformly across modules (e.g., crypto, moderator, database), and each module will handle the data according to its specific requirements.

---
