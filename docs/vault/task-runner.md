---
tags: [task-runner, appium, automation]
---

# Task Runner Pipeline

25 emülatör üzerinde uçtan uca otomasyon. Asynq (queue) + Appium (UI) + orchestrator (emulator farm).

## 14 Günlük Plan

`internal/scheduler/scheduler.go` her test için 16 job kaydeder:

| Gün | Job | Task | Tipik Süre |
| --- | --- | --- | --- |
| 0 | `test_start` | Login + opt-in + download | 5-15dk (5x retry) |
| 1-13 | `daily_engagement` | 2-5dk rastgele swipe/tap/back | 2-5dk (gün 7, 10: 5dk) |
| 14 | `write_review` | 5 yıldız + yorum + post | 1-2dk |

## Job Tipleri

```go
const (
    TaskTypeTestStart        = "test_start"
    TaskTypeDailyEngagement  = "daily_engagement"
    TaskTypeWriteReview      = "write_review"
    TaskTypeLoginGoogle      = "login_google"  // TestStart'tan önce, 5x retry
)
```

Her job'ın `TaskID` = `JobID(testID, type, day)` — stable, idempotent re-enqueue.

## Akış (gün 3 engagement örneği)

```
1. Asynq scheduler fires "test_7:engage:3"
   │
2. Worker pulls job → RunnerClient.StartEngagement()
   │
3. POST {RUNNER_URL}/v1/tasks/engage/start
   { test_id: "uuid", day: 3, package_name: "..." }
   │
4. Orchestrator TaskHandler.Submit()
   ├─ pool.AcquireForTestBlocking(testID, 5s)  // CAS ready→busy
   ├─ appium.NewSession(AndroidEmulatorCaps)
   ├─ defer: session.Quit() + pool.Release()
   ├─ defer: watchdog (10dk timeout)
   │
5. Task.Engage (gün 3, 2-5dk)
   ├─ Step: open app (AppActivate, AppLaunchPause 1.8-3.5s)
   ├─ Step: 20-30 random gesture
   │   ├─ swipe (Bezier curve, ±6px jitter)
   │   ├─ tap (Bezier curve)
   │   ├─ back (KEYCODE_BACK)
   │   └─ pause (Gaussian μ=1.2s σ=0.4s)
   ├─ report each step → ActivitySink
   │   └─ batch 50/2s → POST /v1/activity (X-Activity-Token)
   └─ final event: engagement_completed
   │
6. Return result → Worker → Asynq success
```

## Anti-Detect

`internal/taskrunner/antidetect.go`:
- `GaussianDelay(μ, σ)` — Box-Muller normal distribution
- `BezierSwipe(x1, y1, x2, y2)` — Doğal swipe path
- `JitterXY(x, y, ±N)` — Tap koordinat perturbasyonu
- `AppLaunchPause()` — 1.8-3.5s launch sonrası bekleme
- `EngagementDuration(min, maxSec)` — 2-5dk rastgele

## 25 Hesap Rotasyonu

`tester_daily_usage` tablosu günlük görev sayısını tutar. `PickAccount`:
- `status='active'` AND
- `tasks_completed_today < 3` (over-use engelle)
- 7 gün rolling: en az kullanılan önce

## Retry & Dead Letter

- `test_start`: 3 retry, exp backoff 30s/60s/120s
- `engage`: 2 retry
- `review`: 1 retry
- Tüm retry'lar tükenirse task `dead` queue'ya düşer, admin panel'de görünür

## İlgili

- [[services/orchestrator]] — `internal/taskrunner/`, `internal/task/`
- [[services/worker]] — Asynq consumer
- [[services/scheduler]] — Plan kaydı
- `apps/orchestrator/internal/appium/` — Appium client
- `apps/orchestrator/internal/task/comments.go` — 60+ TR yorum bank
- `apps/orchestrator/internal/profile/manager.go` — Device fingerprint
