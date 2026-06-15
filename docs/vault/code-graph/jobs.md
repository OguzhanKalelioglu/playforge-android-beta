---
tags: [code-graph, jobs, asynq]
---

# Asynq Jobs

Background job kataloğu. Asynq Redis-backed queue.

## Queue Yapısı

| Queue | Öncelik | Tipik Job | Retry |
| --- | --- | --- | --- |
| `critical` | 6 | login_google | 5 |
| `default` | 3 | test_start, daily_engagement, write_review | 2-3 |
| `low` | 1 | system:cleanup_stale, system:check_warming | 1 |

Worker concurrency: 10 (Asynq server config).

## Job Tipleri

### `test_start` — Gün 0

- **File**: `apps/api/internal/worker/test_start.go`
- **TaskType**: `"test_start"`
- **Payload**: `model.TestStartPayload { test_id, assignment_id, package_name }`
- **Worker**: `TestStartHandler.ProcessTask` → `RunnerClient.StartTest`
- **Orchestrator**: `POST /v1/tasks/test_start/start`
- **Retry**: 3, exp backoff 30s/60s/120s
- **Timeout**: 15dk (worker middleware)

### `daily_engagement` — Gün 1-13

- **File**: `worker/daily_engagement.go`
- **Payload**: `DailyEngagementPayload { test_id, assignment_id, package_name, day }`
- **Worker**: `DailyEngagementHandler.ProcessTask` → `RunnerClient.StartEngagement`
- **Orchestrator**: `POST /v1/tasks/engage/start`
- **Retry**: 2
- **Gün 7, 10**: 5dk (heavy), diğer 2-3dk

### `write_review` — Gün 14

- **File**: `worker/write_review.go`
- **Payload**: `WriteReviewPayload { test_id, assignment_id, package_name, stars, comment, language }`
- **Worker**: `WriteReviewHandler.ProcessTask` → `RunnerClient.StartReview`
- **Orchestrator**: `POST /v1/tasks/review/start`
- **Retry**: 1

### `login_google` — Manual / 5x retry

- **File**: `worker/login_google.go`
- **Payload**: `LoginGooglePayload { test_id, assignment_id, email, password_encrypted, ... }`
- **Worker**: `LoginGoogleHandler.ProcessTask` → `RunnerClient.StartLoginGoogle`
- **Orchestrator**: `POST /v1/tasks/login_google/start`
- **Retry**: 5 (warming sonrası CAPTCHA çıkabilir)

## Stable Job IDs

`model.JobID(testID, taskType, day)` deterministik ID üretir. Aynı test + aynı gün için aynı ID, idempotent re-enqueue.

Örnek: `JobID("abc-123", TaskTypeEngage, 3)` → `abc-123:engage:3`

## Scheduler

`apps/api/internal/scheduler/scheduler.go` → `Registrar.Register14DayPlan(testID, packageName, startTime)`:

- Gün 0: `ProcessIn(startTime - now)`
- Gün 1-13: `ProcessIn(day * 24h)`
- Gün 14: `ProcessIn(14 * 24h)`

## Periyodik System Jobs

| Type | Cron (Europe/Istanbul) | Amaç |
| --- | --- | --- |
| `system:cleanup_stale` | `0 4 * * *` | Pending order > 30dk, expire et |
| `system:check_warming` | `0 5 * * *` | 3 gün geçmiş warming → active |

## Asynqmon (Queue UI)

- URL: `https://queue.testerscomm.net`
- Basic auth + IP whitelist
- Realtime: pending, active, retry, dead queue
- Retry, dead letter cleanup, job detail

## Dead Letter Handling

- `retrying` queue 3 retry sonrası `dead`'a düşer
- Admin panel'de `failed_tasks_24h` metric ile görünür
- TaskJobs tablosunda `status='dead'` olarak audit

## İlgili

- [[task-runner]] — Pipeline detayı
- [[services/worker]] — Worker binary
- [[services/scheduler]] — Scheduler binary
- `apps/api/internal/worker/` — Handler'lar
- `apps/api/internal/scheduler/` — Plan registrar
