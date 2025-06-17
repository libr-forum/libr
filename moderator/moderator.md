#  Moderator Module Documentation

##  Module Overview

The **Moderator Module** is responsible for:
- Receiving user messages containing unique ID (timestamp) and the message content
- Sending the message content to the **Google Gemini API** for moderation
- Receiving the response and signing it 
- Forwarding messages to client node for aggregation with modSign and public key

---

##  File Structure
```
moderator/
│
├── handlers/
│   └── messageHandler.go       # Contains HandleMsg() — main HTTP endpoint
│
├── services/
│   ├── moderation.go           # ModerateMsg(), AnalyzeToxicity()
│   └── signer.go               # ModSign() — uses private key to sign approved messages
│
├── crypto/
│   └── keys.go                 # Key generation, loading, and signature helpers
│
├── models/
│   └── message.go              # Structs: Msg, SignedMsg, ModSign
│
├── config/
│   └── env.go                  # LoadConfig(), RunChecks()
│
├── utils/
│   └── httpClient.go           # Reusable HTTP client for Gemini requests
│
├── main.go                     # Entry point — sets up server, routes, and config
├── go.mod                      # Go module definition
├── .env                        # Contains GEMINI_API_KEY and GEMINI_API_URL
└── README.md                   # This exact module documentation
```

---

## Endpoints

The Moderator Module exposes the following REST API for client interaction:

---

### POST `/api/moderate`

**Description:**  
Accepts a user message for moderation, processes it through the Gemini API, signs it, and returns the signed result with the moderator's public key.

---

**Request Body (JSON):**

```json
{
  "timestamp": 1718609422,
  "content": "This is a user message."
}
```
---

## 1. `HandleMsg(w http.ResponseWriter, r *http.Request)`

### Purpose:
- Accepts a JSON message from a client
- Sends it for moderation
- Signs the message 
- Returns the signed message and public key

### Logic:
```
1. Parse JSON body to extract Msg {timestamp, content }
2. Call ModerateMsg(msg)
3. Call ModSign(msg) to sign the content
4. Respond with signed message, public key if approved
```

## 2.  `ModerateMsg(msg Msg) (string)`

### Purpose:
- To check if the msg is good or not 

### Logic
```
1. Extract content from msg
2. Call AnalyzeToxicity(content)
3. Return ("approved") if clean, or ("rejected") if toxic
```

## 3. `AnalyzeToxicity(content string) (string)`

### Purpose:
- Communicates with the Google Gemini moderation API to evaluate message content.

### Logic
```
1. Prepare request with content as JSON
2. Load API key from environment
3. Send POST request to Gemini API
4. Parse response
5. If harmful/toxic return "rejected"
6. Else return ("accepted")
```

## 4. `ModSign(msg Msg) (SignedMsg, error)`

### Purpose:
- Digitally signs the approved message and attaches a public key.

### Logic
```
1. Serialize msg content (Timestamp + Content)
2. Generate hash
3. Sign hash using private key 
4. Export public key
5. Return SignedMsg {
   timestamp,
   content,
   modSign,
   publicKey
}

** if rejected msg is to be stored then add one more attribute as accepted or rejected.

```

## 5. `LoadConfig()`

### Purpose:
- Loads environment variables.

### Logic:
```
1. Use godotenv to load `.env` file
2. Set up global config variables:
   - GEMINI_API_KEY
   - GEMINI_API_URL
```

## 6. `RunChecks() error`

### Purpose:
- safety check to ensure config and signing keys are correctly loaded at startup.

### Logic:
```
1. Check if GEMINI_API_KEY is loaded
2. Check if private key for signing is available
3. If either missing, return error
4. Else return nil

```

## 🔄 Interactions

The Moderator Module interacts with other parts of the LIBR system in the following ways:

### 1. Client Module → Moderator Module
- The **Client Module** sends a user-generated message to the Moderator Module for validation.
- The message includes:
  - A unique identifier (UUID)
  - A content string
  - A timestamp

### 2. Moderator Module (Internal Interaction)
- The Moderator calls `analyzeToxicity()` to send the message content to the **Google Gemini API**.
- This function performs **content moderation**, determining whether the message adheres to community guidelines.

### 3. Moderator Module → Crypto Module
  - The Moderator Module sends the message to the **Crypto Module**.
  - The Crypto module:
    - Generates a hash of the message.
    - Signs it using the moderator’s **private key**.
    - Returns a `ModSign` which includes:
      - `sign`: the digital signature
      - `public_key`: the moderator's public key

### 4. Moderator Module → Client Module
  - The Moderator signs the message using its private key.
  - It sends back a **ModSign**, which contains:
    - The message
    - The timestamp
    - A digital signature
    - The Moderator’s public key

---

### 💡 Summary of Interactions

| Source        | Target          | Purpose                           |
|---------------|------------------|----------------------------------|
| Client Module | Moderator Module | Submit message for moderation    |
| Moderator     | Gemini API       | Analyze message toxicity         |
| Moderator     | Crypto Module    | Generate digital signature       |
| Moderator     | Client Module    | Return ModSign                   |






