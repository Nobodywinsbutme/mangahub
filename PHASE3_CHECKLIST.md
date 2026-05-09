# MangaHub - Phase 3: CLI Client Integration Checklist

**Status:** ⏳ Starting Phase 3  
**Objective:** Connect CLI commands to Phase 2 servers  
**Timeline:** Weeks 6-8  
**Learning Focus:** Client-side networking, context management, error handling

---

## 📋 Phase 3 Overview

Phase 3 integrates the Phase 2 backend servers with CLI client commands. Users will now interact with the backend through CLI commands that connect to TCP, UDP, gRPC, and WebSocket servers.

### Command Flow (Phase 3 Goal)

```
User runs CLI command
    ↓
CLI command handler (in cmd/)
    ↓
Client package connects to Phase 2 server
    ↓
Send/receive data
    ↓
Display result to user
```

---

## 🎯 Phase 3A: TCP Sync Client

### Purpose
- Connect to TCP Sync Server (9090)
- Send reading progress updates
- Receive real-time progress from other users

### Checklist

#### 3A.1: Create TCP Client Package

- [ ] **Create file:** `internal/tcp_client/client.go`
  
  **Pseudocode:**
  ```go
  package tcp_client
  
  type Client struct {
      conn net.Conn
      addr string  // "localhost:9090"
  }
  
  func Connect(host string, port string) (*Client, error) {
      // 1. Dial TCP connection
      // 2. Return client or error
  }
  
  func (c *Client) SendProgress(userID, mangaID string, chapter int) error {
      // 1. Format message: user_id|manga_id|chapter
      // 2. Write to connection
      // 3. Handle errors
  }
  
  func (c *Client) Close() error {
      // Close connection cleanly
  }
  ```

- [ ] **Key Go Concept - Dialing TCP:**
  ```go
  conn, err := net.Dial("tcp", "localhost:9090")
  if err != nil {
      log.Fatal("Failed to connect:", err)
  }
  defer conn.Close()
  ```

#### 3A.2: Implement Listen Loop

- [ ] **Create function:** `ListenForUpdates(client *Client, handler func(string) error)`
  
  **What it does:**
  - Run in goroutine
  - Continuously read from server
  - Call handler function for each message
  
  **Pseudocode:**
  ```go
  func ListenForUpdates(client *Client, handler func(string) error) error {
      scanner := bufio.NewScanner(client.conn)
      for scanner.Scan() {
          msg := scanner.Text()
          if err := handler(msg); err != nil {
              return err
          }
      }
      return scanner.Err()
  }
  ```

#### 3A.3: Create CLI Command

- [ ] **Create file:** `cmd/sync.go`
  
  **New command:** `mangahub sync connect`
  
  **Pseudocode:**
  ```go
  var syncConnectCmd = &cobra.Command{
      Use:   "connect",
      Short: "Connect to real-time sync server",
      Run: func(cmd *cobra.Command, args []string) {
          client, err := tcp_client.Connect("localhost", "9090")
          if err != nil {
              log.Fatalf("Failed to connect: %v", err)
          }
          defer client.Close()
          
          fmt.Println("Connected to sync server. Listening for updates...")
          
          err = tcp_client.ListenForUpdates(client, func(msg string) error {
              fmt.Println("Update:", msg)
              return nil
          })
          if err != nil {
              log.Fatalf("Listen error: %v", err)
          }
      },
  }
  ```

#### 3A.4: Integrate with Progress Update Command

- [ ] **Modify:** `cmd/progress.go` (new file for progress commands)
  
  **Command:** `mangahub progress update --manga-id <id> --chapter <num>`
  
  **Pseudocode:**
  ```go
  var progressUpdateCmd = &cobra.Command{
      Use:   "update",
      Short: "Update reading progress",
      Run: func(cmd *cobra.Command, args []string) {
          mangaID, _ := cmd.Flags().GetString("manga-id")
          chapter, _ := cmd.Flags().GetInt("chapter")
          
          database.Init("./mangahub.db")
          
          // Get current user (from JWT token or config)
          userID := "usr_123"  // TODO: Extract from token
          
          // 1. Update database
          err := updateProgressInDB(userID, mangaID, chapter)
          if err != nil {
              log.Fatalf("DB update failed: %v", err)
          }
          
          // 2. Broadcast to TCP sync server
          client, err := tcp_client.Connect("localhost", "9090")
          if err != nil {
              log.Fatalf("Failed to connect to sync server: %v", err)
          }
          defer client.Close()
          
          err = client.SendProgress(userID, mangaID, chapter)
          if err != nil {
              log.Fatalf("Failed to send progress: %v", err)
          }
          
          fmt.Printf("✓ Progress updated: %s - Chapter %d\n", mangaID, chapter)
      },
  }
  ```

