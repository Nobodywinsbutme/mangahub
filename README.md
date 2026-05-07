# MangaHub - Manga & Comic Tracking System

**Course:** Net-centric Programming (IT096IU)  
**Instructors:** Lê Thanh Sơn - Nguyễn Trung Nghĩa

## Team Members (Group 25)

- Lê Đoàn Minh Ngọc (ID: ITITDK23023)
- Phạm Trung Kiên (ID: ITCSIU23020)

## Project Overview

MangaHub is a command-line interface (CLI) manga tracking system built in Go. It demonstrates advanced network programming concepts by integrating five distinct communication protocols into a single cohesive application. Users can search for manga, manage their personal reading libraries, synchronize their reading progress across devices in real-time, receive notifications, and chat with other readers.

### Core Network Protocols Implemented

1. **HTTP (REST API):** Handles user authentication (JWT), core CRUD operations, and SQLite database interactions.
2. **TCP (Progress Sync):** A raw socket server utilizing goroutines to broadcast real-time reading progress to all connected client devices.
3. **UDP (Notifications):** A connectionless broadcaster for pushing new chapter release alerts to registered clients.
4. **WebSocket (Chat):** A full-duplex communication hub for real-time community discussions.
5. **gRPC (Internal Service):** A high-performance RPC framework using Protocol Buffers for fast internal microservice communication.

---

## System Architecture Diagram

```text
                                 +-------------------------+
                                 |   mangahub CLI Client   |
                                 | (Cobra Command Parser)  |
                                 +-------------------------+
                                   /      |      |      |
                          HTTP    /  TCP  |  UDP |   WS |
                        (8080)   / (9090) |(9091)|(9093)|
                                v         v      v      v
     +-----------------+  +--------+ +--------+ +--------+
     | HTTP REST API   |  | TCP    | | UDP    | | Web    |
     | (Gin/Auth/CRUD) |  | Sync   | | Notify | | Socket |
     +-----------------+  +--------+ +--------+ +--------+
               |               ^         |          ^
               v               v         v          v
     +---------------------------------------------------+
     |               gRPC Internal Service               |
     |                      (9092)                       |
     +---------------------------------------------------+
               |                 |                  |
               v                 v                  v
     +---------------------------------------------------+
     |                 SQLite Database                   |
     |           (users, manga, user_progress)           |
     +---------------------------------------------------+
```

## Project Directory Structure

```plaintext
mangahub/
│
├── cmd/                          ← Cobra CLI entry points
│   ├── root.go                   ← Root command, global flags
│   ├── server.go                 ← `mangahub server start/stop/status`
│   ├── auth.go                   ← `mangahub auth register/login/logout`
│   ├── manga.go                  ← `mangahub manga search/info/list`
│   ├── library.go                ← `mangahub library add/remove/list`
│   ├── progress.go               ← `mangahub progress update/history`
│   ├── sync.go                   ← `mangahub sync connect/disconnect`
│   ├── chat.go                   ← `mangahub chat join/send`
│   └── main.go                   ← main() — calls cmd.Execute()
│
├── internal/                     ← Private app logic (not importable externally)
│   ├── auth/
│   │   ├── handler.go            ← HTTP handlers: register, login
│   │   ├── middleware.go         ← JWT validation middleware
│   │   └── service.go            ← Business logic: hash, verify, generate token
│   │
│   ├── manga/
│   │   ├── handler.go            ← HTTP handlers: search, get, list
│   │   ├── repository.go         ← DB queries for manga table
│   │   └── service.go            ← Business logic layer
│   │
│   ├── user/
│   │   ├── handler.go            ← HTTP handlers: library, progress
│   │   ├── repository.go         ← DB queries: user_progress, users
│   │   └── service.go
│   │
│   ├── database/
│   │   ├── db.go                 ← Open connection, run migrations
│   │   └── schema.sql            ← CREATE TABLE statements
│   │
│   ├── tcp/
│   │   ├── server.go             ← TCP listener, accept loop
│   │   ├── client.go             ← TCP client (for `mangahub sync connect`)
│   │   └── hub.go                ← Broadcast channel, connection registry
│   │
│   ├── udp/
│   │   ├── server.go             ← UDP listener, registration handler
│   │   └── notifier.go           ← Broadcast to registered []UDPAddr
│   │
│   ├── websocket/
│   │   ├── server.go             ← HTTP→WS upgrade, start hub
│   │   ├── hub.go                ← Register/Unregister/Broadcast channels
│   │   ├── client.go             ← Per-connection read/write pumps
│   │   └── ws_client.go          ← CLI client (for `mangahub chat join`)
│   │
│   └── grpc/
│       ├── server.go             ← gRPC server setup, service registration
│       └── manga_service.go      ← Implements MangaServiceServer interface
│
├── proto/                        ← Protobuf definitions
│   └── manga/
│       ├── manga.proto           ← Service + message definitions
│       └── manga.pb.go           ← Auto-generated (do NOT edit by hand)
│       └── manga_grpc.pb.go      ← Auto-generated
│
├── pkg/                          ← Reusable, potentially exportable packages
│   ├── models/
│   │   └── models.go             ← Shared structs: User, Manga, Progress, etc.
│   ├── config/
│   │   └── config.go             ← Load ~/.mangahub/config.yaml
│   └── utils/
│       └── utils.go              ← Shared helpers: JSON encode, timestamp, etc.
│
├── data/
│   └── manga_seed.json           ← Initial manga data (100+ series)
│
├── docs/
│   └── api.md                    ← API documentation
│
├── go.mod
├── go.sum
├── docker-compose.yml            ← Optional bonus
└── README.md
```

## Setup & Installation

**Prerequisites:**

- Go 1.19 or later

- GCC/CGO enabled (required for SQLite3)

1. Clone the repository

```Bash
git clone <repository_url>
cd mangahub
```

2. Install dependencies

```Bash
go mod tidy
```

3. Build the CLI

```Bash
go build -o mangahub ./cmd/mangahub
```

4. Run the servers

```Bash
./mangahub server start
```
