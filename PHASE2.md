# MangaHub - Phase 2: Protocol Servers Implementation Guide

**Status:** ⏳ In Progress (Phase 2)  
**Objective:** Build & integrate 5 concurrent backend servers  
**Timeline:** Weeks 3-5  
**Focus:** Goroutines, channels, network protocols, concurrent architecture

---

## 📋 Phase 2 Overview & Architecture

### What Happens in Phase 2?

Phase 2 transforms the single HTTP server from Phase 1 into a **production-grade multi-protocol backend system**. When a user runs `mangahub server start`, all 5 servers launch simultaneously in separate goroutines and communicate via shared SQLite database.

### The 5 Backend Servers

```
mangahub server start
    ↓
database.Init()
    ↓
    ├─→ [Goroutine] http_server.Start("8080")        HTTP REST API (Phase 1 ✓)
    ├─→ [Goroutine] tcp_server.Start("9090")         TCP Sync Server (Phase 2 🔨)
    ├─→ [Goroutine] udp_server.Start("9091")         UDP Notifications (Phase 2 🔨)
    ├─→ [Goroutine] grpc_server.Start("9092")        gRPC Service (Phase 2 🔨)
    └─→ [Goroutine] websocket_server.Start("9093")   WebSocket Chat (Phase 2 🔨)

Main goroutine blocks (select {}) ← All servers run in background
```

### Key Architectural Patterns in Phase 2

| Pattern | Used In | Purpose |
|---------|---------|---------|
| **Hub-and-Spoke** | TCP, WebSocket | Central hub routes messages to all clients |
| **Goroutines + Channels** | TCP, WebSocket | Concurrent client handling |
| **Broadcast Loop** | TCP, UDP, WebSocket | Fan-out messages to all subscribers |
| **Event-Driven** | WebSocket hub | Register/Unregister/Broadcast events |
| **Protocol Buffers** | gRPC | Type-safe, binary serialization |
| **Request/Response** | gRPC | Synchronous RPC calls with context |

---

## 🏗️ Phase 2 Project Structure

```
mangahub/
├── internal/
│   ├── tcp_server/
│   │   ├── server.go          ← Main TCP listener (Port 9090)
│   │   └── hub.go             ← Connection registry + broadcast channel
│   │
│   ├── udp_server/
│   │   └── server.go          ← UDP listener (Port 9091)
│   │
│   ├── grpc_server/
│   │   └── server.go          ← gRPC server (Port 9092)
│   │
│   ├── websocket_server/
│   │   ├── server.go          ← WebSocket hub + HTTP upgrade (Port 9093)
│   │   ├── hub.go             ← Client registry + broadcast
│   │   └── client.go          ← Per-connection read/write pumps
│   │
│   └── http_server/
│       ├── server.go          ← HTTP REST (Port 8080) from Phase 1
│       └── routes.go          ← Route registration
│
├── proto/
│   └── manga/
│       ├── manga.proto        ← Service definitions
│       ├── manga.pb.go        ← Generated (do NOT edit)
│       └── manga_grpc.pb.go   ← Generated (do NOT edit)
│
├── cmd/
│   ├── server.go              ← MODIFIED: Launch 5 servers concurrently
│   └── main/main.go           ← Entry point
│
└── go.mod                      ← UPDATED: Add gRPC, WebSocket dependencies
```

---

## 🔄 Data Flow in Phase 2

### Example 1: User Updates Reading Progress (All Protocols)

```
User runs CLI:
./mangahub.exe progress update --manga-id manga_1 --chapter 50

    ↓
[HTTP Server 8080]
POST /progress/update
    ├─→ Validate JWT
    ├─→ Update SQLite: user_progress table
    └─→ Response: "✓ Updated"

    ↓
[TCP Server 9090]
Broadcast to all connected clients:
"USER_123|MANGA_1|50"

    ↓
All TCP clients listening receive:
"Update: USER_123 reading OnePiece Ch. 50"
```

### Example 2: New Chapter Released (UDP Server)

```
Admin sends notification:
./mangahub.exe notifications send --title "OnePiece" --chapter "1050"

    ↓
[UDP Server 9091]
Receive command packet
    ├─→ Parse: "NOTIFY|OnePiece|1050"
    ├─→ For each registered client UDP address
    └─→ SendTo: "New Chapter: OnePiece - Chapter 1050"

    ↓
All registered UDP clients receive:
"🔔 New Chapter: OnePiece - Chapter 1050"
```

### Example 3: Search Manga (gRPC Server)

