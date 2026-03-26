# The Button Project - Agent Guide

## Project Overview
- Full-stack real-time button game with leaderboard
- Frontend: React + TypeScript + Vite + Tailwind CSS + Radix UI
- Backend: Go (Gin, GORM, SQLite, Redis, Alibaba Cloud SMS, gorilla/websocket)
- Containerized with Docker Compose

## Essential Commands

### Frontend (`fe/`)
- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run lint` - Run ESLint
- `npm run preview` - Preview production build

### Backend (`be/`)
- `go run cmd/main.go` - Run backend server
- `go test ./...` - Run tests
- `go mod tidy` - Clean up dependencies
- Load test: `cd be/test && go run main.go` (requires backend running)

### Docker
- `docker-compose up` - Start all services (Redis, backend, frontend)
- `docker-compose down` - Stop services
- See `docker-compose.yml` for environment variables

## Local Development Setup

### Without Docker (requires Redis)
1. Start Redis: `docker run -d -p 6379:6379 redis:7-alpine`
2. Backend: `cd be && go run cmd/main.go`
3. Frontend: `cd fe && npm run dev`

### With Docker Compose
- Use `docker-compose up` as described above

## Code Organization

### Frontend structure
- `src/components/ui/` - Radix-based UI components
- `src/lib/utils.ts` - Utility functions (`cn`, etc.)
- `App.tsx` - Main application logic
- Uses environment variables: `VITE_API_BASE_URL`, `VITE_WS_URL`

### Backend structure
- `cmd/main.go` - Entry point
- `router/router.go` - HTTP routes
- `api/` - HTTP handlers and WebSocket
- `service/` - Business logic (button, SMS)
- `dao/` - Data access (Redis, SQLite)
- `model/` - Data models
- `config/` - Configuration loading
- `errorx/` - Error definitions
- `middleware/` - CORS middleware
- `connx/` - Connection pooling

## Development Patterns

### Go patterns
- Uses Gin for HTTP routing
- GORM for SQLite ORM
- Redis for leaderboard and locking
- Atomic operations for shared state
- WebSocket messages: `type` + `data` JSON
- Error handling via custom error types in `errorx/`

### TypeScript patterns
- Functional components with hooks
- Custom hooks pattern for WebSocket
- Utility functions for parsing API responses
- Environment variable validation
- Local storage for session persistence

## Code Style

### Go
- Use `gofmt` (standard Go formatting)
- No specific linter beyond `go vet`
- Error variables named with `Err` prefix (see `errorx/`)
- Package organization follows domain separation (api, service, dao, etc.)

### TypeScript/React
- ESLint with TypeScript recommended rules
- Functional components with `const` arrow functions
- Tailwind CSS for styling, using `cn` utility for conditional classes
- Import order: external libraries first, then internal modules

## Common Development Tasks

### Adding a new API endpoint
1. Add route in `router/router.go`
2. Create handler in `api/` package (or extend existing)
3. Implement business logic in `service/`
4. Add error definitions in `errorx/` if needed
5. Update frontend `App.tsx` to call new endpoint

### Adding a new UI component
1. Create component in `src/components/ui/` following Radix patterns
2. Export from `src/components/ui/index.ts` (if exists)
3. Use Tailwind classes with `cn` for variants
4. Add to `App.tsx` or relevant feature

### Modifying the leaderboard algorithm
- See `service/button.go` `PressButton` and `GetLeaderboard`
- Leaderboard stored in Redis sorted set `button_leaderboard`
- Scoring uses `countdownTime - (now - prev)` (lower is better)

## Testing
- Backend: Standard Go testing
- Load testing: Separate `test/` directory with its own `go.mod`
- Frontend: No test framework observed

## Environment Configuration

### Backend (`.env` in `be/`)
| Variable | Purpose | Example |
|----------|---------|---------|
| `ACCESS_KEY_ID` | Alibaba Cloud SMS API key | `your_access_key_id` |
| `ACCESS_KEY_SECRET` | Alibaba Cloud SMS API secret | `your_access_key_secret` |
| `REDIS_ADDR` | Redis connection address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password (optional) | |
| `REDIS_DB` | Redis database index | `0` |
| `GAME_START_TIME` | Game start timestamp (Asia/Shanghai) | `2026-02-16 20:00:00` |
| `GAME_END_TIME` | Game end timestamp (Asia/Shanghai) | `2026-02-16 23:59:59.999` |
| `GAME_TIMEZONE` | Timezone for game times | `Asia/Shanghai` |

### Frontend (`.env` in `fe/`)
| Variable | Purpose | Example |
|----------|---------|---------|
| `VITE_API_BASE_URL` | Base URL for API requests | `/api` |
| `VITE_WS_URL` | WebSocket endpoint path | `/ws` |

- Copy `.env.example` to `.env` and set appropriate values.
- Backend reads from environment variables via `config.InitConfig()`.

## Gotchas & Important Notes
1. Game time window configured via `GAME_START_TIME` and `GAME_END_TIME` in Asia/Shanghai timezone
2. SMS verification requires Alibaba Cloud credentials
3. WebSocket authentication via `session_id` query parameter
4. Button press lock: 5‑second cooldown per user via Redis
5. Leaderboard uses Redis sorted sets with LT (less than) update
6. Frontend uses Chinese locale for time display
7. Backend uses `atomic.Int64` for start time sharing

## Deployment
- Docker Compose for full stack
- Frontend served via nginx (see `nginx.conf`)
- Backend health check endpoint: `/sms/captcha`
- Redis persistence via volume