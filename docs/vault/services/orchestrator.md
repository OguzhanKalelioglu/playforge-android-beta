---
tags: [service, orchestrator, go, appium, adb]
---

# Orchestrator

> Emulator farm controller + Appium UI otomasyonu. Mini PC'de çalışır.

## Sorumluluk

- 25 emulator'un lifecycle'ı (boot, health check, reset)
- ADB üzerinden uygulama kontrolü (AppActivate, InstallApp, WipeApp)
- Appium session yönetimi (W3C WebDriver)
- Task queue (test_start, engage, review, login_google)
- Anti-detect davranış (Gaussian delay, Bezier swipe, jitter)
- Activity event'lerini API'ye yollama

## Entry Point

`apps/orchestrator/cmd/orchestrator/main.go`:
1. Config load
2. Logger
3. Emulator pool init (25 entry)
4. Lifecycle manager (boot loop + health monitor)
5. Appium client init
6. TaskRunner init (queue + pool integration)
7. HTTP server (`:9000`)

## Modüller

### `internal/emulator/`
- `pool.go` — 25 emulator'u track eder. `AcquireForTestBlocking(testID, timeout)` ile CAS ready→busy
- `models.go` — `Emulator`, status enum (Booting, Ready, Busy, Error, ...)
- `container_manager.go` — Docker Compose control (start/stop/restart)

### `internal/adb/`
- `shell.go` — `adb shell` komutları
- App-specific: `AppActivate`, `InstallApp`, `WipeApp`, `GoogleAccountSignIn`, `Screenshot`

### `internal/appium/`
- W3C WebDriver transport. 10 dosya:
  - `client.go` — HTTP client
  - `session.go` — Active state, Quit
  - `locator.go` — 7 By stratejisi
  - `gestures.go` — Tap, Swipe, Scroll, LongPress
  - `wait.go` — Wait, WaitForElement, WaitForText
  - `screenshot.go` — PNG capture
  - `app.go` — AppActivate, ActivityStart
  - `capabilities.go` — AndroidEmulatorCaps, FreshCaps
  - `errors.go` — W3C structured errors
  - `base64.go` — response decoder

### `internal/task/`
- `task.go` — Task interface, Env, Step, Result
- `opt_in.go` — Test link → "Become a tester"
- `download.go` — Play Store search → install
- `engage.go` — 2-5dk random gesture
- `review.go` — 5 yıldız + yorum
- `login_google.go` — 8 adımlı Add Account wizard
- `comments.go` — 60+ TR yorum bank + selector

### `internal/taskrunner/`
- `runner.go` — In-process queue, CAS pool, appium session, defer quit
- `antidetect.go` — Gaussian delay, Bezier curve, jitter
- `activity.go` — ActivitySink (buffered chan 256, batch 50/2s, disk fallback)

### `internal/profile/`
- `manager.go` — Device fingerprint randomization (10 farklı cihaz profili)
- `time.go` — Test edilebilir saat

### `internal/lifecycle/`
- Boot orchestrator
- Health monitor (60s interval)
- Reset cycle (her test öncesi `pm clear --user 0`)

## Anti-Detect Detay

`internal/taskrunner/antidetect.go`:
- `GaussianDelay(μ, σ)` — Box-Muller normal distribution
- `BezierSwipe(x1, y1, x2, y2, steps)` — Doğal swipe path, 3 kontrol noktası
- `JitterXY(x, y, range)` — ±6px tap perturbasyonu
- `AppLaunchPause()` — 1.8-3.5s launch sonrası bekleme
- `EngagementDuration(min, maxSec)` — 2-5dk Gaussian weighted

## Konfigürasyon

```
APPIUM_URL=http://appium:4723
ACTIVITY_API_URL=http://host.docker.internal:8080/v1/activity
ACTIVITY_API_TOKEN=...
EMULATOR_MAX_INSTANCES=25
EMULATOR_CHECK_INTERVAL=60
ADB_HOST=127.0.0.1
ADB_SERVER_PORT=5037
COMPOSE_PATH=/app/infra/minipc/docker-compose.yml
SCREENSHOT_DIR=/data/screenshots
TASK_WATCHDOG_MIN=10
```

## Docker

`infra/minipc/deploy/docker-compose.yml`:
- `appium` (image: appium/appium:2.11.0, port 4723)
- `orchestrator` (custom Go build, port 9000)
- Mounts: `/var/run/docker.sock`, `infra/`, `data/`

## İlgili

- [[architecture]]
- [[task-runner]]
- [[services/api]] — Kontrat karşı taraf
- [[services/worker]] — Job gönderen
- [[services/scheduler]] — Plan kaynağı
- [[deployment]] — Mini PC deploy
