# The Button - Real-time Game Challenge

> Chinese version: [README.zh-CN.md](README.zh-CN.md)

A full-stack real-time button game with leaderboard ranking. Players compete to press a virtual button that resets a 60-second countdown, aiming for the lowest remaining time on the leaderboard.

## Features

- **SMS Authentication** - Phone verification via Alibaba Cloud SMS with captcha
- **Real-time WebSocket** - Live game state updates and leaderboard broadcasts
- **Leaderboard System** - Redis-sorted ranking with LT (less than) updates
- **Player Cooldown** - 5-second lock per user to prevent spamming
- **Game Time Window** - Configurable start/end times with timezone support
- **Modern UI** - Responsive React interface with Tailwind CSS and Radix components
- **Docker Support** - Complete containerized deployment with Redis persistence

## Tech Stack

**Frontend**
- React 19 + TypeScript
- Vite 7
- Tailwind CSS + shadcn-style UI components (Radix UI)
- WebSocket API

**Backend**
- Go 1.24.4 + Gin web framework
- SQLite (GORM) + Redis (sorted sets)
- Alibaba Cloud SMS integration
- Gorilla WebSocket

**Infrastructure**
- Docker Compose
- Nginx (frontend serving)
- Redis (leaderboard & session storage)

## Quick Start

### Option 1: Docker Compose (Recommended)
```bash
# Copy environment examples
cp be/.env.example be/.env
cp fe/.env.example fe/.env

# Edit be/.env with your SMS credentials
# Edit fe/.env if needed (defaults work with Docker)

# Start all services
docker-compose up
```

Access the game at `http://localhost:5173`

### Option 2: Manual Setup
```bash
# 1. Start Redis
docker run -d -p 6379:6379 redis:7-alpine

# 2. Configure backend
cd be
cp .env.example .env
# Edit .env with SMS credentials and Redis settings

# 3. Run backend
go run cmd/main.go

# 4. Configure frontend
cd ../fe
cp .env.example .env
# Set VITE_API_BASE_URL=http://localhost:8080
# Set VITE_WS_URL=ws://localhost:8080/ws

# 5. Run frontend
npm install
npm run dev
```

## Environment Variables

### Backend (`be/.env`)
| Variable | Description | Required |
|----------|-------------|----------|
| `ACCESS_KEY_ID` | Alibaba Cloud SMS API key | Yes |
| `ACCESS_KEY_SECRET` | Alibaba Cloud SMS API secret | Yes |
| `REDIS_ADDR` | Redis connection (host:port) | Yes |
| `REDIS_PASSWORD` | Redis password (if set) | No |
| `REDIS_DB` | Redis database index | No (default: 0) |
| `GAME_START_TIME` | Game start (Asia/Shanghai) | Yes |
| `GAME_END_TIME` | Game end (Asia/Shanghai) | Yes |
| `GAME_TIMEZONE` | Timezone for game times | No (default: Asia/Shanghai) |

### Frontend (`fe/.env`)
| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_BASE_URL` | API base URL | `/api` |
| `VITE_WS_URL` | WebSocket endpoint | `/ws` |

## API Endpoints

### HTTP API
- `GET /sms/captcha` - Get captcha image for SMS verification
- `POST /sms/code` - Send SMS verification code (requires captcha)
- `POST /sms/verify` - Verify SMS code and create session

### WebSocket
Connect to `ws://host:port/ws?session_id={session_id}`

#### Commands (client → server)
- `"1"` - Request current time
- `"2"` - Request leaderboard
- `"3"` - Press button (requires authenticated session)

#### Messages (server → client)
```json
{"type": "time", "data": {"time": 1676543210000}}
{"type": "leaderboard", "data": {"entries": [...]}}
{"type": "button_press", "data": {"username": "player1"}}
{"type": "lock"}  // 5-second cooldown active
{"type": "pending"}  // Game hasn't started
{"type": "finished"} // Game has ended
{"type": "unauthorized"} // Invalid session
```

## Game Rules

1. **60-Second Countdown** - Button press resets timer to 60 seconds
2. **Leaderboard Ranking** - Players ranked by remaining time (lower = better)
3. **5-Second Cooldown** - Each player must wait 5 seconds between presses
4. **Time Window** - Game only active between `GAME_START_TIME` and `GAME_END_TIME`
5. **SMS Authentication** - Phone verification required to play

## Development

### Project Structure
```
the-button/
├── fe/                 # React frontend
│   ├── src/
│   │   ├── components/ui/  # Radix-based UI components
│   │   ├── lib/utils.ts    # Utility functions
│   │   └── App.tsx         # Main application
│   └── package.json
├── be/                 # Go backend
│   ├── cmd/main.go          # Entry point
│   ├── router/router.go     # HTTP routes
│   ├── api/                 # HTTP & WebSocket handlers
│   ├── service/             # Business logic
│   ├── dao/                 # Data access (Redis, SQLite)
│   ├── model/               # Data models
│   ├── config/              # Configuration
│   ├── errorx/              # Error definitions
│   ├── middleware/          # CORS middleware
│   └── connx/               # Connection pooling
└── docker-compose.yml
```

### Development Commands
```bash
# Frontend
cd fe
npm run dev      # Development server
npm run build    # Production build
npm run lint     # ESLint

# Backend
cd be
go run cmd/main.go
go test ./...
go mod tidy

# Load testing
cd be/test
go run main.go   # Requires backend running
```

## Deployment

### Docker Compose Production
```bash
docker-compose up -d
```

Services:
- **Redis** (port 6379) - Leaderboard and session storage
- **Backend** (port 8080) - Go application with health check
- **Frontend** (port 5173) - Nginx serving React build

### Health Checks
- Backend: `GET /sms/captcha` (returns 200 if healthy)
- Redis: Automatic health check in docker-compose

## License

MIT