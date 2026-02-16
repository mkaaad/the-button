# The Button Frontend

React + TypeScript + Vite frontend for the The Button game.

## Tech

- React 19
- Vite 7
- Tailwind CSS
- shadcn-style UI components (Radix + class-variance-authority)

## Features

- SMS login flow
  - `GET /sms/captcha`
  - `POST /sms/code`
  - `POST /sms/verify`
- Cookie-based session reuse for WebSocket
- Real-time game panel
  - connect/disconnect
  - 60s countdown
  - 5s press cooldown
  - live feed
  - leaderboard

## Run

```bash
npm install
npm run dev
```

Frontend runs on `http://localhost:5173` by default.

## Backend URL

Set API base URL with env var:

```bash
VITE_API_BASE_URL=http://localhost:8080
```

When omitted, it defaults to `http://localhost:8080`.
