# MangaHub - Manga & Comic Tracking System

**Course:** Net-centric Programming (IT096IU)  
**Instructors:** Lê Thanh Sơn - Nguyễn Trung Nghĩa

## 👥 Team Members (Group 25)

- Lê Đoàn Minh Ngọc (ID: ITITDK23023)
- Phạm Trung Kiên (ID: ITCSIU23020)

---

## 📖 Project Overview

MangaHub is a **command-line interface (CLI) manga tracking system** built in Go. It demonstrates advanced network programming concepts by integrating **five distinct communication protocols** into a unified backend system.

### Core Features

- **User Authentication:** Register/login with bcrypt hashing and JWT tokens
- **Real-time Progress Sync:** TCP server broadcasts reading progress to all connected clients
- **Chapter Notifications:** UDP server sends new chapter alerts to registered users
- **Manga Search & Library:** gRPC internal service for fast queries and operations
- **Community Chat:** WebSocket-based real-time chat for discussing manga
- **Multi-Protocol Architecture:** HTTP (REST), TCP (raw sockets), UDP (stateless), gRPC (RPC), WebSocket (bidirectional)

---

## 🏛️ System Architecture

### Network Protocols Implemented

| Protocol | Port | Purpose | Use Case |
|----------|------|---------|----------|
| **HTTP (REST)** | 8080 | Authentication, CRUD operations | User login, manga updates via CLI |
| **TCP** | 9090 | Real-time progress synchronization | Broadcast reading progress to all clients |
| **UDP** | 9091 | Lightweight notifications | Push chapter releases to subscribers |
| **gRPC** | 9092 | Fast internal RPC calls | Manga queries, library management |
| **WebSocket** | 9093 | Bidirectional real-time communication | Live chat discussions |

### Architecture Diagram

```
                          mangahub CLI
                       (Cobra Command)
                              |
                 _____________________________|
                |         |        |         |
              HTTP       TCP      UDP    WebSocket
              8080      9090     9091     9093
                |         |        |         |
        ┌───────┴─────────┴────────┴─────────┘
        |
    ┌───────────────────────────────────────┐
    |    gRPC Internal Service (9092)       |
    │  (Fast microservice communication)    |
    └───────────────────────────────────────┘
        |
    ┌───────────────────────────────────────┐
    |    SQLite Database                    |
    | (users, manga, user_progress tables)  |
    └───────────────────────────────────────┘
```

---

## 📁 Project Structure

```
mangahub/
├── cmd/                                ← CLI Entry Points (Cobra)
│   ├── main/
│   │   └── main.go                    ← Binary entry point
│   ├── root.go                        ← Root command "mangahub"
│   ├── server.go                      ← `mangahub server start` (launches 5 servers)
│   ├── auth.go                        ← `mangahub auth register/login`
│   ├── manga.go                       ← `mangahub manga search/info`
│   ├── library.go                     ← `mangahub library add/list/remove`
│   ├── progress.go                    ← `mangahub progress update`
│   ├── sync.go                        ← `mangahub sync connect` (Phase 3)
│   ├── chat.go                        ← `mangahub chat join` (Phase 3)
│   └── notifications.go               ← `mangahub notifications listen` (Phase 3)
│
├── internal/                           ← Private Application Logic
│   ├── database/
│   │   └── db.go                      ← SQLite initialization & migrations
│   │
│   ├── auth/
│   │   ├── handler.go                 ← HTTP handlers for register/login
│   │   ├── service.go                 ← Business logic (bcrypt, JWT)
│   │   └── middleware.go              ← JWT validation middleware
│   │
│   ├── manga/
│   │   ├── repository.go              ← Database queries for manga table
│   │   ├── service.go                 ← Business logic
│   │   └── handler.go                 ← HTTP handlers (Phase 3)
│   │
│   ├── user/
│   │   ├── repository.go              ← Database queries for user_progress
│   │   └── service.go                 ← Business logic
│   │
│   ├── http_server/
│   │   ├── server.go                  ← Gin HTTP server (Port 8080)
│   │   └── routes.go                  ← Route registration
│   │
│   ├── tcp_server/
│   │   ├── server.go                  ← TCP listener & handler (Port 9090)
│   │   └── hub.go                     ← Connection registry + broadcast
│   │
│   ├── udp_server/
│   │   └── server.go                  ← UDP listener & notifier (Port 9091)
│   │
│   ├── grpc_server/
│   │   ├── server.go                  ← gRPC server setup (Port 9092)
│   │   ├── manga_service.go           ← Implements MangaServiceServer RPC
│   │   └── interceptor.go             ← gRPC middleware (auth, logging)
│   │
│   ├── websocket_server/
│   │   ├── server.go                  ← WebSocket hub setup (Port 9093)
│   │   ├── hub.go                     ← Client registry + broadcast
│   │   └── client.go                  ← Per-connection read/write pumps
│   │
│   ├── tcp_client/                    ← Phase 3
│   │   └── client.go                  ← TCP client for `sync connect`
│   │
│   ├── udp_client/                    ← Phase 3
│   │   └── client.go                  ← UDP client for notifications
│   │
│   ├── grpc_client/                   ← Phase 3
│   │   └── client.go                  ← gRPC client wrapper
│   │
│   └── websocket_client/              ← Phase 3
│       └── client.go                  ← WebSocket client for chat
│
├── models/                             ← Shared Data Models
│   └── models.go                      ← User, Manga, UserProgress structs
│
├── proto/                              ← Protocol Buffer Definitions
│   └── manga/
│       ├── manga.proto                ← Service + message definitions
│       ├── manga.pb.go                ← Auto-generated (never edit)
│       └── manga_grpc.pb.go           ← Auto-generated (never edit)
│
├── PHASE1.md                          ← Phase 1 Documentation (✅ Complete)
├── PHASE2_CHECKLIST.md                ← Phase 2 Detailed Checklist
├── PHASE2.md                          ← Phase 2 Implementation Guide
├── PHASE3_CHECKLIST.md                ← Phase 3 Detailed Checklist
├── go.mod                             ← Dependency management
├── go.sum                             ← Dependency lock file
├── mangahub.db                        ← SQLite database file
└── README.md                          ← This file
```

