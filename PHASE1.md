# MangaHub — Phase 1 (Foundation: CLI + Auth + HTTP)

**Status:** ✅ Complete

Phase 1 builds the foundation of MangaHub: SQLite persistence, CLI auth commands, and an HTTP API with JWT authentication.

---

## ✅ Objectives Completed
- Go module + dependencies
- SQLite schema (users, manga, user_progress)
- CLI framework (Cobra)
- Auth: register + login (bcrypt)
- HTTP API: `/auth/register`, `/auth/login`, `/ping`
- JWT generation on login
- DB initialization + migrations

---

## 📁 Project Structure (Phase 1 scope)
```
mangahub/
├── cmd/
│   ├── main/
│   │   └── main.go
│   ├── root.go
│   ├── auth.go
│   └── server.go
├── internal/
│   ├── auth/
│   │   ├── handler.go
│   │   └── service.go
│   ├── database/
│   │   └── db.go
│   └── http_server/
│       └── server.go
├── models/
│   └── models.go
└── PHASE1.md
```

---

## 🗄️ Database Schema
### users
```sql
CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    username      TEXT UNIQUE NOT NULL,
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### manga
```sql
CREATE TABLE IF NOT EXISTS manga (
    id             TEXT PRIMARY KEY,
    title          TEXT NOT NULL,
    author         TEXT NOT NULL,
    genres         TEXT DEFAULT '[]',
    status         TEXT NOT NULL,
    total_chapters INTEGER DEFAULT 0,
    description    TEXT DEFAULT ''
);
```

### user_progress
```sql
CREATE TABLE IF NOT EXISTS user_progress (
    user_id         TEXT NOT NULL,
    manga_id        TEXT NOT NULL,
    current_chapter INTEGER DEFAULT 0,
    status          TEXT NOT NULL,
    updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, manga_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (manga_id) REFERENCES manga(id)
);
```

---

## 🔐 Auth Flow
**Register:**
1. Validate JSON (required + min length)
2. Hash password (bcrypt)
3. Insert user

**Login:**
1. Find user by username
2. Compare bcrypt hash
3. Generate JWT

JWT claims:
```
user_id
username
exp
```

---

## 🚀 Usage
### Build
```bash
go build -o mangahub.exe ./cmd/main
```

### Start server
```bash
./mangahub.exe server start
```

### CLI Register
```bash
./mangahub.exe auth register --username alice --email alice@example.com
```

### CLI Login
```bash
./mangahub.exe auth login --username alice
```

### HTTP (PowerShell)
**Register:**
```powershell
Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/auth/register `
  -ContentType "application/json" `
  -Body '{"username":"alice","email":"alice@example.com","password":"password123"}'
```

**Login:**
```powershell
Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/auth/login `
  -ContentType "application/json" `
  -Body '{"username":"alice","password":"password123"}'
```

---

## ✅ Phase 1 Checklist
| Component | Status |
|-----------|--------|
| CLI (Cobra) | ✅ |
| SQLite schema | ✅ |
| Auth (bcrypt + JWT) | ✅ |
| HTTP API (Gin) | ✅ |
| Migrations | ✅ |

---

## 🔄 Transition to Phase 2
Phase 2 adds **multi-protocol servers** (TCP/UDP/gRPC/WebSocket) and runs all servers concurrently under `mangahub server start`.
