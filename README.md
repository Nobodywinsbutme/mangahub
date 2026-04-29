# MangaHub - Manga & Comic Tracking System

**Course:** Net-centric Programming (IT096IU)  
**Instructors:** Lê Thanh Sơn - Nguyễn Trung Nghĩa  

## Team Members (Group 25)
* Lê Đoàn Minh Ngọc (ID: ITITDK23023)
* Phạm Trung Kiên (ID: ITCSIU23020)

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
├── cmd/
│   └── mangahub/          # Entry point for the application
│       └── main.go        # Initializes CLI and root commands
├── internal/
│   ├── cli/               # CLI subcommands (server start, auth, search, etc.)
│   ├── server/            # Orchestrator to boot all 5 servers concurrently
│   ├── api/               # HTTP (Gin) routes and handlers
│   ├── tcp/               # TCP sync server and connection pool
│   ├── udp/               # UDP broadcaster and client registry
│   ├── websocket/         # Gorilla WS chat hub and client logic
│   ├── grpc/              # gRPC service implementations
│   ├── models/            # Data structs (User, Manga, Progress)
│   └── repository/        # SQLite database operations and queries
├── pkg/
│   ├── auth/              # JWT generation, validation, and bcrypt hashing
│   └── config/            # Configuration and environment variables
├── proto/                 # Protocol Buffer definitions
│   └── service.proto      
├── data/                  # Local storage for data.db (SQLite file)
├── go.mod                 # Go module definition
└── go.sum                 # Dependency checksums
```
## Setup & Installation
**Prerequisites:**
* Go 1.19 or later

* GCC/CGO enabled (required for SQLite3)
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