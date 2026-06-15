---
tags: [service, scheduler, go, asynq]
---

# Scheduler

> Asynq Scheduler + 14-day plan registrar. VPS'te çalışır.

## Sorumluluk

- Yeni test için 16 job kaydet (gün 0, 1-13, 14)
- Periyodik system job'ları (cron)
- Timezone-aware (Europe/Istanbul)

## Entry Point

`apps/api/cmd/scheduler/main.go`:
1. Config + Logger + Redis
2. Asynq Scheduler (Europe/Istanbul location)
3. Registrar init
4. Periyodik job'lar:
   - `0 4 * * *` — `system:cleanup_stale`
   - `0 5 * * *` — `system:check_warming`
5. Run + graceful shutdown

## Registrar

`apps/api/internal/scheduler/scheduler.go`:
- `Register14DayPlan(testID, packageName, startTime)` — order ödemesi tamamlandığında handler tarafından çağrılır
- Gün 0: `test_start`, ProcessIn(startTime - now)
- Gün 1-13: `daily_engagement`, ProcessIn(day * 24h)
- Gün 14: `write_review`, ProcessIn(14 * 24h)
- Her job stable `JobID(testID, type, day)` ile idempotent

## Asynq Scheduler API

```go
sched.Register(cronSpec string, task *asynq.Task, opts ...Option) (entryID, error)
```

Cron spec: standard 5-field (`min hour day month weekday`)

## Timezone

- `time.LoadLocation("Europe/Istanbul")`
- Tüm zamanlar bu TZ'ye göre
- Asynq Scheduler'ın `Location` field'ı set edilir

## Yeni Test Akışı

```
[Stripe Webhook] order.MarkPaid() başarılı
   │
   ▼
[Order Service] scheduler.Register14DayPlan(testID, pkg, time.Now())
   │
   ▼
[Asynq Scheduler] Redis'e 16 job ekler
   │
   ▼
[T+0] test_start fires → Worker → Orchestrator
[T+24h] engage fires
...
[T+14d] write_review fires
```

## İlgili

- [[services/api]] — Aynı monorepo
- [[services/worker]] — Job consumer
- [[task-runner]]
- [[code-graph/jobs]]
