# Crypto Document

## 🎯 Module objectives

This module holds the functions used to:

- Generating public-private key pairs
- Storing and retrieving keys
- Signing messages with private key
- Verifying messages with public key

All operations are called via the internal protocol interface.

---

## 📁 File Structure

```
crypto/
├── config/
│   └── config.go             # Config and env management
├── functions/
│   └── functions.go          # Functions
├── models/
│   └── models.go             # Structs
├── keys/                     # Stored binary keys
│   ├── pub.key     
│   └── priv.key
├── main.go                   # Entry point
├── go.mod                    # Go module definition
└── README.md                 
```

---

## 🔨 Config File

```
const (
    PrivateKeyPath = "keys/priv.key"
    PublicKeyPath  = "keys/pub.key"
)
```

---

## Functions

### 1. GenerateKeyPair() (ed25519.PrivateKey, ed25519.PublicKey, error)

- **Role**: Generates an Ed25519 keypair and store them.
- **Input**: None
- **Output**: ```ed25519.PrivateKey```, ```ed25519.PublicKey```, ```err```
- **Body**:
```
priv, pub, err := ed25519.GenerateKey(rand.Reader)
if err != nil {
    return nil, nil, err
}
if err := os.MkdirAll(filepath.Dir(config.PrivateKeyPath), 0700); err != nil {
    return nil, nil, err
}
if err := os.WriteFile(config.PrivateKeyPath, priv, 0600); err != nil {
    return nil, nil, err
}
if err := os.WriteFile(config.PublicKeyPath, pub, 0644); err != nil {
    return nil, nil, err
}
return priv, pub, nil
```

### 2. LoadKeys() (ed25519.PrivateKey, ed25519.PublicKey, error)

- **Role**: Loads keys and if not found generates a new pair
- **Input**: None
- **Output**: ```ed25519.PrivateKey```, ```ed25519.PublicKey```, ```err```
- **Body**:
```
privData, err := os.ReadFile(config.PrivateKeyPath)
if err != nil {
    return GenerateKeyPair()
}
pubData, err := os.ReadFile(config.PublicKeyPath)
if err != nil {
    if len(privData) != ed25519.PrivateKeySize {
        return nil, nil, errors.New("invalid private key size")
    }
    pubKey := privData[32:]
    return ed25519.PrivateKey(privData), ed25519.PublicKey(pubKey), nil
}
return ed25519.PrivateKey(privData), ed25519.PublicKey(pubData), nil
```

### 3. SignMessage(privateKey ed25519.PrivateKey, message string) ([]byte, error)

- **Role**: Signs a given message with provided private key
- **Input**: ```privateKey```, ```message```
- **Output**: ```[]byte```, ```err```
- **Body**: 
```
sign := ed25519.Sign(privateKey, []byte(message))
return sign, nil
```

### 4. VerifySignature(publicKey ed25519.PublicKey, message string, sign []byte) bool

- **Role**: Verifies a message-signature pair
- **Input**: ```publicKey```, ```message```, ```sign```
- **Output**: ```bool```
- **Body**: 
```
return ed25519.Verify(publicKey, []byte(message), sign)
```