# MangaHub - Phase 1 Flow

Tài liệu này giải thích toàn bộ flow của Phase 1 trong project MangaHub hiện tại. Phase 1 tập trung vào nền móng của chương trình: CLI bằng Cobra, SQLite database, HTTP server bằng Gin, và chức năng authentication gồm register/login với bcrypt + JWT.

## 1. Mục tiêu của Phase 1

Phase 1 chưa triển khai đầy đủ 5 protocol như đặc tả tổng thể. Ở code hiện tại, Phase 1 đang hoàn thành các phần nền tảng sau:

- Tạo CLI chính `mangahub`.
- Tạo nhóm lệnh `auth` cho đăng ký và đăng nhập.
- Tạo nhóm lệnh `server` để khởi động HTTP API server.
- Khởi tạo SQLite database `mangahub.db`.
- Tạo schema cho `users`, `manga`, `user_progress`.
- Xử lý đăng ký tài khoản với password được hash bằng bcrypt.
- Xử lý đăng nhập và sinh JWT token.
- Cung cấp HTTP endpoint `/ping`, `/auth/register`, `/auth/login`.

Các protocol TCP, UDP, WebSocket và gRPC được nhắc trong README/đặc tả tổng thể nhưng chưa nằm trong flow chạy của Phase 1.

## 2. Cấu trúc code liên quan đến Phase 1

```text
mangahub/
├── cmd/
│   ├── main/main.go          # Entry point: gọi cmd.Execute()
│   ├── root.go               # Root command: mangahub
│   ├── auth.go               # CLI auth register/login
│   └── server.go             # CLI server start
│
├── internal/
│   ├── auth/
│   │   ├── service.go        # Logic register/login, bcrypt, JWT
│   │   └── handler.go        # HTTP handlers cho /auth/register, /auth/login
│   │
│   ├── database/
│   │   └── db.go             # Mở SQLite connection + chạy migrations
│   │
│   └── http_server/
│       └── server.go         # Khởi tạo Gin router và start HTTP server
│
├── models/
│   └── models.go             # Struct User, Manga, UserProgress
│
├── mangahub.db               # SQLite database file
├── mangahub.exe              # Binary đã build
├── go.mod
└── README.md
```

## 3. Flow tổng quan của Phase 1

```text
User chạy CLI
    |
    v
cmd/main/main.go
    |
    v
cmd.Execute()
    |
    v
Cobra root command "mangahub"
    |
    +--> mangahub auth register
    |       |
    |       +--> database.Init("./mangahub.db")
    |       +--> auth.RegisterUser(...)
    |       +--> bcrypt hash password
    |       +--> INSERT INTO users
    |
    +--> mangahub auth login
    |       |
    |       +--> database.Init("./mangahub.db")
    |       +--> auth.LoginUser(...)
    |       +--> SELECT user by username
    |       +--> bcrypt compare password
    |       +--> generate JWT token
    |
    +--> mangahub server start
            |
            +--> database.Init("./mangahub.db")
            +--> http_server.Start("8080")
            +--> Gin routes:
                    GET  /ping
                    POST /auth/register
                    POST /auth/login
```

## 4. Entry point của chương trình

File `cmd/main/main.go` là điểm bắt đầu khi chạy chương trình.

```go
func main() {
    cmd.Execute()
}
```

Nó không xử lý logic trực tiếp. Nhiệm vụ của nó chỉ là chuyển quyền điều khiển sang package `cmd`.

Sau đó, `cmd/root.go` định nghĩa root command:

```go
Use:   "mangahub",
Short: "MangaHub - Your manga tracking system",
```

Root command là gốc của toàn bộ CLI. Các nhóm lệnh như `auth` và `server` được gắn vào root command thông qua `rootCmd.AddCommand(...)`.

## 5. Flow khởi động server

Lệnh:

```powershell
.\mangahub.exe server start
```

Hoặc khi dùng Go:

```powershell
go run .\cmd\main server start
```

Flow xử lý:

```text
User chạy "mangahub server start"
    |
    v
cmd/server.go
    |
    v
database.Init("./mangahub.db")
    |
    +--> tạo/mở SQLite database
    +--> ping database
    +--> runMigrations()
    |
    v
http_server.Start("8080")
    |
    +--> gin.Default()
    +--> auth.RegisterRoutes(r)
    +--> tạo GET /ping
    +--> r.Run(":8080")
```

Trong Phase 1, server chỉ start HTTP API server ở port `8080`. Comment trong code nói rõ các server TCP/UDP/WebSocket/gRPC sẽ được thêm ở Phase 2.

## 6. Flow database

File `internal/database/db.go` quản lý database.

Khi gọi:

```go
database.Init("./mangahub.db")
```

Chương trình thực hiện:

1. Đảm bảo thư mục chứa database tồn tại.
2. Mở kết nối SQLite bằng driver `github.com/mattn/go-sqlite3`.
3. Ping database để kiểm tra connection.
4. Chạy migration để tạo bảng nếu chưa tồn tại.
5. Gán connection vào biến global `database.DB`.

