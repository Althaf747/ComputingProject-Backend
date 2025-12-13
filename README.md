# Compro Backend API

A Gin-based Go backend for user authentication (JWT), approval workflow, password reset requests, and log management.

## Prerequisites
- Go 1.25+
- MySQL running locally (or reachable)
- Port `8080` available

## Quick Start (Windows)

```powershell
# Clone your repo (if not already)
# git clone <your-repo-url>
cd ComputingProject-Backend

# Ensure dependencies are present
go mod tidy

# Start MySQL and create the database (adjust credentials as needed)
# Example using mysql CLI
# mysql -u root -p -e "CREATE DATABASE IF NOT EXISTS comprodb CHARACTER SET utf8mb4;"

# Run the server
go run .\main.go
```

Server runs at: `http://localhost:8080`

## Configuration
Database DSN is currently hardcoded in `config/db.go`:

```go
dsn := "root:@tcp(127.0.0.1:3306)/comprodb?charset=utf8mb4&parseTime=True&loc=Local"
```

Update it to match your MySQL credentials or refactor to read from environment variables.

JWT secret is hardcoded in `middleware/auth.go`:

```go
var JWTSecret = []byte("your-secret-key-change-this-in-production")
```

Change this for production. Optionally load from env.

## API Overview

OpenAPI spec is available at `openapi.yaml` for full details.

### Auth
- `POST /api/users/register` — Create user
- `POST /api/users/login` — Login, returns JWT
- `POST /api/users/logout` — Client should discard token
- `POST /api/users/reset_request` — User requests password reset (flags account `needReset=true`, awaits verifier)
#### Auth (JWT Protected, Protected Route, Login required)
- `GET /api/users/pending` — List pending signups and reset requests (auth, verifier)
- `POST /api/users/approve?id={userId}` — Verifier approves user or reset; optional body `{ "role": "verifier" }`

### Logs (JWT Protected, Protected Route, Login required)
- `GET /api/logs` — List all logs
- `GET /api/logs/filter?period=today&name=John` — Filter logs by period (today, date, or range)
  - **Today**: `?period=today` or no parameters (default)
  - **Specific date**: `?period=date&date=YYYY-MM-DD`
  - **Date range**: `?period=range&start=YYYY-MM-DD&end=YYYY-MM-DD`
  - **With name filter**: Add `&name=John` to any query
- `POST /api/logs` — Create log
- `DELETE /api/logs/{id}` — Hard delete log

### Utility
- `GET /health` — Healthcheck
- `GET /api/camera` — Webcam proxy (path matches current route config)
- `GET /ws` — WebSocket endpoint

## Request/Response Examples

### Register
```bash
curl -X POST http://localhost:8080/api/users/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"password123","ConfirmPassword":"password123"}'
```

### Login
```bash
curl -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"password123"}'
```
Response:
```json
{
  "message": "Login successful",
  "token": "<JWT_TOKEN>",
  "data": {"id":1, "username":"john", "role":"user"}
}
```

### Reset Password Request
```bash
curl -X POST http://localhost:8080/api/users/reset_request \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"newpass","confirmPassword":"newpass"}'
```
Response:
```json
{ "message": "wait for verificator approval" }
```

### Pending / Approvals (verifier only)
List pending users and reset requests:
```bash
curl -H "Authorization: Bearer <JWT_TOKEN>" \
  http://localhost:8080/api/users/pending
```

Approve user/reset:
```bash
curl -X POST "http://localhost:8080/api/users/approve?id=5" \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"role":"user"}'
```

### Auth Header
Use JWT token in the `Authorization` header:
```text
Authorization: Bearer <JWT_TOKEN>
```

### Get Filtered Logs

**Today's logs:**
```bash
curl -X GET "http://localhost:8080/api/logs/filter?period=today" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**Specific date:**
```bash
curl -X GET "http://localhost:8080/api/logs/filter?period=date&date=2025-12-12" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**Date range:**
```bash
curl -X GET "http://localhost:8080/api/logs/filter?period=range&start=2025-12-01&end=2025-12-10" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

**With name filter:**
```bash
curl -X GET "http://localhost:8080/api/logs/filter?period=today&name=John" \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

### Create Log
```bash
curl -X POST http://localhost:8080/api/logs \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "authorized": false,
    "confidence": 0.91,
    "name": "Unknown",
    "role": "Guest",
    "timestamp": "2025-12-12T06:22:27"
  }'
```

### Delete Log
```bash
curl -X DELETE http://localhost:8080/api/logs/1 \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

## Project Structure
```
comp/
  controllers/
    log_controller.go
    user_controller.go
    connector.go
  middleware/
    auth.go
  models/
    log_model.go
    user_model.go
  routes/
    routes.go
  config/
    db.go
  go.mod
  go.sum
  main.go
  openapi.yaml
```

## License
Made to fulfill the requirements of the Computing Project course
