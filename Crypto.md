# Crypto Document

## 🎯 Module objectives

This module holds the functions used to:

- Generating public-private key pairs
- Signing messages after review by mdoerators
- Aggregating moderator signature into message certificates
- Verifying messages with public keys
- Verifying moderation certificates

All operations are called via the internal protocol interface.

---

## 📁 File Structure

```
crypto/
├── config/
│   └── config.go             # Config and env management
├── functions/
│   └── functions.go         # Functions
├── models/
│   └── models.go             # Structs
├── main.go                   # Entry point
├── go.mod                    # Go module definition
└── README.md                 
```

## Functions

### 1. GenerateKeyPair() (ed25519.PrivateKey, ed25519.PublicKey, error)

- **Role**: Generates an Ed25519 keypair.
- **Input**: None
- **Output**: ```ed25519.PrivateKey```, ```ed25519.PublicKey```, ```err```
- **Body**:
```
publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
```
- **Logic**: Uses Go's standard crypto library to securely generate a keypair


### 2. SignMessage(privateKey ed25519.PrivateKey, message string) (ModSign, error)

- **Role**: Signs a given message
- **Input**: ```privateKey```, ```message```
- **Output**: ```ModSign```, ```err```
- **Body**: 
```
sign := ed25519.Sign(privateKey, []byte(message))
modSig := ModSign{
    public_key: privateKey.Public().(ed25519.PublicKey),
    sign: sign,
}
```
- **Logic**:
    - Converts messages into bytes to sign with private key
    - Returns ModSign struct with public key and sign

### 3. VerifySignature(publicKey ed25519.PublicKey, message string, signature []byte) bool

- **Role**: Verifies a message-signature pair
- **Input**: ```publicKey```, ```message```, ```sign```
- **Output**: ```bool```
- **Body**: 
```
verified := ed25519.Verify(publicKey, []byte(message), signature)
```
- **Logic**: Verifies the authenticity of signed message

### 4. AggregateSignatures(senderPrivateKey ed25519.PrivateKey, message string, timestamp int64, signs []ModSignature) (MsgCert, error)

- **Role**: Combines multiple moderator signatures into a certificate.
- **Input**: ```signs []ModSign```
- **Output**: ```MsgCert```
- **Body**: 
```
sign := ed25519.Sign(senderPrivateKey, []byte(message+strconv.Itoa(timestamp)))
MsgCert{
    sender: senderPrivateKey.Public().(ed25519.PublicKey),
    msg: message,
    ts: timestamp,
    mod_cert: signs,
    sign: sign
}
```
- **Logic**: 

### 5. VerifyMsgCert(cert MsgCert) bool

- **Role**: Validates the message certificate before storing
- **Input**: ```MsgCert```
- **Output**: ```bool```
- **Body**: 
```
data := cert.Msg+strconv.Itoa(cert.Timestamp)
if !ed25519.Verify(cert.sender, []byte(data), cert.sign) {
    return false
}
for _, modSign := range cert.ModCert {
    if !ed25519.Verify(modSign.public_key, []byte(message), modSign.sign) {
        return false
    }
}
return true
```
- **Logic**: 
    - Verifies sender's sign over message and timestamp
    - Verifies modsigns over message