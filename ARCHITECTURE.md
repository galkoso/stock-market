# Stock Market Platform — Adjusted Architecture

This document adapts the Bloomberg-style spec to **this repository's current state** and defines the target architecture.

## Current Stack (kept)

| Layer | Choice | Notes |
|-------|--------|-------|
| Frontend | **Angular 21** | Angular 22 was skipped due to Node compatibility; patterns match Angular 22 (standalone, signals) |
| Backend | **Go 1.23 + Gin** | Evolving toward clean architecture without a full rewrite |
| Auth | **MongoDB + JWT** | Username/password, refresh cookie (CareerCoach pattern) — already implemented |
| Market data | **Finnhub** | REST + single shared WebSocket hub — already implemented |
| Charts | **TradingView widget** | Free embed — already implemented |

## Added Stack

| Layer | Choice | Purpose |
|-------|--------|---------|
| PostgreSQL | Relational data | Server-side watchlists, alerts, earnings cache, company profiles |
| Redis | Cache | Finnhub response cache, rate-limit protection |
| Docker Compose | Dev/prod | Mongo, Postgres, Redis, backend |

## Target Backend Layout

```
backend/
├── main.go
├── internal/
│   ├── config/
│   ├── auth/              # Mongo users + JWT (existing)
│   ├── mongo/             # User store (existing)
│   ├── database/          # PostgreSQL connection + migrations
│   ├── cache/             # Redis client
│   ├── provider/          # Market data abstraction
│   │   └── finnhub/       # Finnhub implementation
│   ├── repositories/      # Postgres data access
│   ├── services/          # Business logic
│   ├── handler/           # HTTP + WebSocket handlers
│   ├── finnhub/           # WS hub (existing, shared upstream connection)
│   ├── scheduler/         # Morning refresh + alert checks
│   └── middleware/
```

## API Surface

### Auth (existing)
- `POST /api/auth/register`
- `POST /api/auth/login`
- `GET /api/auth/refresh`
- `POST /api/auth/logout`

### Stocks (extended)
- `GET /api/stocks/search?q=` — symbol, name, exchange, industry
- `GET /api/stocks/quotes?symbols=` — batch quotes
- `GET /api/stocks/:symbol` — full details (price, OHLC, market cap, profile)

### Earnings
- `GET /api/earnings?from=&to=` — earnings calendar
- `GET /api/watchlist/earnings?window=3|7|14` — upcoming earnings for user's watchlist

### News & Filings
- `GET /api/news/:symbol`
- `GET /api/filings/:symbol` — 10-K, 10-Q, 8-K

### Watchlist (server-side, Postgres)
- `GET /api/watchlist`
- `POST /api/watchlist`
- `DELETE /api/watchlist/:symbol`

### Alerts (Postgres)
- `GET /api/alerts`
- `POST /api/alerts`
- `DELETE /api/alerts/:id`

### Real-time (existing)
- `GET /ws/stocks?symbols=` — price ticks via shared Finnhub connection

### Bonus
- `GET /api/movers` — top gainers/losers (watchlist or market)
- `GET /api/stocks/:symbol/recommendations` — analyst consensus

## Frontend Pages

| Route | Page |
|-------|------|
| `/login` | Auth |
| `/` | Dashboard — overview, watchlist, upcoming earnings, alerts |
| `/search` | Stock search |
| `/stock/:symbol` | Details — chart, news, earnings, filings |
| `/earnings` | Earnings calendar with filters |
| `/watchlist` | Manage watchlist + earnings countdown |
| `/alerts` | Create/manage alerts |

## Provider Abstraction

Business logic depends on `provider.MarketDataProvider`, not Finnhub directly. Future providers (FMP, Polygon, Alpha Vantage) can be swapped without changing handlers.

## Caching Strategy (Redis)

| Key pattern | TTL | Data |
|-------------|-----|------|
| `profile:{symbol}` | 24h | Company profile |
| `earnings:{from}:{to}` | 6h | Earnings calendar |
| `news:{symbol}` | 15m | Company news |
| `filings:{symbol}` | 24h | SEC filings |
| `quote:{symbol}` | 30s | Quote snapshot |

## Scheduler (daily morning jobs)

1. Refresh earnings calendar cache
2. Sync watchlist company profiles
3. Evaluate alert conditions (earnings in X days, price targets)

## Future Roadmap

- OpenAPI/Swagger docs
- Full test suite
- Migrate users Mongo → Postgres (optional)
- TradingView remains widget-based; owned OHLCV via provider later
- SEC EDGAR direct integration as alternate filings provider
