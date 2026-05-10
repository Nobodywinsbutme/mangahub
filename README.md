# MangaHub - Manga & Comic Tracking System

**Course:** Net-centric Programming (IT096IU)  
**Instructors:** Lê Thanh Sơn - Nguyễn Trung Nghĩa

## 👥 Team Members (Group 25)

- Lê Đoàn Minh Ngọc (ID: ITITDK23023)
- Phạm Trung Kiên (ID: ITCSIU23020)

---

## 📖 Project Overview

MangaHub is a CLI-based manga tracking system built in Go.  
It demonstrates **multi-protocol networking** using:

- HTTP (REST)
- TCP
- UDP
- gRPC
- WebSocket

### Core Features

- User authentication (bcrypt + JWT)
- Real-time progress sync (TCP)
- Chapter notifications (UDP)
- Manga search (gRPC)
- Community chat (WebSocket)

---

## 🏛️ System Architecture

### Network Protocols Implemented

| Protocol | Port | Purpose |
|----------|------|---------|
| HTTP | 8080 | Auth + REST APIs |
| TCP | 9090 | Progress broadcast |
| UDP | 9091 | Notifications |
| gRPC | 9092 | Manga queries |
| WebSocket | 9093 | Chat |

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
├── .gitignore
│   go.mod
│   go.sum
│   mangahub.db
│   mangahub.exe
│   PHASE1.md
│   PHASE2.md
│   PHASE3.md
│   README.md
│   test_ws.html
│   
├───cmd
│   │   auth.go
│   │   chat.go
│   │   notifications.go
│   │   progress.go
│   │   root.go
│   │   server.go
│   │   sync.go
│   │   
│   └───main
│           main.go
│           mangahub.db
│           
├───data
├───internal
│   ├───api
│   ├───auth
│   │       handler.go
│   │       service.go
│   │       
│   ├───cli
│   ├───database
│   │       db.go
│   │       
│   ├───grpc
│   │       server.go
│   │       
│   ├───http_server
│   │       server.go
│   │       
│   ├───models
│   ├───repository
│   ├───server
│   ├───tcp
│   │       server.go
│   │       
│   ├───tcp_client
│   │       client.go
│   │       
│   ├───udp
│   │       server.go
│   │       
│   ├───udp_client
│   │       client.go
│   │       
│   ├───websocket
│   │       server.go
│   │       
│   └───ws_client
│           client.go
│           
├───mangahub
│   │   main.exe
│   │   
│   └───proto
│       └───manga
├───models
│       models.go
│       
├───pkg
│   ├───auth
│   └───config
└───proto
        manga.pb.go
        manga.proto
        manga_grpc.pb.go
        service.proto
```

---

## ✅ Project Phases

### Phase 1: Foundation
Status: ✅ Complete  
Docs: [PHASE1.md](./PHASE1.md)

### Phase 2: Multi‑Protocol Servers
Status: ✅ Complete  
Docs: [PHASE2.md](./PHASE2.md)

### Phase 3: Client Integration
Status: ✅ Complete  
Docs: [PHASE3.md](./PHASE3.md)

---

## 🚀 Quick Start

### Build
```bash
go build -o mangahub.exe ./cmd/main
```

### Start all servers
```bash
./mangahub.exe server start
```

---

## 🧪 Testing (CLI)

### Auth (HTTP)
```powershell
Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/auth/register `
  -ContentType "application/json" `
  -Body '{"username":"alice","email":"alice@example.com","password":"password123"}'
```

### TCP Sync
```bash
./mangahub.exe sync connect
./mangahub.exe progress update --manga-id manga_1 --chapter 50
```

### UDP Notifications
```bash
./mangahub.exe notifications listen --username alice
./mangahub.exe notifications send --title "One Piece" --chapter 1100
```

### WebSocket Chat
```bash
./mangahub.exe chat --username alice
./mangahub.exe chat --username bob
```

---

## ✅ Status
All 5 protocols are live and tested.

**Last Updated:** 2026-05-10  
