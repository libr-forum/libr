# LIBR

> A Moderated, Censorship-Resilient Decentralized Social Network Framework.

---

## 📖 Overview

**LIBR** is a decentralized protocol and application framework designed for digital public forums that are both **provably censorship-resilient** and **protected against harmful content**.

Traditional centralized social platforms exercise unilateral control over their databases, enabling arbitrary censorship and shadowbanning. Purely unmoderated decentralized networks, on the other hand, often devolve into spam and harassment.

LIBR balances free speech and community safety through:
1. **Decentralized Storage (DHT)**: Messages are stored across a replicated Distributed Hash Table rather than a single corporate server.
2. **Multi-Moderator Quorum Consensus**: Content requiring review (e.g. image attachments or reported posts) is submitted to an open quorum of independent moderators. A message is only approved or deleted when a majority threshold of cryptographic moderator signatures is reached.
3. **Decoupled Architecture**: Discovery, circuit relays, storage nodes, desktop clients, and mobile clients interact seamlessly over libp2p.

---

## 🏛️ System Architecture

```
                  ┌─────────────────────────────────────────┐
                  │        services/librserver              │
                  │   (Discovery, Auth & Node Registry)     │
                  └───────────────────▲─────────────────────┘
                                      │ HTTP / REST
         ┌────────────────────────────┼────────────────────────────┐
         │                            │                            │
┌────────┴────────┐          ┌────────┴────────┐          ┌────────┴────────┐
│ services/relay  │          │    core/db      │          │ core/mod_client │
│ (Circuit Relay) │          │  (Storage Node) │          │ (Desktop & Mod) │
└────────┬────────┘          └────────┬────────┘          └────────┬────────┘
         │ libp2p                     │ libp2p                     │ libp2p
         └────────────────────────────┼────────────────────────────┘
                                      │
                             ┌────────┴────────┐
                             │   mobile-app    │
                             │ (React Native)  │
                             └─────────────────┘
```

The system consists of 5 main components:
1. **`services/librserver`**: Node.js & MongoDB discovery service. Manages active nodes, online moderators, relays, and Ed25519 authentication.
2. **`services/relay`**: Standalone libp2p Circuit Relay (v2) for NAT traversal and peer routing.
3. **`core/mod_client`**: Wails v2 Desktop application for users and moderators with multi-provider LLM auto-moderation (Groq, Gemini, OpenAI, DeepSeek).
4. **`core/db`**: Distributed Hash Table (DHT) storage node that persists message certificates and report certificates.
5. **`mobile-app`**: Cross-platform mobile client (React Native / Expo) embedding a compiled Go Mobile libp2p runtime (`go-core`).

---

## 📋 Prerequisites

Ensure your development environment has the following installed:

* **Go**: `1.21+` (Go 1.23+ recommended) — [Download](https://go.dev/dl/)
* **Node.js**: `18.x` or `20.x` LTS & **npm** — [Download](https://nodejs.org/)
* **MongoDB**: Local MongoDB instance or free [MongoDB Atlas](https://www.mongodb.com/atlas) cluster
* **Wails CLI** (for Desktop client):
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@latest
  ```
* **For Mobile Development (Android)**:
  * Android Studio, Android SDK (API 34/35), Android NDK (r27+), and JDK 17
  * `gomobile`:
    ```bash
    go install golang.org/x/mobile/cmd/gomobile@latest
    gomobile init
    ```

---

## 🚀 Setup & Execution Guide

Follow these steps to run each component locally or in a distributed environment.

```
Recommended Startup Order:
1. services/librserver  ──►  2. services/relay  ──►  3. core/db  ──►  4. core/mod_client  ──►  5. mobile-app
```

---

### 1. Discovery Server (`services/librserver`)

The discovery server maintains the active directory of online moderators, storage nodes, and relays.

1. **Navigate to directory**:
   ```bash
   cd services/librserver
   ```

2. **Install dependencies**:
   ```bash
   npm install
   ```

3. **Configure environment variables**:
   Create a `.env` file (refer to `.env.example`):
   ```env
   PORT=3000
   MONGO_URI=mongodb+srv://<username>:<password>@cluster0.example.mongodb.net/
   ENTRY_TTL_SECONDS=180
   EXPIRY_JOB_INTERVAL_SECONDS=90
   LOG_LEVEL=info
   ```

4. **Start the server**:
   ```bash
   npm start
   # or for development:
   node src/index.js
   ```
   *Health Check*: Verify at `http://localhost:3000/health`.

---

### 2. Circuit Relay Server (`services/relay`)

The relay facilitates peer-to-peer connectivity for clients and nodes located behind NATs and firewalls.

1. **Navigate to directory**:
   ```bash
   cd services/relay
   ```

2. **Install Go dependencies**:
   ```bash
   go mod tidy
   ```

3. **Set environment variables**:
   * **Linux / macOS**:
     ```bash
     export SERVER_URL="http://localhost:3000"
     export RELAY_DOMAIN="localhost"   # or public domain/IP in production
     export PORT="10000"
     ```
   * **Windows (PowerShell)**:
     ```powershell
     $env:SERVER_URL = "http://localhost:3000"
     $env:RELAY_DOMAIN = "localhost"
     $env:PORT = "10000"
     ```

4. **Run the relay**:
   ```bash
   go run relay.go
   ```
   The relay will generate its cryptographic identity, start listening on `/ip4/0.0.0.0/tcp/10000/ws`, and automatically register its multiaddress with `librserver`.

---

### 3. Storage / DHT Node (`core/db`)

Storage nodes listen on libp2p, replicate the DHT routing table, validate moderator certificates, and persist messages.

1. **Navigate to directory**:
   ```bash
   cd core/db
   ```

2. **Configure environment**:
   Create a `core/db/.env` file or export the variables:
   ```env
   SERVER_URL=http://localhost:3000
   DB_PATH=data/libr.db
   ```

3. **Run the node**:
   ```bash
   go run main.go
   ```
   The storage node will query `librserver` for active relays, establish libp2p connections, register itself, and begin serving DHT requests.

---

### 4. Moderator & Desktop Client (`core/mod_client`)

The Desktop application functions both as a client for regular messaging and as a full moderation workstation with automated AI review.

1. **Navigate to directory**:
   ```bash
   cd core/mod_client
   ```

2. **Install frontend dependencies**:
   ```bash
   cd frontend
   npm install
   cd ..
   ```

3. **Configure discovery server (optional)**:
   * Defaults to production or can be overridden via environment:
     ```bash
     # Linux / macOS
     export SERVER_URL="http://localhost:3000"
     # Windows (PowerShell)
     $env:SERVER_URL="http://localhost:3000"
     ```

4. **Run in development mode (with Hot Reload)**:
   ```bash
   wails dev
   ```

5. **Build standalone binary**:
   ```bash
   wails build
   ```
   The compiled executable will be created in `core/mod_client/build/bin/`.

6. **Configuring AI Auto-Moderation**:
   * Open the app and navigate to **Mod Config**.
   * Select your preferred LLM provider (**Groq**, **Gemini / Google NLP**, **OpenAI**, or **DeepSeek**).
   * Enter your API key and save. Keys are securely stored locally on your device in `%APPDATA%\libr\modconfig\modkeys.json` (Windows) or `~/.config/libr/` (Linux/macOS).

---

### 5. Mobile Client (`mobile-app`)

The mobile client is an Expo React Native application backed by a compiled Go library (`gomobile`) for decentralized networking and background moderation.

1. **Compile Go Core for Android**:
   * **Windows (PowerShell)**:
     ```powershell
     cd mobile-app/go-core
     .\build-android.ps1
     cd ../..
     ```
   * **Linux / macOS**:
     ```bash
     cd mobile-app/go-core
     chmod +x build-android.sh
     ./build-android.sh
     cd ../..
     ```
   *This compiles `libr-core.aar` into `mobile-app/android/app/libs/libr-core.aar`.*

2. **Install mobile dependencies**:
   ```bash
   cd mobile-app
   npm install
   ```

3. **Run on Android device or emulator**:
   Ensure an Android emulator is running or a physical device is connected with USB debugging enabled (`adb devices`).
   ```bash
   npx expo run:android
   ```

4. **Start Metro bundler (if not started automatically)**:
   ```bash
   npx expo start
   ```

---

## 🔒 Security & Privacy Architecture

* **Decentralized Identities**: Ed25519 keypairs are generated and stored exclusively on each user's local hardware (`context.filesDir` on Android, `%APPDATA%\libr` on desktop). No private keys ever leave the client.
* **Incognito Mode**: Allows switching between persistent identity and disposable anonymous identities on demand.
* **Moderation Threshold**: Messages with media attachments require consensus from multiple independent moderators before being committed to the public feed.
* **Zero Secrets in Repository**: Database connection strings, API keys, and environment variables are excluded via `.gitignore`.

---

## 🤝 Contributing

We welcome contributions across all layers of the LIBR stack!

1. Fork the repository and create your branch from `main`:
   ```bash
   git checkout -b feat/your-feature-name
   ```
2. Ensure commit messages follow [Conventional Commits](https://www.conventionalcommits.org/):
   ```
   feat: add support for custom relays
   fix: handle websocket reconnection timeout
   ```
3. Test all modified components locally before opening a pull request.
4. Submit a Pull Request targeting `main`.

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