### Testing TCP Client

```bash
# Terminal 1: Start server
./mangahub.exe server start

# Terminal 2: CLI client connects and listens
./mangahub.exe sync connect

# Terminal 3: Another user updates progress
./mangahub.exe progress update --manga-id manga_1 --chapter 50

# Terminal 2: Should receive: "Update: usr_456|manga_1|50"
```

---

## 🎯 Phase 3B: UDP Notification Client

### Purpose
- Register with UDP Notification Server (9091)
- Receive chapter release notifications

### Checklist

#### 3B.1: Create UDP Client Package

- [ ] **Create file:** `internal/udp_client/client.go`
  
  **Pseudocode:**
  ```go
  package udp_client
  
  type Client struct {
      conn *net.UDPConn
      addr *net.UDPAddr
  }
  
  func Connect(host string, port string) (*Client, error) {
      // 1. Resolve server address
      // 2. Create UDP connection
      // 3. Return client
  }
  
  func (c *Client) Register(username string) error {
      // Send "REGISTER|username" message
  }
  
  func (c *Client) ListenForNotifications(handler func(string) error) error {
      // Listen for incoming notifications
  }
  ```

#### 3B.2: Implement Notification Listener

- [ ] **Create function:** `ListenForNotifications(client *Client, handler func(string) error)`
  
  **Pseudocode:**
  ```go
  func ListenForNotifications(client *Client, handler func(string) error) error {
      buffer := make([]byte, 1024)
      for {
          n, _, err := client.conn.ReadFromUDP(buffer)
          if err != nil {
              return err
          }
          
          msg := string(buffer[:n])
          if err := handler(msg); err != nil {
              return err
          }
      }
  }
  ```

#### 3B.3: Create CLI Command

- [ ] **Create file:** `cmd/notifications.go`
  
  **New command:** `mangahub notifications listen`
  
  **Pseudocode:**
  ```go
  var notificationsListenCmd = &cobra.Command{
      Use:   "listen",
      Short: "Listen for chapter release notifications",
      Run: func(cmd *cobra.Command, args []string) {
          username, _ := cmd.Flags().GetString("username")
          
          client, err := udp_client.Connect("localhost", "9091")
          if err != nil {
              log.Fatalf("Failed to connect: %v", err)
          }
          defer client.conn.Close()
          
          err = client.Register(username)
          if err != nil {
              log.Fatalf("Failed to register: %v", err)
          }
          
          fmt.Printf("Listening for notifications (%s)...\n", username)
          
          err = client.ListenForNotifications(func(msg string) error {
              fmt.Println("🔔", msg)
              return nil
          })
          if err != nil {
              log.Fatalf("Listen error: %v", err)
          }
      },
  }
  ```

### Testing UDP Client

```bash
# Terminal 1: Start server
./mangahub.exe server start

# Terminal 2: Listen for notifications
./mangahub.exe notifications listen --username alice

# Terminal 3: Send notification
./mangahub.exe notifications send --title "OnePiece" --chapter "1050"

# Terminal 2: Should receive: "New Chapter: OnePiece - Chapter 1050"
```

---

## 🎯 Phase 3C: gRPC Client Integration

### Purpose
- Connect to gRPC Manga Service (9092)
- Query manga information
- Search manga by title
- Update reading progress

### Checklist

#### 3C.1: Create gRPC Client Wrapper

