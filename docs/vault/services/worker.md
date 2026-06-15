---
tags: [service, worker, go, asynq]
---

# Worker

> Asynq job consumer. VPS'te çalışır. Orchestrator'a HTTP ile bağlanır.

## Sorumluluk

- Redis queue'dan job consume et
- Type-safe payload decode
- Orchestrator'a HTTP POST
- Asynq retry/expire yönetimi
- Logging + recovery middleware

## Entry Point

`apps/api/cmd/worker/main.go`:
1. Config + Logger + Redis
2. RunnerClient (orchestrator HTTP)
3. Asynq ServerMux + middleware (Recovery, Logging, Timeout 15dk)
4. Handler registration: test_start, daily_engagement, write_review, login_google
5. Concurrency 10, queue priority

## Handler'lar

`apps/api/internal/worker/`:

| Dosya | Job Type | Orchestrator Endpoint |
| --- | --- | --- |
| `test_start.go` | `test_start` | `POST /v1/tasks/test_start/start` |
| `daily_engagement.go` | `daily_engagement` | `POST /v1/tasks/engage/start` |
| `write_review.go` | `write_review` | `POST /v1/tasks/review/start` |
| `login_google.go` | `login_google` | `POST /v1/tasks/login_google/start` |

## Client

`runner_client.go`:
- 30s timeout
- `X-API-Token` header
- Tüm POST'lar `application/json`

## Middleware

- `RecoveryMiddleware` — panic → error
- `LoggingMiddleware` — start/duration/error structured log
- `TimeoutMiddleware(15 * time.Minute)` — task context cancel

## Queue Config

```go
asynq.Config{
    Concurrency: 10,
    Queues: map[string]int{
        "critical": 6,
        "default":  3,
        "low":      1,
    },
    Logger: &lib.AsynqLogger{Logger: zapLogger},
}
```

## Enqueue (Client tarafı)

`client.go`:
- `EnqueueTestStart`, `EnqueueDailyEngagement`, `EnqueueWriteReview`, `EnqueueLoginGoogle`
- `EnqueueHealthcheck` — daily sistem kontrol
- Her biri stable `JobID` ile idempotent

## Retry Pattern

- Default: 2 retry, exp backoff
- login_google: 5 retry (warming CAPTCHA)
- Her retry `task_jobs.attempts++`

## İlgili

- [[services/api]] — Aynı monorepo
- [[services/scheduler]] — Job üretir
- [[services/orchestrator]] — HTTP target
- [[task-runner]] — Akış detayı
- [[code-graph/jobs]] — Job kataloğu
