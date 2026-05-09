# MangaHub - Phase 2: Protocol Implementation Checklist

**Status:** ⏳ Starting Phase 2  
**Objective:** Build 5 concurrent servers with distinct protocols (TCP, UDP, gRPC, WebSocket)  
**Timeline:** Weeks 3-5  
**Learning Focus:** Goroutines, channels, protocol buffers, socket programming

---

## 📋 Phase 2 Overview

Phase 2 transforms the single HTTP server from Phase 1 into a **multi-protocol backend**. All 5 servers must launch concurrently when running `mangahub server start`.

### Architecture Diagram

```
mangahub server start
    ↓
database.Init("./mangahub.db")
    ↓
    ├─→ [Goroutine 1] http_server.Start("8080")       ✓ Exists from Phase 1
    ├─→ [Goroutine 2] tcp_server.Start("9090")        🔨 Build in Phase 2
    ├─→ [Goroutine 3] udp_server.Start("9091")        🔨 Build in Phase 2
    ├─→ [Goroutine 4] grpc_server.Start("9092")       🔨 Build in Phase 2
    └─→ [Goroutine 5] websocket_server.Start("9093")  🔨 Build in Phase 2
```

### Concurrent Model (New Go Concepts)

```go
go http_server.Start("8080")       // Launch in goroutine (doesn't block)
go tcp_server.Start("9090")        // Launch in goroutine (doesn't block)
go udp_server.Start("9091")        // Launch in goroutine (doesn't block)
go grpc_server.Start("9092")       // Launch in goroutine (doesn't block)
go websocket_server.Start("9093")  // Launch in goroutine (doesn't block)

// Main goroutine waits so program doesn't exit
select {}  // Block forever (all servers run in background)
```

---

## 🎯 Phase 2A: TCP Sync Server (Port 9090)

### Purpose
- Real-time synchronization of reading progress across all connected clients
- When user updates progress, broadcast to all listeners
- Persistent connection (unlike HTTP request/response)

### Checklist

#### 2A.1: Create TCP Server Structure

- [ ] **Create file:** `internal/tcp_server/server.go`
  
  **Pseudocode needed:**
  ```
  package tcp_server
  
  type Server struct {
      listener net.Listener
      clients  map[net.Conn]bool  // Track connected clients
      mu       sync.Mutex         // Protect concurrent map access
      broadcast chan []byte       // Channel for messages to send
  }
  
  func Start(port string) {
      // 1. Create TCP listener on :9090
      // 2. Accept connections in loop
      // 3. For each connection, spawn goroutine to handle it
      // 4. Broadcast messages to all connected clients
  }
  ```

- [ ] **Key Go Concept - Goroutines:**
  ```go
  go handleConnection(conn)  // Launch handler in new goroutine
  // Returns immediately; handler runs in background
  ```

- [ ] **Key Go Concept - Channels:**
  ```go
  broadcast := make(chan []byte)  // Channel to pass messages
  broadcast <- message            // Send message into channel
  msg := <-broadcast              // Receive message from channel
  ```

- [ ] **Key Go Concept - Mutexes:**
  ```go
  mu.Lock()                       // Acquire lock (wait if locked)
  clients[conn] = true            // Modify shared map
  mu.Unlock()                     // Release lock
  ```

#### 2A.2: Implement Client Connection Handler

- [ ] **Create function:** `handleTCPConnection(conn net.Conn, server *Server)`
  
  **What it does:**
  - Register client in `server.clients` map
  - Listen for incoming messages from this connection
  - Parse message (format TBD)
  - Send message to broadcast channel
  - Clean up when connection closes

- [ ] **Message Format (ASCII Protocol):**
  ```
  USER_ID|MANGA_ID|CHAPTER_NUMBER\n
  Example: usr_123|manga_456|42\n
  ```
  
  **Pseudocode:**
  ```go
  func handleTCPConnection(conn net.Conn, server *Server) {
      defer conn.Close()
      
      // Register this client
      server.mu.Lock()
      server.clients[conn] = true
      server.mu.Unlock()
      
      scanner := bufio.NewScanner(conn)
      for scanner.Scan() {
          line := scanner.Text()
          // Parse: user_id|manga_id|chapter
          // Validate
          // Send to broadcast channel
          message := fmt.Sprintf("Progress: %s", line)
          server.broadcast <- []byte(message)
      }
      
      // Deregister on disconnect
      server.mu.Lock()
      delete(server.clients, conn)
      server.mu.Unlock()
  }
  ```