- [ ] **Create file:** `internal/grpc_client/client.go`
  
  **Pseudocode:**
  ```go
  package grpc_client
  
  type Client struct {
      conn   *grpc.ClientConn
      client pb.MangaServiceClient
  }
  
  func Connect(host string, port string) (*Client, error) {
      // 1. Dial gRPC server
      // 2. Create client stub
      // 3. Return client
  }
  
  func (c *Client) GetManga(ctx context.Context, mangaID string) (*pb.GetMangaResponse, error) {
      req := &pb.GetMangaRequest{MangaId: mangaID}
      return c.client.GetManga(ctx, req)
  }
  
  func (c *Client) SearchManga(ctx context.Context, query string) (*pb.SearchMangaResponse, error) {
      req := &pb.SearchMangaRequest{Query: query}
      return c.client.SearchManga(ctx, req)
  }
  
  func (c *Client) UpdateProgress(ctx context.Context, userID, mangaID string, chapter int) (*pb.UpdateProgressResponse, error) {
      req := &pb.UpdateProgressRequest{
          UserId: userID,
          MangaId: mangaID,
          CurrentChapter: int32(chapter),
      }
      return c.client.UpdateProgress(ctx, req)
  }
  
  func (c *Client) Close() error {
      return c.conn.Close()
  }
  ```

- [ ] **Key Go Concept - Context:**
  ```go
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()
  
  // Pass ctx to gRPC call
  result, err := c.client.GetManga(ctx, req)
  ```

#### 3C.2: Create Manga Commands

- [ ] **Create file:** `cmd/manga.go`
  
  **Commands:**
  - `mangahub manga search <query>`
  - `mangahub manga info <manga-id>`
  
  **Pseudocode:**
  ```go
  var mangaSearchCmd = &cobra.Command{
      Use:   "search",
      Short: "Search manga by title",
      Args:  cobra.ExactArgs(1),  // Require exactly 1 argument
      Run: func(cmd *cobra.Command, args []string) {
          query := args[0]
          
          ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
          defer cancel()
          
          client, err := grpc_client.Connect("localhost", "9092")
          if err != nil {
              log.Fatalf("Failed to connect: %v", err)
          }
          defer client.Close()
          
          resp, err := client.SearchManga(ctx, query)
          if err != nil {
              log.Fatalf("Search failed: %v", err)
          }
          
          for _, manga := range resp.Results {
              fmt.Printf("- %s by %s (Ch. %d)\n", manga.Title, manga.Author, manga.TotalChapters)
          }
      },
  }
  
  var mangaInfoCmd = &cobra.Command{
      Use:   "info",
      Short: "Get manga information",
      Args:  cobra.ExactArgs(1),
      Run: func(cmd *cobra.Command, args []string) {
          mangaID := args[0]
          
          ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
          defer cancel()
          
          client, err := grpc_client.Connect("localhost", "9092")
          if err != nil {
              log.Fatalf("Failed to connect: %v", err)
          }
          defer client.Close()
          
          resp, err := client.GetManga(ctx, mangaID)
          if err != nil {
              log.Fatalf("Get failed: %v", err)
          }
          
          fmt.Printf("Title: %s\nAuthor: %s\nChapters: %d\nStatus: %s\n",
              resp.Title, resp.Author, resp.TotalChapters, resp.Status)
      },
  }
  ```

#### 3C.3: Integrate with Library Commands

- [ ] **Create file:** `cmd/library.go`
  
  **Commands:**
  - `mangahub library add --manga-id <id> --status <status>`
  - `mangahub library list`
  - `mangahub library remove --manga-id <id>`
  
  **Pseudocode:**
  ```go
  var libraryAddCmd = &cobra.Command{
      Use:   "add",
      Short: "Add manga to library",
      Run: func(cmd *cobra.Command, args []string) {
          mangaID, _ := cmd.Flags().GetString("manga-id")
          status, _ := cmd.Flags().GetString("status")
          
          database.Init("./mangahub.db")
          userID := "usr_123"  // TODO: From token
          
          // Use gRPC to update progress
          ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
          defer cancel()
          
          client, err := grpc_client.Connect("localhost", "9092")
          if err != nil {
              log.Fatalf("Failed to connect: %v", err)
          }
          defer client.Close()
          
          resp, err := client.UpdateProgress(ctx, userID, mangaID, 0)
          if err != nil {
              log.Fatalf("Add failed: %v", err)
          }
          
          fmt.Printf("✓ %s\n", resp.Message)
      },
  }
  ```

### Testing gRPC Client

```bash
# Terminal 1: Start server
./mangahub.exe server start

# Terminal 2: Search manga
./mangahub.exe manga search "OnePiece"

# Terminal 2: Get manga info
./mangahub.exe manga info "manga_1"

# Terminal 2: Add to library
./mangahub.exe library add --manga-id manga_1 --status reading
```