```
CLI client needs to search:
./mangahub.exe manga search "OnePiece"

    ↓
[gRPC Client] (internal to CLI)
Call: MangaService.SearchManga(query="OnePiece")

    ↓
[gRPC Server 9092]
RPC Handler processes:
    ├─→ Query SQLite: SELECT * FROM manga WHERE title LIKE "%OnePiece%"
    └─→ Return: [Manga{id, title, author, chapters...}, ...]

    ↓
CLI displays:
"- OnePiece by Oda (Ch. 1050)"
```

### Example 4: Real-time Chat (WebSocket Server)

```
User joins chat:
./mangahub.exe chat join --username alice

    ↓
[HTTP Server 8080]
GET /ws?username=alice
    ├─→ Upgrade connection to WebSocket
    └─→ Create Client in Hub

    ↓
[WebSocket Hub 9093]
Register: hub.register <- client_alice

    ↓
User types: "hello everyone!"

    ↓
[WebSocket Client] readPump()
Send to hub: hub.broadcast <- Message{username:"alice", text:"hello..."}

    ↓
[WebSocket Hub] run()
For each client in hub.clients:
    client.send <- message

    ↓
All connected WebSocket clients receive:
"[alice]: hello everyone!"
```

---

## 🎯 Phase 2 Implementation Details

### 2.1: TCP Sync Server (Port 9090)

**What it does:**
- Maintains persistent connections to multiple clients
- Broadcasts reading progress updates to all connected clients
- Enables real-time awareness of other users' reading activity

**Architecture:**
```
TCP Listener (9090)
    ↓
Accept connection
    ↓
Spawn goroutine: handleConnection()
    ├─→ Register conn in hub.clients map
    ├─→ Read messages from conn (loop)
    ├─→ Send to hub.broadcast channel
    ├─→ Cleanup on close
    
Separate goroutine: broadcastLoop()
    └─→ Receive from hub.broadcast channel
        └─→ Write to all registered connections
```

**Key Components:**
```go
// Hub manages all connections
type Hub struct {
    clients   map[net.Conn]bool  // All connected clients
    broadcast chan []byte        // Messages to send
    mu        sync.Mutex         // Protect map access
}

// Server wraps the hub
type Server struct {
    listener net.Listener
    hub      *Hub
}
```

**Message Format (ASCII):**
```
USER_ID|MANGA_ID|CHAPTER\n
Example: usr_123|manga_456|42\n
```

**Testing:**
```bash
# Terminal 1
./mangahub.exe server start

# Terminal 2 (Client 1)
telnet localhost 9090
usr_001|manga_100|10

# Terminal 3 (Client 2)
telnet localhost 9090
# Should receive: usr_001|manga_100|10 from Terminal 2
```

---

### 2.2: UDP Notification System (Port 9091)

**What it does:**
- Stateless, fire-and-forget notifications
- Clients register their UDP address with server
- When new chapters release, send UDP packets to all registered clients
- No persistent connection (lightweight)

**Architecture:**
```
UDP Listener (9091)
    ↓
Listen for packets (loop)
    ├─→ If "REGISTER|username" → Store (username, remote_addr)
    └─→ If "NOTIFY|..." → Send to all registered addresses

Clients:
    ├─→ Send "REGISTER|alice" to server
    └─→ Receive UDP packets when broadcasts happen
```

**Key Components:**
```go
type Server struct {
    conn    *net.UDPConn
    clients map[string]*net.UDPAddr  // username -> remote address
    mu      sync.Mutex
}
```

**Message Format:**
```
REGISTER|username\n
NOTIFY|manga_title|chapter_number\n

Example:
REGISTER|alice
NOTIFY|OnePiece|1050
```

**Testing:**
```bash
# Terminal 1
./mangahub.exe server start

# Terminal 2 (Register)
echo "REGISTER|alice" | nc -u localhost 9091

# Terminal 3 (Send notification)
echo "NOTIFY|OnePiece|1050" | nc -u localhost 9091

# Terminal 2 receives:
# "New Chapter: OnePiece - Chapter 1050"
```

---

### 2.3: gRPC Manga Service (Port 9092)

**What it does:**
- Fast binary RPC for internal service communication
- Strongly typed messages (Protocol Buffers)
- Three RPC calls: GetManga, SearchManga, UpdateProgress
- Used internally by other servers and CLI

**Architecture:**
```
gRPC Listener (9092)
    ↓
Register MangaService implementation
    ├─→ GetManga RPC → Query DB by ID
    ├─→ SearchManga RPC → Query DB by title (LIKE)
    └─→ UpdateProgress RPC → Insert/Update user_progress

Clients call via:
    ctx := context.WithTimeout(...)
    resp, err := client.GetManga(ctx, req)
```

