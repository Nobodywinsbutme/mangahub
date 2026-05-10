# MangaHub - Phase 4 Demo Preparation

Status: Complete

Phase 4 is the final demo-preparation phase. The goal is not to add a new protocol, but to make the whole system easy to present, test, explain, and divide fairly between two team members.

## 1. Phase 4 Scope

Based on the project specification, the final demo must prove that MangaHub uses all five required network protocols in one coherent manga tracking system:

| Protocol | Port | Demo Feature |
| --- | ---: | --- |
| HTTP REST | 8080 | Auth, manga search/details, user library, reading progress |
| TCP | 9090 | Real-time reading progress broadcast |
| UDP | 9091 | Chapter release notification broadcast |
| gRPC | 9092 | Internal manga search/get/update service |
| WebSocket | 9093 | Real-time chat |

Phase 4 adds final demo readiness:

- HTTP endpoints required by the spec.
- Seed manga data for search/get demos.
- gRPC CLI commands so the service can be shown without extra tools.
- Browser WebSocket test page compatibility.
- Demo checklist and team task split.

## 2. Final Feature Checklist

| Requirement | Status | Evidence |
| --- | --- | --- |
| `POST /auth/register` | Done | `internal/auth/handler.go` |
| `POST /auth/login` | Done | `internal/auth/handler.go` |
| `GET /manga` | Done | `internal/http_server/server.go` |
| `GET /manga/:id` | Done | `internal/http_server/server.go` |
| `POST /users/library` | Done | `internal/http_server/server.go` |
| `GET /users/library` | Done | `internal/http_server/server.go` |
| `PUT /users/progress` | Done | `internal/http_server/server.go` |
| JWT authentication | Done | `jwtMiddleware()` |
| SQLite database | Done | `internal/database/db.go` |
| Manga seed data | Done | `seedManga()` |
| TCP sync | Done | `internal/tcp/server.go`, `internal/tcp_client/client.go` |
| UDP notification | Done | `internal/udp/server.go`, `internal/udp_client/client.go` |
| WebSocket chat | Done | `internal/websocket/server.go`, `internal/ws_client/client.go` |
| gRPC service | Done | `internal/grpc/server.go`, `proto/manga.proto` |
| gRPC CLI demo | Done | `cmd/manga.go` |

## 3. Build And Run

Build:

```powershell
go build -o mangahub.exe .\cmd\main
```

Start all servers:

```powershell
.\mangahub.exe server start
```

Expected ports:

```text
HTTP      localhost:8080
TCP       localhost:9090
UDP       localhost:9091
gRPC      localhost:9092
WebSocket localhost:9093/ws
```

## 4. Demo Script

Use separate terminals for the live demo.

### Terminal 1 - Start backend

```powershell
.\mangahub.exe server start
```

Explain:

- One command starts all protocol servers.
- SQLite is initialized first.
- HTTP, TCP, UDP, gRPC, and WebSocket then run concurrently.

### Terminal 2 - HTTP auth and manga API

Register:

```powershell
$register = Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/auth/register `
  -ContentType "application/json" `
  -Body '{"username":"alice","email":"alice@example.com","password":"password123"}'
```

Login:

```powershell
$login = Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/auth/login `
  -ContentType "application/json" `
  -Body '{"username":"alice","password":"password123"}'

$token = $login.token
```

Search manga:

```powershell
Invoke-RestMethod "http://localhost:8080/manga?query=One"
```

Get manga detail:

```powershell
Invoke-RestMethod "http://localhost:8080/manga/one-piece"
```

Add to library:

```powershell
Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/users/library `
  -Headers @{ Authorization = "Bearer $token" } `
  -ContentType "application/json" `
  -Body '{"manga_id":"one-piece","current_chapter":1,"status":"reading"}'
```

Update progress:

```powershell
Invoke-RestMethod -Method Put `
  -Uri http://localhost:8080/users/progress `
  -Headers @{ Authorization = "Bearer $token" } `
  -ContentType "application/json" `
  -Body '{"manga_id":"one-piece","current_chapter":50,"status":"reading"}'
