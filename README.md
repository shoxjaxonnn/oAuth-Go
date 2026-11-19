# oAuth Demo (Go backend)

Same authentication demo as [oAuth](https://github.com/shoxjaxonnn/oAuth), with the backend rewritten in **Go** (stdlib `net/http`, Go 1.22+ pattern routing) instead of Node.js — same 4 sign-in methods, same security model.

**Frontend:** React + TypeScript + Vite (unchanged)
**Backend:** Go, stdlib `net/http`, SQLite (`modernc.org/sqlite`, pure Go, no CGO)

---

## Features

### 4 sign-in methods
| Method | Flow |
|---|---|
| **Google** | Hand-rolled OAuth 2.0 against Google's HTTP endpoints — no SDK |
| **Facebook** | Hand-rolled OAuth 2.0 against the Graph API — no SDK |
| **Telegram** | 6-digit OTP via the Telegram Bot API, 5 min TTL, max 5 attempts |
| **Email/Password** | Firebase Authentication REST API — no Admin SDK |

### Security
- Access token (15m, JSON body) + refresh token (30d, **httpOnly cookie**) with rotation on every `/auth/refresh` call
- Rate limiting per IP on login/register/OTP endpoints (hand-written fixed-window limiter, no dependency)
- Generic login error — never reveals whether an email is registered
- Account linking by email across providers
- Firebase email verification on register + password-reset flow

### Engineering
- SQLite persistence (users, identities, refresh tokens)
- **20 automated tests** (`go test ./...`) across JWT, account linking, refresh rotation, OTP, and HTTP handlers (`net/http/httptest`), with outbound calls faked via a custom `http.RoundTripper` — no real network hit during tests
- OpenAPI/Swagger docs at `/docs` (spec embedded in the binary with `go:embed`)
- GitHub Actions CI: `go vet`, `go build`, `go test` for the backend; TypeScript check + build for the frontend

---

## Project structure

```
oAuth-Go/
├── backend/
│   ├── main.go
│   └── internal/
│       ├── config/      # env loading
│       ├── db/          # SQLite open + migrations
│       ├── store/       # users/identities (account linking) + OTP store
│       ├── auth/        # JWT, refresh token rotation, session/cookie helpers
│       ├── handlers/     # google, facebook, telegram, firebase, session
│       ├── middleware/   # auth, rate limit, CORS
│       ├── docs/        # embedded OpenAPI spec + Swagger UI
│       └── server/      # route wiring + integration tests
└── frontend/            # same React+TS UI as the Node version
```

---

## Setup

Copy `backend/.env.example` to `backend/.env` and fill in the same credentials described in the [Node version's README](https://github.com/shoxjaxonnn/oAuth#1-credentials) (Google/Facebook OAuth apps, a Telegram bot token, a Firebase Web API key).

```bash
# Backend
cd backend
go run .            # http://localhost:4000

# Frontend (new terminal)
cd frontend
npm install
npm run dev          # http://localhost:5173
```

## Testing

```bash
cd backend
go vet ./...
go test ./...        # 20 tests, no real network calls
```

Manual testing follows the same steps as the [Node version](https://github.com/shoxjaxonnn/oAuth#manual-testing--each-sign-in-method): try each provider, confirm the httpOnly `refresh_token` cookie, clear `localStorage` and reload to see silent refresh, hammer "Send code" to see `429`, and link accounts by signing in with the same email through two different providers.
