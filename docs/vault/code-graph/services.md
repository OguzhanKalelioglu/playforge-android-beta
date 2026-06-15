---
tags: [code-graph, index, services]
---

# Services Index

5 binary, hepsi monorepo `apps/` altında.

## API Service — `apps/api`

> Gin REST API. Ana business logic. VPS'te çalışır.

```
apps/api/
├── cmd/
│   ├── server/main.go          # HTTP entry (auth, orders, tests, admin)
│   ├── worker/main.go          # Asynq job consumer
│   └── scheduler/main.go       # 14-day plan kayıt
├── internal/
│   ├── config/                 # Viper-based env loader
│   ├── db/postgres.go          # pgxpool init
│   ├── lib/                    # logger, redis, jwt, middleware, asynq logger
│   ├── middleware/             # AuthRequired, AdminOnly, Recovery
│   ├── repository/             # SQL: user, test, order, payment, session, tester
│   ├── service/                # auth, order, stripe_client, notification
│   ├── handler/                # Gin routes: auth, order, test, admin, activity
│   ├── worker/                 # Asynq handlers (test_start, engage, review, login_google)
│   ├── scheduler/              # 14-day plan registrar
│   └── model/                  # Paylaşılan DTO'lar (orchestrator ile aynı)
├── migrations/                 # 0001_init, 0002_idx, 0003_jobs, 0004_orders, 0005_stripe
├── queries/                    # (gelecekte sqlc için)
├── Dockerfile                  # Multi-target: server | worker | scheduler
├── Dockerfile.worker
├── Dockerfile.scheduler
└── go.mod
```

**Public endpoints**: `/api/v1/*`
**Internal endpoints**: `/api/v1/activity` (X-Activity-Token), `/api/v1/payments/stripe/webhook` (Stripe-Sig)
**Binary size**: 40MB
**Build**: `go build -o bin/api ./cmd/server`

## Orchestrator — `apps/orchestrator`

> Emulator farm controller. Appium UI otomasyonu. Mini PC'de çalışır.

```
apps/orchestrator/
├── cmd/orchestrator/main.go    # Entry
├── internal/
│   ├── config/                 # ENV loader
│   ├── adb/                    # ADB shell, screenshot, signin
│   ├── emulator/               # Pool (CAS ready↔busy), models, container
│   ├── container/              # Docker compose control
│   ├── lifecycle/              # Boot + health monitor + reset
│   ├── profile/                # Device fingerprint randomization
│   ├── health/                 # Periodic health check
│   ├── appium/                 # W3C WebDriver client (10 dosya)
│   ├── task/                   # opt_in, download, engage, review, login_google
│   ├── taskrunner/             # Queue + anti-detect + activity sink
│   ├── api/                    # HTTP: emulator + task endpoints
│   └── lib/                    # logger, helpers
├── Dockerfile
└── go.mod
```

**Public endpoints**: `/v1/emulators/*`, `/v1/tasks/*/start` (X-API-Token)
**Internal**: `/liveness`
**Binary size**: 21MB
**Build**: `go build -o bin/orchestrator ./cmd/orchestrator`

## Web (Next.js) — `apps/web`

> Müşteri + Admin dashboard. VPS'te çalışır.

```
apps/web/
├── app/
│   ├── layout.tsx              # Root + globals.css
│   ├── page.tsx                # Landing (Decide lane)
│   ├── (auth)/                 # Learn + Decide
│   │   ├── login/page.tsx
│   │   └── register/page.tsx
│   ├── dashboard/              # Customer (Monitor lane)
│   │   ├── layout.tsx
│   │   ├── page.tsx            # Tests list
│   │   ├── new/page.tsx        # New order (Configure lane)
│   │   ├── [testId]/page.tsx   # Test detail + activity timeline
│   │   ├── orders/[orderId]/
│   │   │   ├── pay/page.tsx
│   │   │   └── success/page.tsx
│   │   └── settings/page.tsx
│   ├── admin/                  # Admin (Operate + Compare lane)
│   │   ├── layout.tsx
│   │   ├── page.tsx            # Overview metrics
│   │   ├── orders/page.tsx
│   │   ├── tests/page.tsx
│   │   ├── testers/page.tsx
│   │   └── payments/page.tsx
│   └── legal/                  # Static: terms, privacy, refund
├── components/
│   ├── ui/                     # button, card, badge, input, label, status-badge, progress, empty-state, package-name, money
│   ├── site-header.tsx
│   └── site-footer.tsx
├── lib/
│   ├── utils.ts                # cn, formatCurrency (TR), formatDate
│   ├── format.ts               # money, date, package truncate, duration
│   ├── api.ts                  # API client (typed)
│   └── auth-server.ts          # getCurrentUser (server-side)
├── Dockerfile
├── next.config.mjs             # standalone output
├── tailwind.config.ts          # Brief-driven tokens
└── package.json                # Next 14, React Query, RHF, Zod, NextAuth
```

**Routes**: 18 (4 static, 14 dynamic)
**Build**: `pnpm build` → standalone output
**Lighthouse target**: 90+ a11y, 95+ perf

## Paylaşılan Kontrat

Orchestrator ↔ API arasındaki sözleşme `apps/api/internal/model/` ve `apps/orchestrator/internal/model/` arasında simetriktir:

- `TestStartPayload`, `DailyEngagementPayload`, `WriteReviewPayload`, `LoginGooglePayload`
- `ActivityEvent`, `EventType`, `TaskStatus`
- `JobID(testID, type, day)` — Asynq stable IDs

## İlgili

- [[services/api]]
- [[services/orchestrator]]
- [[services/worker]]
- [[services/scheduler]]
- [[services/web]]
- [[code-graph/endpoints]]
- [[code-graph/db-tables]]
