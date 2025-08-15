# 🔐 Crypto Module

## 🎯 Module Objectives

This module provides secure cryptographic utilities for:

* 🔑 Generating public–private key pairs
* 💾 Saving and loading keys to/from the filesystem
* ✍️ Signing messages with a private key
* ✅ Verifying signatures with a public key

All cryptographic operations are handled internally using Ed25519 keys.

---

## 📁 File Structure

```
crypto/
├── config/
│   └── config.go             # Platform-specific key file paths
├── cryptoutils/
│   └── cryptoutils.go        # Cryptographic functions
├── go.mod                    # Go module definition
└── README.md                 # Project documentation
```

---

## ⚙️ Key File Configuration

Key paths are determined dynamically at runtime based on the OS:

* **Windows:** `%APPDATA%\yourapp\keys\priv.key`
* **macOS:** `~/Library/Application Support/yourapp/keys/priv.key`
* **Linux:** `~/.config/yourapp/keys/priv.key`
* **Fallback:** `./keys/priv.key`

Public keys are stored in the same folder as `pub.key`.

---

## 🔧 Functions

### 1. `GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error)`

- **Role:** Generate a new Ed25519 key pair and save both keys to disk.
- **Inputs:** None
- **Returns:** public key, private key, error

---

### 2. `LoadKeys() (ed25519.PublicKey, ed25519.PrivateKey, error)`

- **Role:** Load keys from disk. If the private key is missing or invalid, generate a new key pair. If the public key is missing, regenerate it from the private key and store it.
- **Inputs:** None
- **Returns:** public key, private key, error

---

### 3. `SignMessage(privateKey ed25519.PrivateKey, message string) (string, error)`

- **Role:** Sign a message using a private key.
- **Inputs:** private key, message
- **Returns:** sign (88 byte base64-encoded string), error

---

### 4. `VerifySignature(publicKey ed25519.PublicKey, message string, sign string) bool`

- **Role:** Verify a message-signature pair using the public key.
- **Inputs:** public key, message, sign (88 byte base64-encoded string)
- **Returns:** true, false

---