# MangaHub — Phase 2 (Multi‑Protocol Servers)

**Status:** ✅ Complete

Phase 2 adds four new servers and runs all backends concurrently:
- TCP Sync (9090)
- UDP Notifications (9091)
- gRPC Service (9092)
- WebSocket Chat (9093)
- HTTP (8080) from Phase 1

---

## ✅ Objectives Completed
- Concurrent server startup with goroutines
- TCP broadcast server (progress updates)
- UDP notification server (fire‑and‑forget)
- gRPC Manga service
- WebSocket chat server
- Shared SQLite database access

---

## 🧭 Runtime Architecture
```
mangahub server start
    ↓
database.Init()
    ↓
go http_server.Start("8080")
go tcp.Start("9090")
go udp.Start("9091")
go grpc.Start("9092")
go websocket.Start("9093")

select {}   // keep alive
```

---

## 🔁 Data Flow Examples

### 1) TCP Sync (Progress Updates)
```
progress update → TCP broadcast → all TCP clients
```

### 2) UDP Notifications
```
NOTIFY|title|chapter → UDP broadcast → all registered clients
```

### 3) gRPC Manga
```
client → MangaService.SearchManga → DB query → response
```

### 4) WebSocket Chat
```
client → ws://host:9093/ws?username=alice → hub broadcast → all clients
```

---

## 🧩 Server Details

### ✅ TCP Sync (9090)
- Persistent connections
- Broadcast progress messages
- Hub + broadcast loop

**Format:**
```
USER_ID|MANGA_ID|CHAPTER
```

### ✅ UDP Notifications (9091)
- Stateless UDP packets
- Clients send `REGISTER|username`
- Server stores username → IP:port
- `NOTIFY|title|chapter` triggers broadcast

### ✅ gRPC (9092)
- `GetManga`, `SearchManga`, `UpdateProgress`
- Proto definitions in `proto/manga.proto`
- Uses SQLite internally

### ✅ WebSocket Chat (9093)
- Upgrade from HTTP to WebSocket
- Client sends JSON `{ "text": "hello" }`
- Server broadcasts `{ username, text, timestamp }`

---

## ✅ How to Test (CLI)

**Start server:**
```bash
./mangahub.exe server start
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

### gRPC
Use your CLI gRPC commands or grpcurl (optional).

---

## ✅ Phase 2 Checklist
| Component | Status |
|-----------|--------|
| Concurrent startup | ✅ |
| TCP server | ✅ |
| UDP server | ✅ |
| gRPC server | ✅ |
| WebSocket server | ✅ |
| All ports active | ✅ |

---

## 🔄 Transition to Phase 3
Phase 3 builds **client CLI features** for TCP, UDP, and WebSocket (now completed in PHASE3.md).