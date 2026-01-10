# FaceGate AI Backend

Go backend API untuk sistem keamanan pintu pintar Face Lock dengan pengenalan wajah, notifikasi real-time, dan manajemen pengguna.

## Fitur Utama

- **Autentikasi Pengguna** - JWT-based auth dengan role-based access control (pending, user, verifier)
- **Log Deteksi Wajah** - Simpan dan query event deteksi wajah dengan filter periode
- **Push Notifications** - Firebase Cloud Messaging untuk alert real-time ke perangkat mobile
- **Integrasi Kamera** - Proxy endpoint ke Python face recognition service
- **WebSocket** - Real-time event broadcasting untuk update langsung ke client
- **Approval Workflow** - Sistem persetujuan untuk registrasi user dan reset password

## Tech Stack

- **Go 1.25+** - Backend language
- **Gin** - HTTP web framework
- **GORM** - ORM untuk MySQL
- **MySQL 8.0+** - Database
- **Firebase Admin SDK** - Push notifications
- **Gorilla WebSocket** - Real-time communication
- **JWT** - Authentication tokens

## Prerequisites

- Go 1.21+ (tested on 1.25)
- MySQL 8.0+
- Firebase project (untuk push notifications)
- Python face recognition service running di `localhost:5000`

## Quick Start

```bash
# Clone repository
git clone <repo-url>
cd ComputingProject-Backend

# Copy environment file
cp .env.example .env
# Edit .env dengan konfigurasi Anda

# Install dependencies
go mod tidy

# Create database
mysql -u root -p -e "CREATE DATABASE face_lock_backend CHARACTER SET utf8mb4;"

# Run server
go run main.go
```

Server akan berjalan di `http://192.168.18.8:8080` (sesuaikan IP di `main.go`).

## Konfigurasi

Semua konfigurasi dilakukan via environment variables. Lihat `.env.example`:

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_DSN` | MySQL connection string | `root:@tcp(127.0.0.1:3306)/face_lock_backend?charset=utf8mb4&parseTime=True&loc=Local` |
| `JWT_SECRET` | Secret key untuk JWT signing (min 32 chars) | `change-this-in-production-32chars` |
| `SERVICE_AUTH_TOKEN` | Token untuk service-to-service auth | - |
| `FIREBASE_SERVICE_ACCOUNT_PATH` | Path ke Firebase service account JSON | `firebase-service-account.json` |

## Struktur Project

```
ComputingProject-Backend/
├── config/
│   └── db.go                 # Database connection & auto-migration
├── controllers/
│   ├── camera_controller.go  # Proxy ke Python face recognition service
│   ├── log_controller.go     # CRUD log deteksi wajah
│   └── user_controller.go    # Auth, register, login, approval
├── middleware/
│   └── auth.go               # JWT & service token authentication
├── models/
│   ├── log_model.go          # Model Log (deteksi wajah)
│   └── user_model.go         # Model User
├── routes/
│   └── routes.go             # Route definitions
├── services/
│   └── firebase.go           # Firebase FCM push notifications
├── utils/
│   └── websocket.go          # WebSocket manager & handler
├── .env.example              # Template environment variables
├── firebase-service-account.json  # Firebase credentials (gitignored)
├── go.mod                    # Go module dependencies
├── go.sum                    # Dependency checksums
├── main.go                   # Entry point
├── openapi.yaml              # OpenAPI 3.0 specification
└── README.md                 # Dokumentasi ini
```

## API Endpoints

### Authentication (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/users/register` | Register user baru (status: pending) |
| `POST` | `/api/users/login` | Login, returns JWT token |
| `POST` | `/api/users/logout` | Logout (client discard token) |
| `POST` | `/api/users/reset_request` | Request reset password |

### User Management (Auth Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/users/pending` | List user pending & need reset |
| `POST` | `/api/users/approve?id={id}&action={approve\|reject}` | Approve/reject user (verifier only) |
| `POST` | `/api/users/fcm-token` | Update FCM token untuk push notification |

### Logs (Auth Required)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/logs` | Get semua log |
| `GET` | `/api/logs/filter` | Filter log by period/name |
| `POST` | `/api/logs` | Create log entry |
| `DELETE` | `/api/logs/:id` | Delete log (hard delete) |

**Filter Parameters untuk `/api/logs/filter`:**
- `period` - `today` (default), `date`, atau `range`
- `date` - Format `YYYY-MM-DD` (required jika period=date)
- `start` & `end` - Format `YYYY-MM-DD` (required jika period=range)
- `name` - Filter by nama visitor

