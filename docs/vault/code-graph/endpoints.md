---
tags: [code-graph, http, api]
---

# HTTP Endpoints

## API Service (`apps/api/cmd/server`)

Base: `https://api.testerscomm.net`

### Public

| Method | Path | Auth | Handler | Amaç |
| --- | --- | --- | --- | --- |
| GET | `/health` | — | `HealthHandler.Health` | DB + Redis health |
| GET | `/liveness` | — | `HealthHandler.Liveness` | K8s-style liveness |
| GET | `/api/v1/ping` | — | inline | Smoke test |
| GET | `/api/v1/plans` | — | `OrderHandler.ListPlans` | Plan listesi (public) |
| POST | `/api/v1/auth/register` | — | `AuthHandler.Register_` | Hesap oluştur |
| POST | `/api/v1/auth/login` | — | `AuthHandler.Login` | JWT access + refresh |
| POST | `/api/v1/auth/refresh` | cookie/body | `AuthHandler.Refresh` | Token yenile |
| POST | `/api/v1/auth/logout` | cookie | `AuthHandler.Logout` | Session revoke |
| GET | `/api/v1/auth/me` | JWT | `AuthHandler.Me` | Mevcut kullanıcı |
| POST | `/api/v1/payments/stripe/webhook` | Stripe-Sig | `OrderHandler.StripeWebhook` | Ödeme tamamlandı event'i |

### Customer (JWT required)

| Method | Path | Handler | Amaç |
| --- | --- | --- | --- |
| POST | `/api/v1/orders` | `OrderHandler.Create` | Yeni sipariş (Stripe session) |
| GET | `/api/v1/orders` | `OrderHandler.List` | Müşterinin siparişleri |
| GET | `/api/v1/orders/:id` | `OrderHandler.Detail` | Sipariş detayı |
| GET | `/api/v1/tests` | `TestHandler.List` | Müşterinin testleri |
| GET | `/api/v1/tests/:id` | `TestHandler.Detail` | Test detay + assignments |
| GET | `/api/v1/tests/:id/activity` | `TestHandler.Activity` | Activity timeline (200) |
| GET | `/api/v1/tests/:id/reviews` | `TestHandler.Reviews` | Yazılan yorumlar |

### Admin Only (JWT + role=admin)

| Method | Path | Handler | Amaç |
| --- | --- | --- | --- |
| GET | `/api/v1/admin/metrics` | `AdminHandler.Metrics` | Dashboard metrics |
| GET | `/api/v1/admin/orders` | `AdminHandler.Orders` | Tüm siparişler |
| GET | `/api/v1/admin/tests` | `AdminHandler.Tests` | Tüm testler |
| GET | `/api/v1/admin/testers` | `AdminHandler.Testers` | 25 tester durumu |
| GET | `/api/v1/admin/payments` | `AdminHandler.Payments` | Tüm ödemeler |

### Internal (X-Activity-Token)

| Method | Path | Handler | Amaç |
| --- | --- | --- | --- |
| POST | `/api/v1/activity` | `ActivityHandler.Ingest` | Orchestrator'dan batch event'ler |

## Orchestrator (`apps/orchestrator`)

Base: `http://<MINIPC_IP>:9000`

### Public (X-API-Token)

| Method | Path | Handler | Amaç |
| --- | --- | --- | --- |
| GET | `/liveness` | inline | K8s-style liveness |
| GET | `/health` | inline | Emulator pool sağlık |
| GET | `/v1/emulators` | `EmulatorHandler.List` | 25 emulator durumu |
| POST | `/v1/emulators/start-all` | `EmulatorHandler.StartAll` | Tümünü başlat |
| POST | `/v1/emulators/:serial/restart` | `EmulatorHandler.Restart` | Tek emulator restart |
| POST | `/v1/emulators/:serial/wipe` | `EmulatorHandler.Wipe` | App data temizle |
| POST | `/v1/tasks/test_start/start` | `TaskHandler.StartTest` | opt-in + download |
| POST | `/v1/tasks/login_google/start` | `TaskHandler.StartLoginGoogle` | 5x retry |
| POST | `/v1/tasks/engage/start` | `TaskHandler.StartEngagement` | Daily engagement |
| POST | `/v1/tasks/review/start` | `TaskHandler.StartReview` | Gün 14 review |

## Web Routes (Next.js)

Static: `/`, `/legal/*`
Dynamic: `/login`, `/register`, `/dashboard/*`, `/admin/*`

## Asynq Job Tipleri

`apps/api/internal/worker/`:

| Type | Worker Handler | Orchestrator Endpoint |
| --- | --- | --- |
| `test_start` | `TestStartHandler` | `POST /v1/tasks/test_start/start` |
| `daily_engagement` | `DailyEngagementHandler` | `POST /v1/tasks/engage/start` |
| `write_review` | `WriteReviewHandler` | `POST /v1/tasks/review/start` |
| `login_google` | `LoginGoogleHandler` | `POST /v1/tasks/login_google/start` |

## İlgili

- [[services/api]]
- [[services/orchestrator]]
- [[services/worker]]
- [[code-graph/services]]
- [[code-graph/db-tables]]
