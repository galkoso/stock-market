# Stock Market Search

Full-stack MVP for searching live stock prices via [Finnhub](https://finnhub.io/). The Angular frontend talks to a Go backend that keeps the Finnhub API key server-side and proxies live trade data over WebSocket.

## Project structure

```
stock-market/
├── frontend/   # Angular 21 SPA
├── backend/    # Go API (Gin + gorilla/websocket)
└── README.md
```

## Prerequisites

- **Node.js** `^20.19.0`, `^22.12.0`, or `^24.0.0` (required by Angular 21)
- **Go** 1.23+
- **Finnhub API key** — sign up at [finnhub.io](https://finnhub.io/) and create a free key

## Setup

### 1. Backend

```bash
cd backend
cp .env.example .env   # then edit .env and add your Finnhub API key
go mod tidy
go run main.go
```

The API listens on `http://localhost:8080`.

Environment variables are loaded from `backend/.env` automatically.

Optional: set `PORT` in `.env` to override the default port (`8080`).

### 2. Frontend

In a second terminal:

```bash
cd frontend
npm install
npm start
```

The app opens at `http://localhost:4201` and proxies `/api` and `/ws` requests to the Go backend.

## API

### `GET /api/stocks/search?q={query}`

Looks up a symbol with Finnhub search, fetches the current quote, and returns:

```json
{
  "symbol": "AAPL",
  "companyName": "APPLE INC",
  "currentPrice": 261.74,
  "dailyChange": -0.25,
  "dailyChangePercent": -0.0954,
  "lastUpdated": "2026-06-07T12:00:00Z"
}
```

### `GET /ws/stocks?symbol={symbol}` (WebSocket)

Upgrades to a WebSocket connection for live trade updates. Angular connects to the Go backend only — never directly to Finnhub.

**Flow**

1. Client connects with a validated symbol.
2. Backend subscribes to Finnhub: `{"type":"subscribe","symbol":"AAPL"}`.
3. Backend forwards trade messages to the client.
4. On disconnect, backend unsubscribes if no other clients need that symbol.

**Example messages to the browser**

```json
{ "type": "status", "status": "connecting", "symbol": "AAPL" }
{ "type": "status", "status": "live", "symbol": "AAPL" }
{ "type": "trade", "status": "live", "symbol": "AAPL", "price": 261.74, "volume": 10, "timestamp": 1672348887195 }
{ "type": "status", "status": "disconnected", "symbol": "AAPL" }
```

During local development the frontend uses:

`ws://localhost:4201/ws/stocks?symbol=AAPL` (proxied to the Go backend)

You can also connect directly to the backend:

`ws://localhost:8080/ws/stocks?symbol=AAPL`

### REST error responses

| HTTP | Code | When |
|------|------|------|
| 400 | `MISSING_QUERY` | `q` parameter is empty |
| 404 | `SYMBOL_NOT_FOUND` | No matching symbol or quote |
| 502 | `FINNHUB_API_ERROR` | Finnhub request failed |
| 500 | `INTERNAL_ERROR` | Unexpected server error |

Startup fails immediately if `FINNHUB_API_KEY` is not set.

## Architecture

```
Angular (4201)
  │  REST  /api/stocks/search
  │  WS    /ws/stocks?symbol=AAPL
  ▼
Go backend (8080)
  │  REST  Finnhub /search, /quote
  │  WS    wss://ws.finnhub.io?token=...
  ▼
Finnhub API
```

- **REST search** — initial company lookup and fallback snapshot price.
- **WebSocket stream** — live trade prices after a stock is selected.
- **Shared Finnhub WS hub** — one backend connection to Finnhub, ref-counted symbol subscriptions, automatic reconnect.
- **Security** — `FINNHUB_API_KEY` stays on the server only.
- **CORS** — enabled for `http://localhost:4201` during local development.

## Local development commands

From the project root you can run both apps together:

```bash
npm run install:all   # first time only
npm start             # runs backend + frontend
```

Or run them separately:

```bash
npm run dev:backend
npm run dev:frontend
```

**Frontend**

```bash
cd frontend
npm install
npm start
```

**Backend**

```bash
cd backend
cp .env.example .env   # edit .env with your Finnhub API key
go mod tidy
go run main.go
```

## Example searches

Try: `Apple`, `AAPL`, `Tesla`, `TSLA`, `Microsoft`, `MSFT`.

After search, watch the connection badge move from **Connecting** → **Live** as trades arrive. Searching a new stock closes the previous WebSocket and opens a new one.