### Camera (Proxy ke Python Service)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/camera/stream` | MJPEG video stream |
| `GET` | `/api/camera/snapshot` | Single frame capture |
| `GET` | `/api/camera/events` | SSE stream untuk detection events |
| `GET` | `/api/camera/status` | Status kamera |
| `POST` | `/api/camera/start` | Start kamera |
| `POST` | `/api/camera/stop` | Stop kamera |
| `GET` | `/api/camera/config` | Get konfigurasi kamera |
| `POST` | `/api/camera/config` | Update konfigurasi kamera |
| `GET` | `/api/camera/zones` | Get detection zones |
| `POST` | `/api/camera/zones` | Update detection zones |
| `POST` | `/api/camera/test/droidcam` | Setup DroidCam |
| `POST` | `/api/camera/test/rtsp` | Setup RTSP stream |

### Face Management (Proxy ke Python Service)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/faces` | List enrolled face users |
| `POST` | `/api/faces/enroll` | Enroll user baru dengan images |
| `POST` | `/api/faces/enroll/capture` | Enroll dari single capture |
| `DELETE` | `/api/faces/:name` | Delete user dari face database |
| `POST` | `/api/faces/:name/add-sample` | Tambah face sample ke user |

### Utility

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/health` | Health check |
| `GET` | `/api/ws` | WebSocket connection |

## Data Models

### User

```json
{
  "id": 1,
  "username": "johndoe",
  "role": "user",        // pending | user | verifier | rejected
  "needReset": false
}
```

### Log

```json
{
  "id": 1,
  "authorized": true,
  "confidence": 0.95,
  "name": "John Doe",
  "role": "user",
  "timestamp": "2025-12-18T07:30:00"
}
```

## Authentication Flow

### 1. Register
```bash
POST /api/users/register
{
  "username": "johndoe",
  "password": "password123",
  "confirmPassword": "password123"
}
```
User akan mendapat status `pending` dan menunggu approval dari verifier.

### 2. Login
```bash
POST /api/users/login
{
  "username": "johndoe",
  "password": "password123"
}
```
Response:
```json
{
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "data": { "id": 1, "username": "johndoe", "role": "user" }
}
```

### 3. Authenticated Request
```bash
GET /api/logs
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### 4. Service-to-Service Auth
Untuk internal service (Python face recognition), gunakan `SERVICE_AUTH_TOKEN`:
```bash
POST /api/logs
Authorization: Bearer your-service-token-here
```

## Push Notifications

Ketika wajah terdeteksi, push notification dikirim ke semua device yang terdaftar:

- **Known person**: "Access Granted - {Name} was detected at the door"
- **Unknown person**: "Unknown Person Detected - An unknown person was detected at the door"

### Setup Firebase

1. Buka [Firebase Console](https://console.firebase.google.com)
2. Buat project baru atau pilih existing project
3. Project Settings → Service Accounts
4. Generate new private key
5. Simpan sebagai `firebase-service-account.json` di project root
6. Pastikan file ini ada di `.gitignore`

## WebSocket

WebSocket endpoint di `/api/ws` untuk real-time event broadcasting.

### Connect
```javascript
const ws = new WebSocket('ws://192.168.18.8:8080/api/ws');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Detection event:', data);
};
```

### Event Format
```json
{
  "type": "detection",
  "data": {
    "name": "John Doe",
    "authorized": true,
    "confidence": 0.95,
    "timestamp": "2025-12-18T07:30:00"
  }
}
```

### Send Log via WebSocket
Client juga bisa mengirim log entry via WebSocket:
```javascript
ws.send(JSON.stringify({
  "authorized": true,
  "confidence": 0.95,
  "name": "John Doe",
  "role": "user",
  "timestamp": "2025-12-18T07:30:00"
}));
```

## Role-Based Access Control

| Role | Description | Permissions |
|------|-------------|-------------|
| `pending` | User baru menunggu approval | Tidak bisa login |
| `user` | User biasa | Akses logs, update FCM token |
| `verifier` | Admin/verifier | Approve/reject user, semua akses user |
| `rejected` | User ditolak | Tidak bisa login |
| `service` | Internal service | Full access via service token |

## Error Responses

Semua error menggunakan format standar:
```json
{
  "error": "Error message description"
}
```

HTTP Status Codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `500` - Internal Server Error
- `502` - Bad Gateway (Python service unavailable)

## Development

### Run dengan hot reload (opsional)
```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run
air
```

### Build untuk production
```bash
go build -o face-lock-backend main.go
./face-lock-backend
```

## API Documentation

OpenAPI 3.0 specification tersedia di `openapi.yaml`. Import ke Swagger UI atau Postman untuk interactive documentation.

## License

Computing Project - Telkom University