**Protocol Buffer Definition (`proto/manga.proto`):**
```protobuf
syntax = "proto3";
package mangahub;

message GetMangaRequest {
    string manga_id = 1;
}

message GetMangaResponse {
    string id = 1;
    string title = 2;
    string author = 3;
    int32 total_chapters = 4;
    string status = 5;  // "ongoing" | "completed"
}

message SearchMangaRequest {
    string query = 1;
}

message SearchMangaResponse {
    repeated GetMangaResponse results = 1;
}

message UpdateProgressRequest {
    string user_id = 1;
    string manga_id = 2;
    int32 current_chapter = 3;
}

message UpdateProgressResponse {
    bool success = 1;
    string message = 2;
}

service MangaService {
    rpc GetManga(GetMangaRequest) returns (GetMangaResponse);
    rpc SearchManga(SearchMangaRequest) returns (SearchMangaResponse);
    rpc UpdateProgress(UpdateProgressRequest) returns (UpdateProgressResponse);
}
```

**Key Components:**
```go
type MangaServer struct{}

func (s *MangaServer) GetManga(ctx context.Context, req *pb.GetMangaRequest) (*pb.GetMangaResponse, error) {
    // Query DB, return response
}

func (s *MangaServer) SearchManga(ctx context.Context, req *pb.SearchMangaRequest) (*pb.SearchMangaResponse, error) {
    // Query DB with LIKE, return results
}

func (s *MangaServer) UpdateProgress(ctx context.Context, req *pb.UpdateProgressRequest) (*pb.UpdateProgressResponse, error) {
    // Update user_progress table
}
```

**Code Generation:**
```bash
protoc --go_out=. --go-grpc_out=. proto/manga.proto
# Generates: proto/manga.pb.go, proto/manga_grpc.pb.go
```

**Testing:**
```bash
# Terminal 1
./mangahub.exe server start

# Terminal 2 (Install grpcurl first)
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Test GetManga
grpcurl -plaintext -d '{"manga_id":"manga_1"}' \
    localhost:9092 mangahub.MangaService/GetManga

# Test SearchManga
grpcurl -plaintext -d '{"query":"OnePiece"}' \
    localhost:9092 mangahub.MangaService/SearchManga
```

---

### 2.4: WebSocket Chat Server (Port 9093)

**What it does:**
- Real-time bidirectional communication (unlike HTTP)
- Multiple users can join/leave chat rooms
- All connected users receive all messages
- Full-duplex (send and receive simultaneously)

**Architecture:**
```
HTTP Listener (8080)
    ↓
GET /ws?username=alice
    ├─→ Upgrade HTTP → WebSocket
    ├─→ Create Client struct
    └─→ Register in Hub

Hub (Central Message Router)
    ├─→ Channel: register (new client)
    ├─→ Channel: unregister (disconnected client)
    ├─→ Channel: broadcast (message to send)
    └─→ run() goroutine processes all events

Client (Per-Connection Handler)
    ├─→ readPump() goroutine: Read from connection → send to hub.broadcast
    └─→ writePump() goroutine: Receive from client.send → write to connection
```

**Key Components:**
```go
type Hub struct {
    clients    map[*Client]bool     // All connected clients
    broadcast  chan *Message        // Messages to broadcast
    register   chan *Client         // New client to add
    unregister chan *Client         // Client to remove
    mu         sync.RWMutex
}

type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    send     chan *Message          // Only this client receives
    username string
}

type Message struct {
    Username  string
    Text      string
    Timestamp time.Time
}
```

**Flow Diagram:**
```
Client A                          Hub                         Client B
   ↓                              ↓                             ↓
readPump                          run()                       readPump
   ↓                              ↓                             ↓
ReadJSON("hello")             select {                     ReadJSON(...)
   ↓                          case msg := <-broadcast:         ↓
broadcast <- msg                  for client in clients:    writePump
   ↓                              client.send <- msg        write("hello")
(hub multiplexes)                 ↓                         (Client B sees!)
                              writePump
                              write("hello")
                              (Client A sees!)
```

**Testing:**
```bash
# Terminal 1
./mangahub.exe server start

# Terminal 2 (Install wscat)
npm install -g wscat

# Connect as alice
wscat -c ws://localhost:9093?username=alice

# Terminal 3
wscat -c ws://localhost:9093?username=bob

# Terminal 2: Send message
> {"text": "hello bob!"}

# Terminal 3 receives:
< {"username":"alice","text":"hello bob!","timestamp":"..."}
```

---

## 🔄 Modified Files for Phase 2

### cmd/server.go (Updated)

**Before (Phase 1):**
```go
var serverStartCmd = &cobra.Command{
    Run: func(cmd *cobra.Command, args []string) {
        database.Init("./mangahub.db")
        http_server.Start("8080")  // Blocks here
    },
}
```