Biến:

```go
var DB *sql.DB
```

được dùng chung bởi các package khác, ví dụ `internal/auth/service.go`.

## 7. Schema database trong Phase 1

Phase 1 tạo 3 bảng:

### 7.1. Bảng `users`

Dùng cho authentication.

```sql
CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    username      TEXT UNIQUE NOT NULL,
    email         TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Ý nghĩa:

- `id`: ID người dùng, được sinh theo dạng `usr_<timestamp>`.
- `username`: tên đăng nhập, không được trùng.
- `email`: email, không được trùng.
- `password_hash`: password đã hash bằng bcrypt, không lưu password gốc.
- `created_at`: thời điểm tạo tài khoản.

### 7.2. Bảng `manga`

Dùng để lưu thông tin manga cho các phase sau.

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

Trong Phase 1, bảng này đã được tạo nhưng chưa có CLI/API xử lý manga.

### 7.3. Bảng `user_progress`

Dùng để lưu tiến độ đọc manga của user.

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

Trong Phase 1, bảng này cũng mới là nền tảng cho các phase sau.

## 8. Flow đăng ký bằng CLI

Lệnh:

```powershell
.\mangahub.exe auth register --username alice --email alice@example.com
```

Sau khi chạy lệnh, chương trình yêu cầu nhập password:

```text
Password:
```

Password được nhập bằng `term.ReadPassword`, nên terminal không hiển thị ký tự password.

Flow xử lý:

```text
CLI auth register
    |
    v
Đọc flag --username
Đọc flag --email
Đọc password ẩn từ terminal
    |
    v
database.Init("./mangahub.db")
    |
    v
auth.RegisterUser(username, email, password)
    |
    +--> bcrypt.GenerateFromPassword(password, 12)
    +--> tạo models.User
    +--> INSERT INTO users
    |
    v
In kết quả ra terminal
```

Nếu đăng ký thành công:

```text
✓ Account created! User ID: usr_...
```

Nếu username/email bị trùng, SQLite sẽ trả lỗi unique constraint và chương trình báo registration failed.

## 9. Flow đăng nhập bằng CLI

Lệnh:

```powershell
.\mangahub.exe auth login --username alice
```

Sau đó nhập password:

```text
Password:
```

Flow xử lý:

```text
CLI auth login
    |
    v
Đọc flag --username
Đọc password ẩn từ terminal
    |
    v
database.Init("./mangahub.db")
    |
    v
auth.LoginUser(username, password)
    |
    +--> SELECT user FROM users WHERE username = ?
    +--> nếu không có user: account not found
    +--> bcrypt.CompareHashAndPassword(...)
    +--> nếu sai password: invalid credentials
    +--> tạo JWT token hết hạn sau 24 giờ
    |
    v
In token ra terminal
```

Nếu đăng nhập thành công:

```text
✓ Welcome back, alice!
Token: <jwt_token>
```

JWT chứa các claim:

```text
user_id
username
exp
```

Trong Phase 1, token mới được sinh ra và in ra terminal. Middleware kiểm tra JWT và các API cần token sẽ phù hợp để làm ở phase sau.

## 10. Flow HTTP API server

Khi chạy:

```powershell
.\mangahub.exe server start
```

HTTP server chạy ở:

```text
http://localhost:8080
```

Các route hiện có:

```text
GET  /ping
POST /auth/register
POST /auth/login
```

### 10.1. Health check `/ping`

Request:

```http
GET /ping
```

Response:

```json
{
  "message": "pong"
}
```

Endpoint này dùng để kiểm tra server có đang chạy hay không.

### 10.2. HTTP register `/auth/register`

Request:

```http
POST /auth/register
Content-Type: application/json
```

Body:

```json
{
  "username": "alice",
  "email": "alice@example.com",
  "password": "password123"
}
```

Validation trong `internal/auth/handler.go`:

- `username`: bắt buộc, tối thiểu 3 ký tự.
- `email`: bắt buộc, đúng format email.
- `password`: bắt buộc, tối thiểu 8 ký tự.

Flow:

```text
POST /auth/register
    |
    v
Gin bind JSON vào registerRequest
    |
    +--> nếu JSON sai hoặc thiếu field: 400 Bad Request
    |
    v
auth.RegisterUser(...)
    |
    +--> hash password bằng bcrypt
    +--> insert user vào SQLite
    |
    +--> nếu username/email trùng: 409 Conflict
    |
    v
201 Created
```

Response thành công:

```json
{
  "message": "Account created successfully",
  "user_id": "usr_...",
  "username": "alice",
  "email": "alice@example.com"
}
```

### 10.3. HTTP login `/auth/login`

Request:

```http
POST /auth/login
Content-Type: application/json
```

Body:

```json
{
  "username": "alice",
  "password": "password123"
}
```

Flow:

```text
POST /auth/login
    |
    v
Gin bind JSON vào loginRequest
    |
    +--> nếu JSON sai hoặc thiếu field: 400 Bad Request
    |
    v
