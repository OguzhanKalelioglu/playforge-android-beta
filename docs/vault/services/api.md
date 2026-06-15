---
tags: [service, api, go]
---

# API Service

> Gin REST API. VPS'te çalışır. Tüm business logic burada.

## Sorumluluk

- Auth (JWT + refresh token)
- Plan + Order + Payment (Stripe Checkout)
- Test + Assignment + Activity (müşteri + admin)
- Admin metrics + monitoring
- Activity ingest (orchestrator'dan)

## Entry Point

`apps/api/cmd/server/main.go` — startup sırası:
1. Config load (Viper)
2. Logger (zap)
3. PostgreSQL pool (pgx)
4. Redis client
5. Asynq client
6. JWT manager
7. Repository + Service + Handler construction
8. Gin route registration
9. HTTP server (`:8080`)

## Konfigürasyon

Env-based (`.env` veya Docker env). Viper ile yüklenir.

```
APP_ENV, PORT, DATABASE_URL, REDIS_URL, JWT_SECRET
STRIPE_SECRET_KEY, STRIPE_WEBHOOK_SECRET
SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASSWORD
ORCHESTRATOR_API_TOKEN, DOMAIN, LOG_LEVEL
```

## Bağımlılıklar

```
github.com/gin-gonic/gin           # HTTP router
github.com/hibiken/asynq          # Job queue
github.com/jackc/pgx/v5           # PostgreSQL
github.com/redis/go-redis/v9      # Redis
github.com/stripe/stripe-go/v82   # Stripe (yakında)
go.uber.org/zap                   # Logger
github.com/spf13/viper            # Config
golang.org/x/crypto/bcrypt        # Password
github.com/golang-jwt/jwt/v5      # JWT
github.com/google/uuid            # UUID
```

## Middleware Sırası

1. `gin.Recovery()` — panic recovery
2. `lib.RequestLogger(logger)` — structured request log
3. `cors.New(corsConfig)` — dev: localhost:3000, prod: domain
4. Route group'a göre: `AuthRequiredJWT()`, `AdminOnly()`

## Klasör Yapısı

Detay için: [[code-graph/services]]

## Test

Şu an unit test yok. Plan (sonra):
- `internal/service/auth_test.go` — bcrypt, JWT roundtrip
- `internal/service/order_test.go` — Stripe mock, mark paid
- `internal/handler/*_test.go` — httptest, in-memory repo

## İlgili

- [[architecture]]
- [[payment-flow]] — Stripe Checkout
- [[code-graph/endpoints]]
- [[code-graph/db-tables]]
- [[services/worker]] — Aynı monorepo, farklı binary
- [[services/scheduler]] — Aynı monorepo, farklı binary
