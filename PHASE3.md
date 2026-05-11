# PHASE 3 — Multi‑Protocol Clients (TCP Sync, UDP Notifications, WebSocket Chat)

This phase adds three client features to the MangaHub CLI:

1. **TCP Sync client** — receive live progress updates.
2. **UDP Notification client** — receive broadcast chapter release messages.
3. **WebSocket Chat client** — real‑time chat between users.

Below is a complete walkthrough of what was built, how it works, and how to test it.

---

## 1) Phase 3A — TCP Sync Client

### ✅ Goal
Connect to the TCP sync server and receive reading‑progress updates pushed from the server.

### ✅ Implementation
- **Command:** `mangahub sync connect`
- **Update Sender:** `mangahub progress update --manga-id <id> --chapter <n>`

### ✅ Test
**Terminal 1** (server):
```bash
./mangahub.exe server start
```

**Terminal 2** (listener):
```bash
./mangahub.exe sync connect
```

**Terminal 3** (send progress):
```bash
./mangahub.exe progress update --manga-id manga_1 --chapter 50
```

**Expected output (Terminal 2):**
```
Connected to sync server. Listening for updates...
Update: Progress: usr_123|manga_1|50
```

---

## 2) Phase 3B — UDP Notification Client

### ✅ Goal
Allow clients to register a username over UDP and receive broadcast notifications when a new chapter is released.

### ✅ Implementation
#### **Files Added**
- `internal/udp_client/client.go`
  - `Connect(host, port)`
  - `Register(username)`
  - `SendNotification(title, chapter)`
  - `ListenForNotifications(handler)`
  - `Close()`

#### **CLI Command Added**
- `mangahub notifications listen --username <name>`
- `mangahub notifications send --title "<manga>" --chapter <n>`

#### **Why username is required**
UDP does not carry a user identity. We send `REGISTER|username` so the server can map:
```
username -> IP:port
```

### ✅ Test
**Terminal 1** (server):
```bash
./mangahub.exe server start
```

**Terminal 2** (listener):
```bash
./mangahub.exe notifications listen --username alice
```

**Terminal 3** (send):
```bash
./mangahub.exe notifications send --title "One Piece" --chapter 1100
```

**Expected output (Terminal 2):**
```
🔔 New Chapter: One Piece - Chapter 1100
```

---

## 3) Phase 3C — WebSocket Chat Client

### ✅ Goal
Build a real‑time chat client using WebSockets so multiple users can exchange messages.

### ✅ Implementation
#### **Files Added**
- `internal/ws_client/client.go`
  - `Connect(host, port, username)`
  - `Send(text)`
  - `Receive(handler)`
  - `Close()`

#### **CLI Command Added**
- `mangahub chat --username <name>`

#### **Why username is passed via query string**
The server reads `username` from the WebSocket upgrade request (`r.URL.Query()`), so it must be sent as a URL query parameter **before** the connection is established.

### ✅ Test
**Terminal 1** (server):
```bash
./mangahub.exe server start
```

**Terminal 2**:
```bash
./mangahub.exe chat --username alice
```

**Terminal 3**:
```bash
./mangahub.exe chat --username bob
```

**Expected output:**
```
[bob] hello
[alice] hi bob
```

---

## ✅ Phase 3 Completed
At this point, all three protocols are working:

- ✅ TCP sync updates
- ✅ UDP notifications
- ✅ WebSocket chat

Next phase can focus on production hardening (timeouts, graceful shutdown, retry logic, and integration tests).
