# 🛡️ Moderator Module Documentation

## 📌 Module Overview

The **Moderator Module** is responsible for:
- Receiving user messages containing unique ID (timestamp) and the message content
- Sending the message content further for moderation
- Receiving the response and signing it 
- Sending modResponse to client node for aggregation 

---

## 📁 File Structure
```
moderator/
├── main.go                         # Entry point: starts REST server, sets up handlers
│
├── internal/                       # Private implementation code
│   ├── handlers/
│   │   └── moderate_handler.go     # HTTP handler for POST /api/moderate
│   │
│   ├── service/
│   │   ├── moderation.go           # ModerateMsg, SelectModel, AnalyzeContent
│   │   └── signer.go               # ModSign: serialize, sign, emit SignedMsg
│   │
│   └── util/
│       └── http_client.go          # HTTP calls to external moderation APIs
│
├── config/
│   └── config.go                   # LoadConfig(), RunChecks()
│
├── pkg/
│   └── model/
│       └── types.go                # Public types: MsgRequest, ModResponse, SignedMsg
│
├── go.mod
├── go.sum
└── README.md                       # Module documentation
```

---

## 🌐 Endpoints

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
3. Return (1) if clean, or (0) if toxic
```

##  `SelectModel(models []string) (ModelFunc, error)`
- Models for moderation could be selected 


## 3. `AnalyzeToxicity(content string, fn ModelFunc) (string)`

### Purpose:
- A unified wrapper that calls the selected moderation function 

### Logic
```
1. Prepare request with content as JSON
2. Load API key from environment
3. Send POST request 
4. Parse response
5. If harmful/toxic return 0
6. Else return (1)
```

## 4. `ModResponse(msg Msg) (ModResponse, error)`

### Purpose:
- Digitally signs a message (including `timestamp`, `content`, and an optional `status`)

### Logic
```
1. Serialize msg content (Timestamp + Content + (status))
2. Generate hash
3. Sign hash using private key 
4. Export public key
5. Return ModResponse {
   public_key
   sign
   (status)
}

** status included only once full testing confirms acceptable bandwidth/computation tradeoff

```

## 5. `LoadConfig()`

### Purpose:
- Loads environment variables.

### Logic:
```
1. Use godotenv to load `.env` file
2. Set up global config variables
```

## 6. `RunChecks() error`

### Purpose:
- safety check to ensure config and signing keys are correctly loaded at startup.

### Logic:
```
1. Check if loading from dotenv is successful
2. Check if private key for signing is available
3. If either missing, return error
4. Else return nil
```

## 🔄 Interactions

The Moderator Module interacts with other parts of the LIBR system in the following ways:

### 1. Client Module → Moderator Module
- The **Client Module** sends a user-generated message to the Moderator Module for validation.
- The message includes:
  - A content string
  - A timestamp

### 2. Moderator Module (Internal Interaction)
- The Moderator calls `analyzeToxicity()` to send the message content further for moderation.
- This function performs **content moderation**, determining whether the message adheres to community guidelines.

### 3. Crypto Module Interaction 
  - The Moderator Module uses the Crypto Module as a dependency to **hash and sign approved messages**
  - The Crypto module:
    - Generates a hash of the message.
    - Signs it using the moderator’s **private key**.
    - Returns a `ModResponse` 

### 4. Moderator Module → Client Module
  - The Moderator signs the message using its private key.
  - It sends back a **ModResponse**, which contains

---

### 💡 Summary of Interactions

| Source        | Target           | Purpose                          |
|---------------|------------------|----------------------------------|
| Client Module | Moderator Module | Submit message for moderation    |
| Moderator     | Gemini API       | Analyze message toxicity         |
| Moderator     | Crypto Module    | Generate digital signature       |
| Moderator     | Client Module    | Return ModSign                   |