---

## 🎯 Phase 3D: WebSocket Chat Client

### Purpose
- Connect to WebSocket Chat Server (9093)
- Send and receive messages in real-time
- Interactive chat interface

### Checklist

#### 3D.1: Create WebSocket Client Package

- [ ] **Create file:** `internal/websocket_client/client.go`
  
  **Pseudocode:**
  ```go
  package websocket_client
  
  type Client struct {
      conn *websocket.Conn
      send chan *Message
  }
  
  type Message struct {
      Username  string
      Text      string
      Timestamp time.Time
  }
  
  func Connect(url string, username string) (*Client, error) {
      // 1. Dial WebSocket
      // 2. Create client struct
      // 3. Launch read/write goroutines
      // 4. Return client
  }
  
  func (c *Client) SendMessage(text string) error {
      // Send message to server
  }
  
  func (c *Client) ListenForMessages(handler func(*Message) error) error {
      // Receive messages from server
  }
  ```

#### 3D.2: Implement Read/Write Loops

- [ ] **Create functions:** `readPump()` and `writePump()`
  
  **Pseudocode:**
  ```go
  func (c *Client) readPump(handler func(*Message) error) error {
      defer c.conn.Close()
      
      for {
          var msg Message
          err := c.conn.ReadJSON(&msg)
          if err != nil {
              return err
          }
          
          if err := handler(&msg); err != nil {
              return err
          }
      }
  }
  
  func (c *Client) writePump() error {
      for {
          msg := <-c.send
          err := c.conn.WriteJSON(msg)
          if err != nil {
              return err
          }
      }
  }
  ```

#### 3D.3: Create Interactive CLI Command

- [ ] **Create file:** `cmd/chat.go`
  
  **New command:** `mangahub chat join`
  
  **Pseudocode:**
  ```go
  var chatJoinCmd = &cobra.Command{
      Use:   "join",
      Short: "Join manga chat room",
      Run: func(cmd *cobra.Command, args []string) {
          username, _ := cmd.Flags().GetString("username")
          
          url := fmt.Sprintf("ws://localhost:9093?username=%s", username)
          client, err := websocket_client.Connect(url, username)
          if err != nil {
              log.Fatalf("Failed to connect: %v", err)
          }
          defer client.conn.Close()
          
          fmt.Printf("Joined chat as %s. Type messages:\n", username)
          
          // Launch goroutine to read messages from server
          go func() {
              err := client.ListenForMessages(func(msg *websocket_client.Message) error {
                  fmt.Printf("[%s] %s: %s\n", msg.Timestamp.Format("15:04:05"), msg.Username, msg.Text)
                  return nil
              })
              if err != nil {
                  log.Printf("Read error: %v", err)
              }
          }()
          
          // Main goroutine: read from stdin and send
          scanner := bufio.NewScanner(os.Stdin)
          for scanner.Scan() {
              text := scanner.Text()
              err := client.SendMessage(text)
              if err != nil {
                  log.Fatalf("Send error: %v", err)
              }
          }
      },
  }
  ```

#### 3D.4: Handle Graceful Disconnect

- [ ] **Implement:** Clean connection close
  ```go
  defer func() {
      client.conn.WriteMessage(
          websocket.CloseMessage,
          websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
      )
      client.conn.Close()
  }()
  ```

### Testing WebSocket Client

```bash
# Terminal 1: Start server
./mangahub.exe server start

# Terminal 2: User Alice joins chat
./mangahub.exe chat join --username alice

# Terminal 3: User Bob joins chat
./mangahub.exe chat join --username bob

# Terminal 2: Type message
hello everyone

# Terminal 3: Receives: "[10:30:45] alice: hello everyone"

# Terminal 3: Type message
hey alice!

# Terminal 2: Receives: "[10:31:00] bob: hey alice!"
```

---

## 📋 Phase 3 Complete Checklist

### TCP Sync Client (9090)
- [ ] Create `internal/tcp_client/client.go`
- [ ] Implement `Connect(host, port)` function
- [ ] Implement `SendProgress()` function
- [ ] Implement `ListenForUpdates()` function
- [ ] Create `cmd/sync.go` with `sync connect` command
- [ ] Modify progress update to send to TCP server
- [ ] Test with multiple clients