---

## 🎯 Project Phases

### Phase 1: Foundation (✅ Complete)
**Status:** DONE — Weeks 1-2

**Accomplished:**
- Go module setup with Gin, JWT, bcrypt, Cobra
- SQLite database with 3 tables (users, manga, user_progress)
- User authentication: register + login with password hashing
- JWT token generation (24-hour expiration)
- HTTP REST API with health check `/ping`
- Endpoints: `/auth/register`, `/auth/login`
- CLI commands: `mangahub auth register/login`
- CLI command: `mangahub server start` (launches HTTP server)

**Documentation:** See [PHASE1.md](./PHASE1.md)

---

### Phase 2: Multi-Protocol Backend (⏳ In Progress)
**Status:** READY TO BUILD — Weeks 3-5

**Goals:**
- Build 5 concurrent servers launching simultaneously
- Implement TCP, UDP, gRPC, WebSocket protocols
- Master goroutines, channels, context management
- Implement hub patterns and broadcast logic

**Servers to Build:**

| Server | Port | Task |
|--------|------|------|
| TCP Sync | 9090 | Broadcast reading progress, hub pattern with channels |
| UDP Notifications | 9091 | Send chapter release alerts to registered clients |
| gRPC Service | 9092 | Fast RPC for manga queries and operations |
| WebSocket Chat | 9093 | Real-time multi-client chat rooms |
| HTTP (existing) | 8080 | Keep from Phase 1 |

**Documentation:**
- Detailed guide: [PHASE2.md](./PHASE2.md)
- Implementation checklist: [PHASE2_CHECKLIST.md](./PHASE2_CHECKLIST.md)

---

### Phase 3: CLI Client Integration (📋 Planning)
**Status:** READY TO BUILD — Weeks 6-8

**Goals:**
- Connect CLI commands to Phase 2 backend servers
- Implement 5 client types: TCP, UDP, gRPC, WebSocket
- Add 8 major CLI command groups
- Context timeouts and error handling

**CLI Commands to Add:**

```bash
# TCP Sync
mangahub sync connect                           # Listen to progress updates

# UDP Notifications
mangahub notifications listen --username alice  # Subscribe to new chapters

# gRPC Manga Service
mangahub manga search "OnePiece"                # Search by title
mangahub manga info <manga-id>                  # Get details
mangahub library add --manga-id <id> --status reading
mangahub library list
mangahub library remove --manga-id <id>

# WebSocket Chat
mangahub chat join --username alice             # Join live chat

# Progress (Updated)
mangahub progress update --manga-id <id> --chapter 50  # Updates HTTP + TCP broadcast
```

**Documentation:**
- Detailed guide: [PHASE3_CHECKLIST.md](./PHASE3_CHECKLIST.md)

---

### Phase 4: Polish & Bonus (🔧 Optional)
**Status:** PLANNED — Weeks 9-12

**Bonus Features:**
- Error recovery & reconnection logic
- Configuration file for secrets
- Redis caching for manga data
- User reviews & ratings system
- Advanced filtering and sorting
- Load testing & performance tuning
- Docker containerization

---

## 🚀 Quick Start

### Prerequisites

```bash
# Required
- Go 1.19+
- SQLite3 (included with most systems)
- GCC/CGO (for sqlite3 driver)

# Optional (for Phase 2+)
- protoc (Protocol Buffer compiler)
- wscat (WebSocket testing)
- grpcurl (gRPC testing)
```

### Installation

```bash
# 1. Clone repository
git clone <repo-url>
cd mangahub

# 2. Install dependencies
go mod tidy

# 3. Build binary
go build -o mangahub ./cmd/main

# 4. Run Phase 1 (HTTP only)
./mangahub.exe server start
# Output: ✓ HTTP API Server starting on http://localhost:8080

# 5. Test HTTP server (new terminal)
curl http://localhost:8080/ping
# Response: {"message":"pong"}
```

