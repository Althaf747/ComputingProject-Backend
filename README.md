# Compro Backend API

A Gin-based Go backend for user authentication (JWT) and log management.

## Prerequisites
- Go 1.25+
- MySQL running locally (or reachable)
- Port `8080` available

## Quick Start (Windows)

```powershell
# Clone your repo (if not already)
# git clone <your-repo-url>
cd comp

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

### Logs (JWT Protected)
- `GET /api/logs` — List all logs
- `GET /api/logs/today?name=John` — Today logs, optional name filter
- `GET /api/logs/last-7-days?name=John` — Last 7 days, optional name
- `GET /api/logs/last-month?name=John` — Last month, optional name
- `GET /api/logs/date/{YYYY-MM-DD}?name=John` — Specific date, optional name
- `POST /api/logs` — Create log
- `DELETE /api/logs/{id}` — Hard delete log

### Utility
- `GET /health` — Healthcheck
- `GET /api/camera` — Webcam proxy
- `GET /ws` — WebSocket endpoint

## Request/Response Examples

### Register
```bash
curl -X POST http://localhost:8080/api/users/register \
  -H "Content-Type: application/json" \
  -d '{"username":"john","password":"password123","confirmPassword":"password123"}'
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

### Auth Header
Use JWT token in the `Authorization` header:
```text
Authorization: Bearer <JWT_TOKEN>
```

### Get Last 7 Days Logs
```bash
curl -X GET "http://localhost:8080/api/logs/last-7-days?name=John" \
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
  main.go
  openapi.yaml
```

## Development Tips
- Commit `go.mod` and `go.sum` to version control.
- For production, move DSN and JWT secret to environment variables.
- Ensure MySQL user has rights to create/alter tables for auto-migrations.
- Logs timestamps are stored as `TEXT` in ISO 8601; queries compare strings.

## License
No license specified. Add one if you plan to share/distribute.