auth.LoginUser(...)
    |
    +--> tìm user theo username
    +--> so sánh password với password_hash
    +--> tạo JWT nếu hợp lệ
    |
    +--> nếu sai username/password: 401 Unauthorized
    |
    v
200 OK
```

Response thành công:

```json
{
  "message": "Login successful",
  "token": "<jwt_token>",
  "username": "alice",
  "user_id": "usr_..."
}
```

## 11. Quan hệ giữa CLI và HTTP trong Phase 1

Một điểm quan trọng: CLI auth và HTTP auth đang dùng chung service layer.

```text
CLI auth register/login
        |
        v
internal/auth/service.go
        ^
        |
HTTP /auth/register, /auth/login
```

Điều này có nghĩa là:

- Đăng ký bằng CLI và đăng ký bằng HTTP đều insert vào cùng bảng `users`.
- Đăng nhập bằng CLI và đăng nhập bằng HTTP đều dùng cùng logic kiểm tra password.
- Nếu sau này sửa logic auth trong `service.go`, cả CLI và HTTP đều được hưởng chung thay đổi đó.

Đây là cách tách layer khá tốt:

- `cmd/auth.go`: giao diện CLI.
- `internal/auth/handler.go`: giao diện HTTP.
- `internal/auth/service.go`: business logic thật.
- `internal/database/db.go`: database connection.

## 12. Models trong Phase 1

File `models/models.go` định nghĩa các struct chính.

### User

```go
type User struct {
    ID           string
    Username     string
    Email        string
    PasswordHash string
    CreatedAt    time.Time
}
```

`PasswordHash` có tag:

```go
json:"-"
```

Nghĩa là khi convert user sang JSON, password hash sẽ không bị trả ra response.

### Manga

Đã có model `Manga`, nhưng Phase 1 chưa có flow search/list/info.

### UserProgress

Đã có model `UserProgress`, nhưng Phase 1 chưa có flow library/progress.

## 13. Cách chạy và kiểm thử thủ công

### 13.1. Build chương trình

```powershell
go build -o mangahub.exe .\cmd\main
```

### 13.2. Chạy HTTP server

```powershell
.\mangahub.exe server start
```

Server chạy tại:

```text
http://localhost:8080
```

### 13.3. Test health check

```powershell
curl http://localhost:8080/ping
```

Kết quả mong đợi:

```json
{"message":"pong"}
```

### 13.4. Đăng ký bằng CLI

```powershell
.\mangahub.exe auth register --username alice --email alice@example.com
```

### 13.5. Đăng nhập bằng CLI

```powershell
.\mangahub.exe auth login --username alice
```

### 13.6. Đăng ký bằng HTTP

```powershell
curl -X POST http://localhost:8080/auth/register `
  -H "Content-Type: application/json" `
  -d "{\"username\":\"bob\",\"email\":\"bob@example.com\",\"password\":\"password123\"}"
```

### 13.7. Đăng nhập bằng HTTP

```powershell
curl -X POST http://localhost:8080/auth/login `
  -H "Content-Type: application/json" `
  -d "{\"username\":\"bob\",\"password\":\"password123\"}"
```

## 14. Những gì Phase 1 đã làm được

- CLI đã có root command.
- CLI đã có nhóm `auth`.
- CLI đã có nhóm `server`.
- SQLite database đã được khởi tạo tự động.
- Schema database đã được tạo tự động khi start server hoặc dùng auth CLI.
- User có thể đăng ký.
- Password không lưu dạng plain text mà được hash bằng bcrypt.
- User có thể đăng nhập.
- Hệ thống có thể sinh JWT token.
- HTTP server có endpoint kiểm tra `/ping`.
- HTTP server có endpoint auth register/login.

## 15. Những gì chưa nằm trong Phase 1 hiện tại

Các phần sau thuộc phase tiếp theo hoặc chưa được implement trong code hiện tại:

- Manga search/info/list.
- Library add/remove/list.
- Reading progress update/history.
- TCP sync server/client.
- UDP notification server.
- WebSocket chat.
- gRPC internal manga service.
- JWT middleware bảo vệ route.
- Config file cho JWT secret.
- Seed data cho manga.
- Test tự động.

## 16. Tóm tắt ngắn flow Phase 1

```text
Phase 1 = CLI + SQLite + HTTP Auth nền tảng

main()
  -> cmd.Execute()
  -> Cobra parse command
  -> auth command hoặc server command

auth register
  -> đọc username/email/password
  -> init DB
  -> hash password
  -> insert user

auth login
  -> đọc username/password
  -> init DB
  -> query user
  -> compare password
  -> generate JWT

server start
  -> init DB
  -> run migration
  -> start Gin HTTP server
  -> expose /ping, /auth/register, /auth/login
```

Để hiểu code Phase 1 nhanh nhất, nên đọc theo thứ tự:

1. `cmd/main/main.go`
2. `cmd/root.go`
3. `cmd/server.go`
4. `internal/database/db.go`
5. `internal/http_server/server.go`
6. `internal/auth/handler.go`
7. `cmd/auth.go`
8. `internal/auth/service.go`
9. `models/models.go`
