# Cách Demo Toàn Bộ Chương Trình MangaHub

File này là kịch bản demo hoàn chỉnh cho project MangaHub. Khi thuyết trình, bạn có thể mở file này và chạy lần lượt từng phần để chứng minh đủ 5 protocol: HTTP, TCP, UDP, WebSocket, gRPC.

## 1. Chuẩn Bị Trước Khi Demo

Mở PowerShell tại thư mục project:

```powershell
cd "C:\STUDY\Net centric\Project\mangahub"
```

Build lại chương trình:

```powershell
go build -o mangahub.exe .\cmd\main
```

Nếu build thành công, kiểm tra CLI:

```powershell
.\mangahub.exe --help
```

Nên chuẩn bị ít nhất 6 terminal:

| Terminal | Mục đích |
| --- | --- |
| Terminal 1 | Chạy toàn bộ server |
| Terminal 2 | Demo HTTP REST API |
| Terminal 3 | Nghe TCP progress sync |
| Terminal 4 | Gửi TCP progress update |
| Terminal 5 | Nghe UDP notification |
| Terminal 6 | Gửi UDP notification / WebSocket / gRPC |

## 2. Terminal 1 - Chạy Toàn Bộ Server

Chạy:

```powershell
.\mangahub.exe server start
```

Giải thích khi demo:

```text
Lệnh này khởi động toàn bộ backend của MangaHub.
Chương trình init SQLite database trước, sau đó chạy đồng thời 5 server:
HTTP ở port 8080, TCP ở 9090, UDP ở 9091, gRPC ở 9092 và WebSocket ở 9093.
```

Các port cần thấy:

```text
HTTP      http://localhost:8080
TCP       localhost:9090
UDP       localhost:9091
gRPC      localhost:9092
WebSocket ws://localhost:9093/ws
```

Giữ Terminal 1 chạy trong suốt buổi demo.

## 3. Demo HTTP REST API

Mục tiêu phần này:

- Đăng ký user.
- Đăng nhập lấy JWT token.
- Search manga.
- Xem chi tiết manga.
- Thêm manga vào library.
- Update reading progress.
- Xem library của user.

### 3.1. Health Check

Terminal 2:

```powershell
Invoke-RestMethod http://localhost:8080/ping
```

Kết quả mong đợi:

```text
message
-------
pong
```

Giải thích:

```text
Endpoint /ping dùng để kiểm tra HTTP server đang hoạt động.
```

### 3.2. Register User

Nếu username `alice` đã tồn tại, đổi thành `alice2`, `alice3`, hoặc tên khác.

```powershell
Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/auth/register `
  -ContentType "application/json" `
  -Body '{"username":"alice","email":"alice@example.com","password":"password123"}'
```

Giải thích:

```text
HTTP API nhận JSON request, validate input, hash password bằng bcrypt rồi lưu user vào SQLite.
Password gốc không được lưu trong database.
```

### 3.3. Login Và Lưu JWT Token

```powershell
$login = Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/auth/login `
  -ContentType "application/json" `
  -Body '{"username":"alice","password":"password123"}'

$token = $login.token
$token
```

Giải thích:

```text
Khi login thành công, server trả về JWT token.
Token này dùng để truy cập các API cần authentication như user library và progress.
```

### 3.4. Search Manga

```powershell
Invoke-RestMethod "http://localhost:8080/manga?query=One"
```

Giải thích:

```text
Endpoint GET /manga dùng để search manga trong SQLite.
Có thể filter theo query, genre hoặc status.
```

Ví dụ filter theo genre:

```powershell
Invoke-RestMethod "http://localhost:8080/manga?genre=Shounen"
```

Ví dụ filter theo status:

```powershell
Invoke-RestMethod "http://localhost:8080/manga?status=ongoing"
```

### 3.5. Xem Chi Tiết Manga

```powershell
Invoke-RestMethod "http://localhost:8080/manga/one-piece"
```

Giải thích:

```text
Endpoint GET /manga/:id trả về thông tin chi tiết của một manga.
Ví dụ ở đây là One Piece.
```

### 3.6. Thêm Manga Vào Library

```powershell
Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/users/library `
  -Headers @{ Authorization = "Bearer $token" } `
  -ContentType "application/json" `
  -Body '{"manga_id":"one-piece","current_chapter":1,"status":"reading"}'
