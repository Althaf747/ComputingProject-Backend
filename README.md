# Face Lock Backend

Go backend API for Face Lock smart door security system with face recognition, real-time notifications, and user management.

## Features

- **User Authentication** - JWT-based auth with role-based access control
- **Face Detection Logs** - Store and query face detection events
- **Push Notifications** - Firebase Cloud Messaging for real-time alerts
- **Camera Integration** - Proxy endpoints for Python face recognition service
- **WebSocket** - Real-time event broadcasting

## Prerequisites

- Go 1.21+
- MySQL 8.0+
- Firebase project (for push notifications)

## Quick Start

```bash
# Clone repository
git clone <repo-url>
cd ComputingProject-Backend

# Copy environment file
cp .env.example .env
# Edit .env with your configuration

# Install dependencies
go mod tidy

# Create database
mysql -u root -p -e "CREATE DATABASE face_lock_backend CHARACTER SET utf8mb4;"

# Run server
go run main.go
```

## Configuration

All configuration is done via environment variables. See `.env.example`:

| Variable | Description |
|----------|-------------|
| `DATABASE_DSN` | MySQL connection string |
| `JWT_SECRET` | Secret key for JWT signing (min 32 chars) |
| `SERVICE_AUTH_TOKEN` | Token for service-to-service auth |
| `FIREBASE_SERVICE_ACCOUNT_PATH` | Path to Firebase service account JSON |

## API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/users/register` | Register new user |
| POST | `/api/users/login` | Login, returns JWT |
| POST | `/api/users/logout` | Logout |
| POST | `/api/users/reset_request` | Request password reset |
| POST | `/api/users/fcm-token` | Update FCM token (auth required) |

### User Management (Auth Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/users/pending` | List pending users |
| POST | `/api/users/approve` | Approve/reject user |

### Logs (Auth Required)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/logs` | Get all logs |
| GET | `/api/logs/filter` | Filter logs by period/name |
| POST | `/api/logs` | Create log entry |
| DELETE | `/api/logs/:id` | Delete log |

### Camera (Proxied to Python Service)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/camera/stream` | MJPEG video stream |
| GET | `/api/camera/snapshot` | Single frame |
| GET | `/api/camera/status` | Camera status |
| POST | `/api/camera/start` | Start camera |
| POST | `/api/camera/stop` | Stop camera |

### Utility
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/ws` | WebSocket connection |

## Project Structure

```
├── config/         # Database configuration
├── controllers/    # Request handlers
├── middleware/     # Auth middleware
├── models/         # Database models
├── routes/         # Route definitions
├── services/       # Firebase service
├── utils/          # WebSocket manager
├── main.go         # Entry point
└── openapi.yaml    # API specification
```

## Push Notifications

When a face is detected, push notifications are sent to all registered devices:

- **Known person**: "Access Granted - {Name} was detected"
- **Unknown person**: "⚠️ Unknown Person Detected"

### Setup Firebase

1. Go to [Firebase Console](https://console.firebase.google.com)
2. Project Settings → Service Accounts
3. Generate new private key
4. Save as `firebase-service-account.json` in project root

## License

Computing Project - Universitas Multimedia Nusantara