---

## 📝 CLI Command Reference

### Authentication

```bash
# Register new account
./mangahub auth register --username alice --email alice@example.com
# Prompted: Password: ****

# Login
./mangahub auth login --username alice
# Output: Token: eyJhbGciOiJIUzI1NiI...
```

### Server Management

```bash
# Start all servers (Phase 1: just HTTP; Phase 2: all 5 protocols)
./mangahub server start

# Stop servers (Phase 3)
./mangahub server stop

# Check server status (Phase 3)
./mangahub server status
```

### Manga Operations (Phase 3)

```bash
# Search manga by title
./mangahub manga search "OnePiece"

# Get manga details
./mangahub manga info <manga-id>

# Add to library
./mangahub library add --manga-id manga_1 --status reading

# List library
./mangahub library list

# Remove from library
./mangahub library remove --manga-id manga_1
```

### Real-time Features (Phase 3)

```bash
# Connect to TCP progress sync
./mangahub sync connect

# Listen for notifications
./mangahub notifications listen --username alice

# Join chat
./mangahub chat join --username alice

# Update reading progress (broadcasts via TCP)
./mangahub progress update --manga-id manga_1 --chapter 50
```

---

## 🧪 Testing

### Phase 1 Testing

```bash
# Terminal 1: Start server
./mangahub server start

# Terminal 2: Register
./mangahub auth register --username alice --email alice@example.com

# Terminal 2: Login
./mangahub auth login --username alice

# Terminal 2: Health check
curl http://localhost:8080/ping
```

### Phase 2 Testing (After Implementation)

```bash
# Terminal 1: Start all 5 servers
./mangahub server start

# Terminal 2: Connect to TCP (should receive broadcasts)
telnet localhost 9090

# Terminal 3: Register for UDP notifications
echo "REGISTER|alice" | nc -u localhost 9091

# Terminal 4: Send gRPC query
grpcurl -plaintext -d '{"query":"OnePiece"}' \
    localhost:9092 mangahub.MangaService/SearchManga

# Terminal 5: Connect to WebSocket chat
wscat -c ws://localhost:9093?username=alice
```

---

## 🔑 Key Go Concepts Demonstrated

### Phase 1
- Package structure & imports
- Struct definitions & JSON tags
- Error handling (if err != nil)
- Goroutine basics (for CLI commands)

### Phase 2
- **Goroutines:** Concurrent execution with `go` keyword
- **Channels:** Inter-goroutine communication (buffered & unbuffered)
- **Mutexes:** Protecting shared data from races
- **Select:** Multiplexing channels
- **Defer:** Cleanup (closing connections, unlocking)
- **Context:** Timeouts and cancellation

### Phase 3
- **Dialing connections:** TCP, UDP, WebSocket, gRPC clients
- **Event loops:** Reading from stdin/connections
- **Error recovery:** Reconnection logic
- **Graceful shutdown:** Closing goroutines cleanly

---

## 📊 Project Statistics

| Metric | Count |
|--------|-------|
| **Total CLI Commands** | 12+ |
| **Backend Servers** | 5 |
| **Network Protocols** | 5 |
| **Go Packages** | 15+ |
| **Database Tables** | 3 |
| **Lines of Documentation** | 1000+ |
| **Estimated Code** | 2000+ lines |

---

## 🎓 Learning Objectives

By completing this project, you will master:

✅ **Network Programming:** TCP, UDP, HTTP, gRPC, WebSocket protocols  
✅ **Concurrent Programming:** Goroutines, channels, mutexes, context  
✅ **Database Design:** SQLite schema, migrations, queries  
✅ **Go Best Practices:** Error handling, code organization, testing  
✅ **Architecture Patterns:** Hub patterns, broadcast systems, client-server  
✅ **CLI Development:** Cobra framework, argument parsing, user interaction

---

## 🤝 Contributing

Since this is a course project, changes follow the phase schedule:

1. Complete current phase 100%
2. Pass testing checklist
3. Write documentation
4. Move to next phase

---

## 📞 Support

For questions about:
- **Architecture:** See PHASE1.md, PHASE2.md, PHASE3_CHECKLIST.md
- **Go Concepts:** Refer to respective phase documentation
- **Implementation:** Check checklist items with pseudocode

---

## 📜 License

This project is part of IT096IU (Net-centric Programming course).

---

## 🗓️ Timeline

| Phase | Duration | Status |
|-------|----------|--------|
| Phase 1: Foundation | Weeks 1-2 | ✅ Complete |
| Phase 2: Protocols | Weeks 3-5 | ⏳ In Progress |
| Phase 3: Integration | Weeks 6-8 | 📋 Planned |
| Phase 4: Polish | Weeks 9-12 | 🔧 Optional |

---

**Last Updated:** 2026-05-09  
**Project Status:** Phase 2 Ready to Start