```

Giải thích:

```text
Đây là protected endpoint.
Request phải gửi JWT trong Authorization header.
Server lấy user_id từ token và lưu manga vào user_progress.
```

### 3.7. Update Reading Progress

```powershell
Invoke-RestMethod -Method Put `
  -Uri http://localhost:8080/users/progress `
  -Headers @{ Authorization = "Bearer $token" } `
  -ContentType "application/json" `
  -Body '{"manga_id":"one-piece","current_chapter":50,"status":"reading"}'
```

Giải thích:

```text
Endpoint này cập nhật chapter hiện tại của user.
Nó chứng minh chức năng quản lý tiến độ đọc manga.
```

### 3.8. Xem Library

```powershell
Invoke-RestMethod -Uri http://localhost:8080/users/library `
  -Headers @{ Authorization = "Bearer $token" }
```

Giải thích:

```text
Endpoint này trả về toàn bộ manga trong library của user hiện tại.
```

## 4. Demo TCP Progress Sync

Mục tiêu:

```text
Chứng minh TCP server có thể giữ connection liên tục và broadcast progress update cho client đang kết nối.
```

### 4.1. Terminal 3 - Mở TCP Listener

```powershell
.\mangahub.exe sync connect
```

Terminal này sẽ đứng chờ message.

Giải thích:

```text
Client này kết nối đến TCP server ở port 9090 và lắng nghe progress update real-time.
```

### 4.2. Terminal 4 - Gửi Progress Update

```powershell
.\mangahub.exe progress update --manga-id one-piece --chapter 51
```

Terminal 3 sẽ nhận:

```text
Update: Progress: usr_123|one-piece|51
```

Giải thích:

```text
Đây là TCP raw socket communication.
Một client gửi progress update, TCP server broadcast update đó cho các TCP client đang kết nối.
```

## 5. Demo UDP Notification

Mục tiêu:

```text
Chứng minh UDP server có thể nhận đăng ký client và broadcast notification về chapter mới.
```

### 5.1. Terminal 5 - Đăng Ký Nhận Notification

```powershell
.\mangahub.exe notifications listen --username alice
```

Giải thích:

```text
Client gửi REGISTER|alice qua UDP.
Server lưu địa chỉ IP:port của client để gửi notification sau này.
```

### 5.2. Terminal 6 - Gửi Notification

```powershell
.\mangahub.exe notifications send --title "One Piece" --chapter 1111
```

Terminal 5 sẽ nhận:

```text
New Chapter: One Piece - Chapter 1111
```

Giải thích:

```text
UDP là connectionless protocol.
Server không giữ connection cố định như TCP, chỉ lưu địa chỉ client và gửi datagram notification.
```

## 6. Demo WebSocket Chat

Mục tiêu:

```text
Chứng minh WebSocket hỗ trợ full-duplex real-time chat.
Hai user có thể gửi và nhận tin nhắn ngay lập tức.
```

### 6.1. Terminal 6 - User Alice

```powershell
.\mangahub.exe chat --username alice
```

### 6.2. Mở Terminal Mới - User Bob

```powershell
.\mangahub.exe chat --username bob
```

Gõ tin nhắn ở terminal Alice:

```text
hello bob
```

Gõ tin nhắn ở terminal Bob:

```text
hi alice
```

Giải thích:

```text
WebSocket giữ kết nối hai chiều giữa client và server.
Mỗi message được gửi lên server, sau đó hub broadcast cho tất cả client đang online.
```

### 6.3. Demo Bằng Browser Nếu Muốn

Mở file:

```text
test_ws.html
```

Giải thích:

```text
File HTML này kết nối đến ws://localhost:9093/ws?username=browser và gửi message JSON đến WebSocket server.
```

## 7. Demo gRPC Internal Service

Mục tiêu:

```text
Chứng minh gRPC service hoạt động như internal service cho manga query và progress update.
```

### 7.1. Search Manga Qua gRPC

```powershell
.\mangahub.exe manga search --query One
```

Kết quả mong đợi:

```text
Found 3 manga
- one-piece | One Piece | Eiichiro Oda | 1110 chapters | ongoing
- one-punch-man | One-Punch Man | ONE | 200 chapters | ongoing
- honey-and-clover | Honey and Clover | Chica Umino | 64 chapters | completed
```

Giải thích:

```text
CLI gọi gRPC method SearchManga.
Server query SQLite database và trả kết quả thông qua Protocol Buffers.
```

### 7.2. Get Manga Qua gRPC

```powershell
.\mangahub.exe manga get --id one-piece
```

Giải thích:

```text
CLI gọi gRPC method GetManga để lấy thông tin chi tiết một manga theo ID.
```

### 7.3. Update Progress Qua gRPC

```powershell
.\mangahub.exe manga grpc-progress --user-id usr_123 --manga-id one-piece --chapter 52 --status reading
```

Giải thích:

```text
CLI gọi gRPC method UpdateProgress.
Đây là ví dụ internal service có thể cập nhật reading progress mà không cần đi qua HTTP REST.
```

## 8. Thứ Tự Demo Khuyến Nghị

Nên demo theo thứ tự này để giảng viên dễ theo dõi:

1. Giới thiệu kiến trúc 5 protocol.
2. Start all servers bằng `server start`.
3. Demo HTTP auth + manga + library.
4. Demo TCP progress sync.
5. Demo UDP notification.
6. Demo WebSocket chat.
7. Demo gRPC manga service.
8. Tổng kết database và code structure.

## 9. Phân Công Người Thuyết Trình

| Phần demo | Người 1 | Người 2 |
| --- | --- | --- |
| Giới thiệu project | Chính | Hỗ trợ |
| Kiến trúc tổng quan | Chính | Hỗ trợ |
| HTTP REST API | Chính | Hỗ trợ chạy lệnh |
| Database schema | Chính | Hỗ trợ |
| TCP sync | Hỗ trợ | Chính |
| UDP notification | Hỗ trợ | Chính |
| WebSocket chat | Hỗ trợ | Chính |
| gRPC service | Hỗ trợ | Chính |
| Q&A | Cả hai | Cả hai |

## 10. Câu Nói Mẫu Khi Demo

### Giới thiệu

```text
MangaHub là hệ thống tracking manga viết bằng Go.
Project này tập trung vào network programming và tích hợp đủ 5 protocol bắt buộc: HTTP, TCP, UDP, WebSocket và gRPC.
```

### HTTP

```text
HTTP REST API xử lý các chức năng request-response truyền thống như đăng ký, đăng nhập, search manga, quản lý library và update progress.
Các route /users được bảo vệ bằng JWT token.
```

### TCP

```text
TCP được dùng cho real-time progress sync vì nó giữ connection ổn định và đảm bảo thứ tự message.
```

### UDP

```text
UDP được dùng cho notification vì notification có thể gửi nhanh, nhẹ, không cần duy trì connection liên tục.
```

### WebSocket

```text
WebSocket được dùng cho chat vì nó hỗ trợ giao tiếp hai chiều real-time giữa nhiều client.
```

### gRPC

```text
gRPC được dùng như internal service để các service trong hệ thống gọi nhau nhanh và có schema rõ ràng thông qua Protocol Buffers.
```

## 11. Lỗi Thường Gặp Và Cách Xử Lý

### Username đã tồn tại

Nếu register báo lỗi duplicate username/email:

```text
Đổi alice thành alice2 hoặc alice3.
```

### Port đang bị chiếm

Nếu server không start được vì port bận:

```powershell
Get-Process mangahub -ErrorAction SilentlyContinue | Stop-Process -Force
```

Sau đó chạy lại:

```powershell
.\mangahub.exe server start
```

### Quên token

Nếu gọi `/users/library` bị unauthorized, login lại:

```powershell
$login = Invoke-RestMethod -Method Post `
  -Uri http://localhost:8080/auth/login `
  -ContentType "application/json" `
  -Body '{"username":"alice","password":"password123"}'

$token = $login.token
```

### TCP/UDP/WebSocket không thấy output

Kiểm tra:

- Terminal server vẫn đang chạy.
- Listener phải chạy trước sender.
- Dùng đúng port mặc định.
- Không đóng nhầm terminal listener.

## 12. Checklist Cuối Trước Khi Nộp/Demo

- Build được `mangahub.exe`.
- `.\mangahub.exe server start` chạy đủ server.
- `/ping` trả `pong`.
- Register/login nhận JWT.
- Search manga có dữ liệu.
- Add library và update progress thành công.
- TCP listener nhận progress update.
- UDP listener nhận notification.
- WebSocket chat gửi nhận được giữa hai user.
- gRPC search/get/progress chạy được.
- Hai thành viên đều nắm phần mình trình bày.