#### 2A.3: Implement Broadcast Loop

- [ ] **Create function:** `broadcastLoop(server *Server)`
  
  **What it does:**
  - Run in separate goroutine
  - Listen on `server.broadcast` channel
  - Send message to all connected clients
  
  **Pseudocode:**
  ```go
  func broadcastLoop(server *Server) {
      for {
          msg := <-server.broadcast  // Wait for message
          
          server.mu.Lock()
          for conn := range server.clients {
              conn.Write(msg)         // Send to each client
              conn.Write([]byte("\n"))
          }
          server.mu.Unlock()
      }
  }
  ```

#### 2A.4: Handle Graceful Shutdown

- [ ] **Close listener properly:**
  ```go
  listener.Close()
  ```

- [ ] **Close all client connections:**
  ```go
  for conn := range server.clients {
      conn.Close()
  }
  ```

### Testing TCP Server

```bash
# Terminal 1: Start server
./mangahub.exe server start

# Terminal 2: Connect as TCP client
telnet localhost 9090

# Type message
usr_123|manga_456|42

# Terminal 3: Connect another client
telnet localhost 9090
# Should receive broadcast from Terminal 2
```

---

## 🎯 Phase 2B: UDP Notification System (Port 9091)

### Purpose
- Lightweight, connectionless notifications
- Register clients with server (no persistent connection)
- Broadcast chapter release notifications to registered clients
- Fire-and-forget (unlike TCP's persistent connection)

### Checklist

#### 2B.1: Create UDP Server Structure

- [ ] **Create file:** `internal/udp_server/server.go`
  
  **Pseudocode:**
  ```
  package udp_server
  
  type Server struct {
      conn      *net.UDPConn
      clients   map[string]*net.UDPAddr  // username -> UDP address
      mu        sync.Mutex
  }
  
  func Start(port string) {
      // 1. Create UDP listener on :9091
      // 2. Receive UDP packets in loop
      // 3. Parse packet (register or notification)
      // 4. Broadcast to registered clients
  }
  ```

- [ ] **Key Go Concept - UDP:**
  ```go
  addr, _ := net.ResolveUDPAddr("udp", ":9091")
  conn, _ := net.ListenUDP("udp", addr)
  
  buffer := make([]byte, 1024)
  n, remoteAddr, _ := conn.ReadFromUDP(buffer)
  message := string(buffer[:n])
  ```

#### 2B.2: Implement Message Parsing

- [ ] **Message Format:**
  ```
  REGISTER|username\n
  NOTIFY|manga_title|chapter_number\n
  
  Example:
  REGISTER|alice
  NOTIFY|OnePiece|1050
  ```

- [ ] **Pseudocode:**
  ```go
  func handleUDPMessage(msg string, remoteAddr *net.UDPAddr, server *Server) {
      parts := strings.Split(msg, "|")
      
      if parts[0] == "REGISTER" {
          username := parts[1]
          server.mu.Lock()
          server.clients[username] = remoteAddr
          server.mu.Unlock()
      } else if parts[0] == "NOTIFY" {
          mangaTitle := parts[1]
          chapter := parts[2]
          broadcastNotification(server, mangaTitle, chapter)
      }
  }
  ```

#### 2B.3: Implement Broadcast Function

- [ ] **Create function:** `broadcastNotification(server *Server, mangaTitle string, chapter string)`
  
  **Pseudocode:**
  ```go
  func broadcastNotification(server *Server, mangaTitle, chapter string) {
      notification := fmt.Sprintf("New Chapter: %s - Chapter %s", mangaTitle, chapter)
      
      server.mu.Lock()
      for username, addr := range server.clients {
          server.conn.WriteToUDP([]byte(notification), addr)
      }
      server.mu.Unlock()
  }
  ```

#### 2B.4: Handle Client Timeout

- [ ] **Optional:** Remove stale client registrations after no activity
  ```go
  // Advanced: Implement TTL (time-to-live) for client registrations
  ```

### Testing UDP Server

```bash
# Terminal 1: Start server
./mangahub.exe server start

# Terminal 2: Register as client
echo "REGISTER|alice" | nc -u localhost 9091

# Terminal 3: Send notification
echo "NOTIFY|OnePiece|1050" | nc -u localhost 9091

# Terminal 2: Should receive notification
```

---

## 🎯 Phase 2C: gRPC Internal Service (Port 9092)

### Purpose
- Fast binary protocol for internal microservice communication
- Strongly typed messages (Protocol Buffers)
- Handles: `GetManga`, `SearchManga`, `UpdateProgress`

### Checklist

#### 2C.1: Install gRPC & Protocol Buffer Tools

- [ ] **Add to go.mod:**
  ```bash
  go get google.golang.org/grpc
  go get google.golang.org/protobuf
  ```

- [ ] **Install protoc compiler:**
  ```bash
  # Windows: choco install protoc
  # macOS: brew install protobuf
  # Linux: apt install protobuf-compiler
  ```

#### 2C.2: Define Protocol Buffers

- [ ] **Create file:** `proto/manga.proto`
  
  **Pseudocode:**
  ```protobuf
  syntax = "proto3";
  
  package mangahub;
  
  // Request: Get manga by ID
  message GetMangaRequest {
      string manga_id = 1;
  }
  
  // Response: Manga details
  message GetMangaResponse {
      string id = 1;
      string title = 2;
      string author = 3;
      int32 total_chapters = 4;
      string status = 5;  // "ongoing" | "completed"
  }
  
  // Request: Search manga by title
  message SearchMangaRequest {
      string query = 1;
  }
  
  // Response: List of manga
  message SearchMangaResponse {
      repeated GetMangaResponse results = 1;
  }
  
  // Request: Update reading progress
  message UpdateProgressRequest {
      string user_id = 1;
      string manga_id = 2;
      int32 current_chapter = 3;
      string status = 4;  // "reading" | "completed"
  }
  
  // Response: Confirmation
  message UpdateProgressResponse {
      bool success = 1;
      string message = 2;
  }
  
  // Define service
  service MangaService {
      rpc GetManga(GetMangaRequest) returns (GetMangaResponse);
      rpc SearchManga(SearchMangaRequest) returns (SearchMangaResponse);
      rpc UpdateProgress(UpdateProgressRequest) returns (UpdateProgressResponse);
  }
  ```

#### 2C.3: Generate Go Code from Proto

- [ ] **Run protoc:**
  ```bash
  protoc --go_out=. --go-grpc_out=. proto/manga.proto
  ```
  
  This generates:
  - `proto/manga.pb.go` (message structs)
  - `proto/manga_grpc.pb.go` (service interface)

#### 2C.4: Implement gRPC Server

- [ ] **Create file:** `internal/grpc_server/server.go`
  
  **Pseudocode:**
  ```go
  package grpc_server
  
  type MangaServer struct {
      // Implement protobuf MangaServiceServer interface
  }
  
  func (s *MangaServer) GetManga(ctx context.Context, req *pb.GetMangaRequest) (*pb.GetMangaResponse, error) {
      // 1. Query SQLite for manga by ID
      // 2. Return GetMangaResponse
  }
  
  func (s *MangaServer) SearchManga(ctx context.Context, req *pb.SearchMangaRequest) (*pb.SearchMangaResponse, error) {
      // 1. Query SQLite for manga matching query
      // 2. Return SearchMangaResponse with results
  }
  
  func (s *MangaServer) UpdateProgress(ctx context.Context, req *pb.UpdateProgressRequest) (*pb.UpdateProgressResponse, error) {
      // 1. Insert/update user_progress table
      // 2. Return UpdateProgressResponse with success
  }
  
  func Start(port string) {
      listener, _ := net.Listen("tcp", ":" + port)
      server := grpc.NewServer()
      pb.RegisterMangaServiceServer(server, &MangaServer{})
      server.Serve(listener)
  }
  ```

#### 2C.5: Add Database Queries for gRPC

- [ ] **Create file:** `internal/manga/service.go`
  
  **Implement:**
  ```go
  GetMangaByID(id string) (*models.Manga, error)
  SearchManga(query string) ([]*models.Manga, error)
  UpdateUserProgress(userID, mangaID string, chapter int, status string) error
  ```

### Testing gRPC Server

- [ ] **Use grpcurl tool:**
  ```bash
  # Install: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
  
  # Test GetManga
  grpcurl -plaintext \
    -d '{"manga_id":"manga_1"}' \
    localhost:9092 mangahub.MangaService/GetManga
  ```

---

## 🎯 Phase 2D: WebSocket Chat Server (Port 9093)

### Purpose
- Real-time bidirectional communication (unlike HTTP request/response)
- Multiple users can join chat rooms
- Broadcast messages to all connected users

### Checklist

#### 2D.1: Install WebSocket Library

- [ ] **Add to go.mod:**
  ```bash
  go get github.com/gorilla/websocket
  ```

#### 2D.2: Create WebSocket Server Structure

- [ ] **Create file:** `internal/websocket_server/server.go`
  
  **Pseudocode:**
  ```go
  package websocket_server
  
  type Hub struct {
      clients    map[*Client]bool
      broadcast  chan *Message
      register   chan *Client
      unregister chan *Client
      mu         sync.RWMutex
  }
  
  type Client struct {
      hub      *Hub
      conn     *websocket.Conn
      send     chan *Message
      username string
  }
  
  type Message struct {
      Username  string
      Text      string
      Timestamp time.Time
  }
  ```

#### 2D.3: Implement Hub (Central Message Router)

- [ ] **Create function:** `(h *Hub) run()`
  
  **Pseudocode:**
  ```go
  func (h *Hub) run() {
      for {
          select {
          case client := <-h.register:
              // Add client to hub
              h.mu.Lock()
              h.clients[client] = true
              h.mu.Unlock()
              
          case client := <-h.unregister:
              // Remove client from hub
              h.mu.Lock()
              delete(h.clients, client)
              h.mu.Unlock()
              
          case msg := <-h.broadcast:
              // Send message to all clients
              h.mu.RLock()
              for client := range h.clients {
                  client.send <- msg
              }
              h.mu.RUnlock()
          }
      }
  }
  ```

#### 2D.4: Implement Client Connection Handler

- [ ] **Create function:** `(c *Client) readPump()`
  
  **Pseudocode:**
  ```go
  func (c *Client) readPump() {
      defer func() {
          c.hub.unregister <- c
          c.conn.Close()
      }()
      
      for {
          var msg Message
          err := c.conn.ReadJSON(&msg)
          if err != nil {
              break
          }
          
          msg.Username = c.username
          msg.Timestamp = time.Now()
          c.hub.broadcast <- &msg
      }
  }
  ```

#### 2D.5: Implement Message Writer

- [ ] **Create function:** `(c *Client) writePump()`
  
  **Pseudocode:**
  ```go
  func (c *Client) writePump() {
      for {
          msg := <-c.send
          c.conn.WriteJSON(msg)
      }
  }
  ```

#### 2D.6: HTTP Handler for WebSocket Upgrade

- [ ] **Create endpoint:** `GET /ws` (upgrade to WebSocket)
  
  **Pseudocode:**
  ```go
  func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
      username := r.URL.Query().Get("username")
      
      upgrader := websocket.Upgrader{}
      conn, _ := upgrader.Upgrade(w, r, nil)
      
      client := &Client{
          hub:      hub,
          conn:     conn,
          send:     make(chan *Message),
          username: username,
      }
      
      hub.register <- client
      
      go client.readPump()
      go client.writePump()
  }
  ```

### Testing WebSocket Server

```bash
# Terminal 1: Start server
./mangahub.exe server start

# Terminal 2: Connect via WebSocket client (use online tool or wscat)
wscat -c ws://localhost:9093?username=alice

# Terminal 3: Connect another client
wscat -c ws://localhost:9093?username=bob

# Terminal 2: Send message
{"text": "Hello everyone!", ...}

# Terminal 3: Receives message
```

---

## 📋 Phase 2 Complete Checklist

### TCP Sync Server (9090)
- [ ] Create `internal/tcp_server/server.go`
- [ ] Implement `Start(port string)` function
- [ ] Implement `handleTCPConnection(conn net.Conn, server *Server)`
- [ ] Implement `broadcastLoop(server *Server)`
- [ ] Handle client registration/deregistration
- [ ] Handle graceful shutdown
- [ ] Test with telnet/nc

### UDP Notification System (9091)
- [ ] Create `internal/udp_server/server.go`
- [ ] Implement `Start(port string)` function
- [ ] Implement `handleUDPMessage(msg string, remoteAddr *net.UDPAddr, server *Server)`
- [ ] Implement `broadcastNotification(server *Server, title, chapter string)`
- [ ] Handle client registration
- [ ] Test with nc/netcat

### gRPC Internal Service (9092)
- [ ] Create `proto/manga.proto` with service definitions
- [ ] Run `protoc` to generate Go code
- [ ] Create `internal/grpc_server/server.go`
- [ ] Implement `GetManga` RPC handler
- [ ] Implement `SearchManga` RPC handler
- [ ] Implement `UpdateProgress` RPC handler
- [ ] Create `internal/manga/service.go` with database queries
- [ ] Test with grpcurl

### WebSocket Chat (9093)
- [ ] Create `internal/websocket_server/server.go`
- [ ] Implement `Hub` struct and `run()` method
- [ ] Implement `Client` struct
- [ ] Implement `readPump()` and `writePump()`
- [ ] Create HTTP `GET /ws` endpoint with WebSocket upgrade
- [ ] Handle client registration/deregistration
- [ ] Broadcast messages to all clients
- [ ] Test with wscat/online WebSocket client

### Server Orchestration (cmd/server.go)
- [ ] Modify `cmd/server.go` to launch 5 servers concurrently
- [ ] Use goroutines (`go server.Start()`)
- [ ] Use `select {}` to block main goroutine
- [ ] Test all 5 servers start simultaneously
- [ ] Test all 5 ports are listening (9090-9093 + 8080)

---

## 🚨 Common Pitfalls in Phase 2

| Issue | Solution |
|-------|----------|
| **Data races (concurrent map access)** | Use `sync.Mutex` to lock shared data |
| **Goroutine leaks (goroutines never exit)** | Ensure proper cleanup in `defer` statements |
| **Deadlocks (channels blocking forever)** | Use non-blocking `select` with `default` case |
| **Unbuffered channels blocking** | Use buffered channels: `make(chan T, size)` |
| **Forgetting to upgrade HTTP to WebSocket** | Use `websocket.Upgrader` in HTTP handler |
| **Protocol Buffer code not generated** | Run `protoc` before building |

---

## 📚 Go Concepts to Master in Phase 2

### 1. Goroutines (Lightweight Threads)
```go
go functionName()  // Launch in background
// Main goroutine continues immediately (doesn't wait)
```

### 2. Channels (Inter-Goroutine Communication)
```go
ch := make(chan string)       // Unbuffered (sender waits for receiver)
ch := make(chan string, 10)   // Buffered (holds 10 items)

ch <- "message"               // Send
msg := <-ch                   // Receive
```

### 3. Select (Multiplexing Channels)
```go
select {
case msg := <-broadcast:
    // Handle message
case client := <-register:
    // Handle registration
default:
    // No channel ready
}
```

### 4. Mutex (Protecting Shared Data)
```go
mu.Lock()
sharedMap[key] = value  // Safe: only one goroutine accesses
mu.Unlock()
```

### 5. Context (Cancellation & Timeouts)
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Pass ctx to functions that might block
result, err := someFunction(ctx)
```

---

## ✅ Phase 2 Success Criteria

You'll know Phase 2 is complete when:

✅ All 5 servers start simultaneously  
✅ TCP server accepts multiple connections and broadcasts messages  
✅ UDP server registers clients and broadcasts notifications  
✅ gRPC server serves manga queries and updates  
✅ WebSocket server handles multi-client chat  
✅ No race conditions or deadlocks (test with `go run -race`)  
✅ All servers shut down gracefully  
✅ Each protocol works independently

---

## 🔄 Phase 2 → Phase 3 Transition

Once Phase 2 is complete, Phase 3 will:
- Connect CLI clients to these servers
- Implement `mangahub sync connect` (TCP client)
- Implement `mangahub chat join` (WebSocket client)
- Implement `mangahub library add` (gRPC client)
- Add CLI commands for UDP notifications

