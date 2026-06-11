# Stock Market Monitor

Personal Bloomberg-style dashboard for monitoring stocks, earnings, news, filings, and live market activity.

## Stack (adjusted to this repo)

| Layer | Technology |
|-------|------------|
| Frontend | Angular 21 (standalone, signals, Material) |
| Backend | Go 1.23+ / Gin |
| Auth | MongoDB + JWT (refresh cookie) |
| App data | PostgreSQL (watchlists, alerts) |
| Cache | Redis (Finnhub response cache) |
| Market data | Finnhub REST + shared WebSocket hub |
| Charts | TradingView free widget |
| Containers | Docker Compose |

See [ARCHITECTURE.md](./ARCHITECTURE.md) for the full adjusted spec and roadmap.

## Prerequisites

- Finnhub API key → [finnhub.io](https://finnhub.io/)
- Node.js 20+
- Go 1.23+
- Docker (for Mongo, Postgres, Redis)

## Quick start

### 1. Start infrastructure

```bash
npm run docker:up
```

This starts MongoDB, PostgreSQL, and Redis.

### 2. Configure backend

```bash
cd backend
cp .env.example .env   # add FINNHUB_API_KEY
```

### 3. Run the app

```bash
npm run install:all   # first time
npm start
```

Open **http://localhost:4201** → register/login.

## API overview

### Auth
- `POST /api/auth/register` `{ username, password }`
- `POST /api/auth/login`
- `GET /api/auth/refresh` (cookie)
- `POST /api/auth/logout`

### Stocks
- `GET /api/stocks/search?q=`
- `GET /api/stocks/quotes?symbols=`
- `GET /api/stocks/:symbol` — full details + profile
- `GET /api/market/search?q=` — search with exchange/industry

### Earnings & news
- `GET /api/earnings?from=&to=`
- `GET /api/watchlist/earnings?window=3|7|14`
- `GET /api/news/:symbol`
- `GET /api/filings/:symbol`
- `GET /api/stocks/:symbol/recommendations`
- `GET /api/movers`

### Watchlist (Postgres, per user)
- `GET /api/watchlist`
- `POST /api/watchlist` `{ symbol, companyName }`
- `DELETE /api/watchlist/:symbol`

### Alerts
- `GET /api/alerts`
- `POST /api/alerts` `{ symbol, alertType, params }`
- `DELETE /api/alerts/:id`

### Real-time
- `GET /ws/stocks?symbols=AAPL,TSLA` — shared Finnhub connection

## Frontend pages

| Route | Page |
|-------|------|
| `/` | Dashboard — watchlist, live prices, chart |
| `/search` | Stock search |
| `/stock/:symbol` | Details — chart, news, earnings, filings |
| `/earnings` | Earnings calendar |
| `/watchlist` | Server-side watchlist + earnings countdown |
| `/alerts` | Create/manage alerts |

## Development notes

- **Angular 21** is used instead of 22 for Node compatibility.
- **MongoDB** stores users; **PostgreSQL** stores watchlists/alerts.
- **Redis** caches Finnhub responses to reduce API usage.
- Finnhub free tier: 1 WS connection, 50 symbols, 60 REST calls/min.
- Some Finnhub endpoints (filings, etc.) may return 403 on the free plan.

## Docker

```bash
docker compose up -d          # infra only
docker compose up --build     # infra + backend image
```