```

Get library:

```powershell
Invoke-RestMethod -Uri http://localhost:8080/users/library `
  -Headers @{ Authorization = "Bearer $token" }
```

### Terminal 3 - TCP progress sync

Listener:

```powershell
.\mangahub.exe sync connect
```

In another terminal:

```powershell
.\mangahub.exe progress update --manga-id one-piece --chapter 51
```

Expected result:

```text
Update: Progress: usr_123|one-piece|51
```

### Terminal 4 - UDP notifications

Listener:

```powershell
.\mangahub.exe notifications listen --username alice
```

In another terminal:

```powershell
.\mangahub.exe notifications send --title "One Piece" --chapter 1111
```

Expected result:

```text
New Chapter: One Piece - Chapter 1111
```

### Terminal 5 - WebSocket chat

User 1:

```powershell
.\mangahub.exe chat --username alice
```

User 2:

```powershell
.\mangahub.exe chat --username bob
```

Type messages in both terminals and show real-time broadcast.

Optional browser test:

```text
Open test_ws.html while the server is running.
```

### Terminal 6 - gRPC manga service

Search through gRPC:

```powershell
.\mangahub.exe manga search --query One
```

Get one manga through gRPC:

```powershell
.\mangahub.exe manga get --id one-piece
```

Update progress through gRPC:

```powershell
.\mangahub.exe manga grpc-progress --user-id usr_123 --manga-id one-piece --chapter 52 --status reading
```

Explain:

- HTTP is used for public REST endpoints.
- gRPC is used as an internal high-performance service.
- Both share the same SQLite data layer.

## 5. Fair Work Assignment For 2 Members

The split below balances implementation, documentation, demo speaking time, and grading categories.

| Area | Member 1: Lê Đoàn Minh Ngọc | Member 2: Phạm Trung Kiên |
| --- | --- | --- |
| HTTP REST API | Auth endpoints, JWT middleware, protected user routes | Manga search/detail endpoints, API validation |
| Database | User schema, auth persistence | Manga seed data, progress persistence |
| TCP | TCP server broadcast loop | TCP CLI client and progress sender |
| UDP | UDP server registration/broadcast | UDP CLI listener/sender |
| WebSocket | WebSocket hub/server | CLI chat client and browser test page |
| gRPC | Proto/service methods | CLI gRPC demo commands |
| Testing | Auth + HTTP manual tests | TCP/UDP/WebSocket/gRPC manual tests |
| Documentation | README, architecture explanation | PHASE docs, command demo script |
| Demo Speaking | Introduce architecture, HTTP, database | Demonstrate TCP, UDP, WebSocket, gRPC |
| Final QA | Check build and server startup | Check demo commands and screenshots/output |

This gives each member ownership of both server-side and client-side work, so the demo does not look like one person only worked on backend while the other only wrote documents.

## 6. Suggested Demo Timing

| Time | Presenter | Content |
| ---: | --- | --- |
| 1 min | Member 1 | Project objective and architecture |
| 2 min | Member 1 | HTTP auth, manga API, JWT protected routes |
| 1 min | Member 1 | SQLite schema and seed manga data |
| 2 min | Member 2 | TCP progress sync |
| 1 min | Member 2 | UDP notification broadcast |
| 2 min | Member 2 | WebSocket chat |
| 1 min | Member 2 | gRPC service and CLI calls |
| 1 min | Both | Summary and Q&A |

## 7. Risk Checklist Before Demo

- Build the binary before the demo.
- Start with a clean terminal layout.
- Keep `server start` running during all protocol tests.
- Use different terminals for blocking commands like `sync connect`, `notifications listen`, and `chat`.
- If `auth/register` says the username already exists, use another username such as `alice2`.
- If ports are busy, stop old `mangahub.exe` processes first.
- Keep this file open during the demo as the command script.

## 8. Final Success Criteria

The project is ready for demo when the team can show:

- User can register/login and receive a JWT.
- Manga can be searched and retrieved.
- User can add manga to library and update progress.
- TCP broadcasts progress updates.
- UDP broadcasts chapter notifications.
- WebSocket sends chat messages in real time.
- gRPC can search/get manga and update progress.
- Both members can explain their assigned parts clearly.