### UDP Notification Client (9091)
- [ ] Create `internal/udp_client/client.go`
- [ ] Implement `Connect(host, port)` function
- [ ] Implement `Register()` function
- [ ] Implement `ListenForNotifications()` function
- [ ] Create `cmd/notifications.go` with `notifications listen` command
- [ ] Create `notifications send` command (admin)
- [ ] Test notifications arrive at registered clients

### gRPC Client (9092)
- [ ] Create `internal/grpc_client/client.go` wrapper
- [ ] Implement `GetManga()` function
- [ ] Implement `SearchManga()` function
- [ ] Implement `UpdateProgress()` function
- [ ] Create `cmd/manga.go` with `search` and `info` commands
- [ ] Create `cmd/library.go` with `add`, `list`, `remove` commands
- [ ] Integrate library commands with gRPC
- [ ] Test search, info, and library management

### WebSocket Chat Client (9093)
- [ ] Create `internal/websocket_client/client.go`
- [ ] Implement `Connect()` function
- [ ] Implement `SendMessage()` function
- [ ] Implement `ListenForMessages()` function
- [ ] Implement `readPump()` and `writePump()` goroutines
- [ ] Create `cmd/chat.go` with `chat join` command
- [ ] Implement interactive stdin for user input
- [ ] Handle graceful disconnect
- [ ] Test multi-user chat

### CLI Command Hierarchy
- [ ] `mangahub auth register/login` (existing)
- [ ] `mangahub server start` (existing, updated to run all 5 servers)
- [ ] `mangahub sync connect` (new: TCP client)
- [ ] `mangahub notifications listen` (new: UDP client)
- [ ] `mangahub manga search <query>` (new: gRPC client)
- [ ] `mangahub manga info <id>` (new: gRPC client)
- [ ] `mangahub library add/list/remove` (new: gRPC client)
- [ ] `mangahub chat join --username <name>` (new: WebSocket client)
- [ ] `mangahub progress update --manga-id <id> --chapter <num>` (new: HTTP + TCP)

---

## 🚨 Common Pitfalls in Phase 3

| Issue | Solution |
|-------|----------|
| **Connection timeout to server** | Ensure server is running, check firewall |
| **gRPC "unimplemented" error** | Ensure protoc generated code, rebuild |
| **WebSocket connection refused** | Verify URL format, check CORS if needed |
| **TCP message format mismatch** | Match message format between server and client |
| **Goroutine leaks in CLI** | Ensure all goroutines exit (use `defer`, `cancel`) |
| **Context timeout too short** | Increase timeout for slow operations |
| **UDP messages not arriving** | UDP is lossy; add retries or TCP fallback |

---

## 📚 Go Concepts to Master in Phase 3

### 1. Dialing Connections (Client Side)
```go
conn, err := net.Dial("tcp", "localhost:9090")
if err != nil {
    log.Fatal("Connection failed:", err)
}
defer conn.Close()
```

### 2. gRPC Client Stubs
```go
conn, _ := grpc.Dial("localhost:9092")
client := pb.NewMangaServiceClient(conn)

resp, err := client.GetManga(ctx, &pb.GetMangaRequest{...})
```

### 3. WebSocket Upgrade (Client)
```go
url := "ws://localhost:9093"
ws, _, err := websocket.DefaultDialer.Dial(url, nil)

err = ws.WriteJSON(&Message{Text: "hello"})
```

### 4. Context for Cancellation
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// If operation takes >5s, it will be cancelled
```

### 5. Buffered Input/Output
```go
scanner := bufio.NewScanner(os.Stdin)
for scanner.Scan() {
    line := scanner.Text()
    // Process line
}
```

---

## ✅ Phase 3 Success Criteria

You'll know Phase 3 is complete when:

✅ All CLI commands work without errors  
✅ TCP sync client receives broadcasts from server  
✅ UDP client receives notifications  
✅ gRPC client queries return correct data  
✅ WebSocket chat works interactively  
✅ Progress updates sync across clients  
✅ No goroutine leaks  
✅ All 5 servers and 5 clients work together

---

## 🔄 Phase 3 → Phase 4 Transition

Once Phase 3 is complete, Phase 4 will add:
- Error recovery & reconnection logic
- Configuration file for JWT secrets
- Redis caching for manga data
- User reviews & ratings system
- Advanced filtering and sorting
- Load testing & performance tuning