**After (Phase 2):**
```go
var serverStartCmd = &cobra.Command{
    Run: func(cmd *cobra.Command, args []string) {
        database.Init("./mangahub.db")
        
        // Launch all 5 servers concurrently
        go http_server.Start("8080")
        go tcp_server.Start("9090")
        go udp_server.Start("9091")
        go grpc_server.Start("9092")
        go websocket_server.Start("9093")
        
        // Block main goroutine so servers keep running
        select {}
    },
}
```

**Key Changes:**
- `go` keyword launches each server in background goroutine
- `select {}` blocks forever (keeps program running)
- All 5 servers run concurrently
- Error from one server doesn't crash others

---

## 🚨 Common Gotchas in Phase 2

| Issue | Cause | Solution |
|-------|-------|----------|
| **Data race on clients map** | Multiple goroutines access without lock | Use `sync.Mutex` |
| **Goroutine leaks** | Goroutines never exit | Ensure proper cleanup in `defer` |
| **Deadlock on channels** | Sender blocks, no receiver | Use buffered channels or non-blocking select |
| **protoc not found** | Protocol Buffer compiler not installed | `brew install protobuf` |
| **WebSocket upgrade fails** | Forgot CORS headers | Set `Upgrader.CheckOrigin` |
| **TCP broadcasts never arrive** | Forgot to unlock mutex | Always use `defer mu.Unlock()` |

---

## 📚 Go Concepts in Phase 2

### Goroutines
```go
// Lightweight green threads managed by Go runtime
go functionName()  // Returns immediately
// functionName() runs in background
```

### Channels
```go
// Safe communication between goroutines
ch := make(chan string)       // Unbuffered (blocking)
ch := make(chan string, 10)   // Buffered (10 items max)

ch <- "msg"                   // Send
msg := <-ch                   // Receive
```

### Select
```go
// Multiplex multiple channels
select {
case msg := <-broadcast:     // Ready when msg available
    // Handle message
case client := <-register:   // Ready when client available
    // Handle registration
default:                      // No channels ready
    // Handle timeout
}
```

### Mutex
```go
// Protect shared data from race conditions
mu.Lock()
sharedMap[key] = value        // Safe: only one goroutine
mu.Unlock()
```

### Context
```go
// Cancellation & timeout for RPC calls
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
result, err := client.GetManga(ctx, req)
```

---

## ✅ Phase 2 Completion Checklist

### TCP Server (9090)
- [ ] Create `internal/tcp_server/server.go`
- [ ] Create `internal/tcp_server/hub.go`
- [ ] Implement `Start(port string)`
- [ ] Implement `handleConnection(conn, hub)`
- [ ] Implement `broadcastLoop(hub)`
- [ ] Test with telnet/netcat
- [ ] Verify no goroutine leaks

### UDP Server (9091)
- [ ] Create `internal/udp_server/server.go`
- [ ] Implement `Start(port string)`
- [ ] Parse REGISTER messages
- [ ] Parse NOTIFY messages
- [ ] Broadcast to registered clients
- [ ] Test with nc -u
- [ ] Handle concurrent clients

### gRPC Server (9092)
- [ ] Create `proto/manga.proto`
- [ ] Run protoc to generate code
- [ ] Create `internal/grpc_server/server.go`
- [ ] Implement `GetManga` RPC
- [ ] Implement `SearchManga` RPC
- [ ] Implement `UpdateProgress` RPC
- [ ] Create `internal/manga/service.go`
- [ ] Test with grpcurl
- [ ] Verify context timeouts work

### WebSocket Server (9093)
- [ ] Create `internal/websocket_server/server.go`
- [ ] Create `internal/websocket_server/hub.go`
- [ ] Create `internal/websocket_server/client.go`
- [ ] Implement hub `run()` with select
- [ ] Implement client `readPump()`
- [ ] Implement client `writePump()`
- [ ] HTTP GET `/ws` endpoint
- [ ] Test with wscat
- [ ] Verify multi-client messaging

### Concurrent Orchestration
- [ ] Update `cmd/server.go` with 5 concurrent servers
- [ ] Use `go` keyword for each server
- [ ] Use `select {}` to block main
- [ ] Test all 5 servers start
- [ ] Test all 5 ports listening: 8080, 9090, 9091, 9092, 9093
- [ ] Test one server crashing doesn't kill others

---

## 🔄 Phase 2 → Phase 3 Transition

Once Phase 2 is complete and all servers are working:

**Phase 3 will add CLI clients that connect to these servers:**

- `cmd/sync.go` → TCP client: `mangahub sync connect`
- `cmd/chat.go` → WebSocket client: `mangahub chat join`
- `cmd/manga.go` → gRPC client: `mangahub manga search`
- `cmd/notifications.go` → UDP client: `mangahub notifications listen`

