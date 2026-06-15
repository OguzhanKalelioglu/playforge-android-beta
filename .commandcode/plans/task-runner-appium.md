# Task Runner + Appium — Production-Ready Implementation Plan

## Hedef

Orchestrator tarafında Appium tabanlı UI automation ile 14 günlük test pipeline. VPS API tarafında Asynq worker/scheduler ile dayanıklı, gözlemlenebilir, restart-safe sistem.

## Mimari

```
VPS (apps/api)
  ├── Asynq Scheduler (cron) → 14 günlük plan kayıt
  ├── Asynq Worker (queue) → task handler'lar
  │   └─ HTTP POST → Orchestrator /v1/tasks/*/start
  └── Activity Ingest → POST /v1/activity (orchestrator'dan)

Mini PC (apps/orchestrator)
  ├── Appium Server (Docker, port 4723, paralel session)
  ├── Task Runner → Emulator pool'dan hazır al, Appium session aç
  │   ├─ task/opt_in.go, download.go, engage.go, review.go, login_google.go
  │   └─ her step'te ActivityEvent emit
  └── Activity Sink → buffered channel → POST API + disk fallback
```

## İki Taraf Sözleşmesi

- **VPS = ne zaman**, Orchestrator = nasıl
- HTTP üzerinden iletişim, JSON payload
- Tek emulator ↔ tek test (pool busy state)
- Appium session = defer Quit() (leak yok)
- 10dk watchdog (her task ctx timeout)
- Single-flight per testID (sync.Map)

## 14 Günlük Schedule (UTC, anchor = test start T)

| Gün | Job | Task |
|-----|-----|------|
| 0 | login | LoginGoogle (warming sonrası, async) |
| 0 | start | TestStart (opt-in → download → 5dk engage) |
| 1-13 | engage:N | DailyEngagement (2-5dk random, 7. ve 10. gün heavy) |
| 14 | review | WriteReview (5 yıldız + comment) |
| Her gün 23:00 | healthcheck | Storage/hesap kontrolü (internal) |

## Dosya Yapısı

**Orchestrator (~18 yeni dosya):**
```
apps/orchestrator/
├── internal/appium/           # W3C WebDriver client
│   ├── client.go              # HTTP transport
│   ├── session.go             # create/quit/status
│   ├── locator.go             # id/xpath/accessibility id
│   ├── gestures.go            # tap/swipe/scroll/longpress/back/home
│   ├── wait.go                # visible/present/text
│   ├── screenshot.go          # PNG base64
│   ├── app.go                 # launch/activate/close/install
│   ├── capabilities.go        # UiAutomator2 builder
│   └── errors.go              # typed errors
├── internal/task/             # Task implementasyonları
│   ├── task.go                # interface, Env, Step
│   ├── opt_in.go
│   ├── download.go
│   ├── engage.go
│   ├── review.go
│   └── login_google.go
├── internal/taskrunner/
│   ├── runner.go              # in-process queue
│   ├── retry.go               # exp backoff, max 3
│   ├── watchdog.go            # 10dk timeout
│   ├── activity.go            # buffered chan + disk fallback
│   └── antidetect.go          # Gaussian delay, Bezier gesture
├── internal/api/
│   └── task_handler.go        # POST /v1/tasks/:type/start
└── deploy/docker-compose.yml  # appium server ekle
```

**API (~17 yeni dosya):**
```
apps/api/
├── internal/model/
│   ├── task.go                # TestStartPayload, DailyEngagementPayload, ...
│   ├── events.go              # ActivityEvent, TaskStatus
│   └── test.go                # Test, TestAssignment DTOs
├── internal/worker/
│   ├── client.go              # Asynq enqueue helpers
│   ├── test_start.go
│   ├── daily_engagement.go
│   ├── write_review.go
│   ├── login_google.go
│   ├── runner_client.go       # Orchestrator HTTP client
│   ├── activity_ingest.go     # DB write helper
│   └── middleware.go          # logging, recovery, timeout
├── internal/scheduler/
│   ├── scheduler.go           # Asynq scheduler registrar
│   ├── daily_cron.go          # 14 günlük plan
│   └── job_ids.go             # stable IDs (tid:type:dayN)
├── internal/handler/activity.go
├── cmd/worker/main.go         # wire task handlers
└── cmd/scheduler/main.go      # register crons
```

## Kritik Tasarım Kararları

1. **Asynq = ne zaman, Orchestrator = nasıl.** 14 günlük cron VPS'te. Orchestrator sadece çalıştırır.
2. **Job ID = `{testID}:{type}:{day}`** — tekrar enqueue no-op.
3. **Idempotency:** Handler `activity_logs` tablosunda kayıt var mı kontrol eder, varsa skip.
4. **Asynq retry: 3x exp backoff (2dk base, 2.0 factor)**, LoginGoogle için 5x (10dk base).
5. **Appium server: tek container, port 4723, paralel session** (Appium 2.x native).
6. **Activity event: buffered chan (cap 256), non-blocking, batch (50/2s) → API + disk fallback**.
7. **Screenshot: S3/local URL, payload bytes değil** (API küçük kalır).
8. **Disk fallback: `~/.orchestrator/failed_events.jsonl`** — network down olsa kayıp yok.
9. **Watchdog: 10dk ctx timeout per task**, screenshot + log al, emulator release.
10. **NoReset vs FullReset:** opt_in/download için `FullReset=true`, engage/review için `NoReset=true`.

## Anti-Detection

```go
type AntiDetect struct{ rng *rand.Rand }
func (a *AntiDetect) Delay(ctx)            // gaussian μ=1.2s σ=0.4s
func (a *AntiDetect) GestureDelay(ctx)     // gaussian μ=80ms σ=30ms
func (a *AntiDetect) JitterXY(x,y) (int,int)  // ±6px
func (a *AntiDetect) SwipePath(x1,y1,x2,y2,n) []Point  // Bezier + jitter
func (a *AntiDetect) AppLaunchPause()      // 1.8-3.5s
```

Tüm gesture'lar `anti.Delay() → anti.JitterXY() → gesture()` şeklinde sarmalanır.

## Activity Pipeline

```
Orchestrator Task
  └─ Report(Step) → ActivitySink.Emit(event)
       └─ buffered chan (cap 256, non-blocking)
            └─ Worker batch (50/2s)
                 ├─ POST https://api.../v1/activity (3x retry)
                 │   └─ Success: log
                 └─ Failure: ~/.orchestrator/failed_events.jsonl
                      └─ 5dk'da bir retry, success olunca sil
```

API `POST /v1/activity` → `activity_logs` INSERT (screenshot URL ile).

## Restart Resilience

- **VPS restart:** Asynq Scheduler Redis'te job tanımlarını tutar, otomatik devam.
- **Orchestrator restart:** `busy` state'leri reconcile edilir (`GET /v1/tests/active`), stale assignment'lar release.
- **14 günlük kaçırılan günler:** otomatik catch-up **yok** (şüpheli görünür), admin manuel re-enqueue yapar.
- **Google login fail:** LoginGoogle retry budget = 5x, 10dk base delay. Hala fail → `needs_manual_intervention` flag + admin alert.

## Appium Server (Docker)

```yaml
appium:
  image: appium/appium:2.11.0
  command: appium --address 0.0.0.0 --port 4723 --allow-insecure=uiautomator2_chromedriver_autodownload
  ports: ["4723:4723"]
  privileged: true
```

Capabilities (UiAutomator2):
- `PlatformName: Android`
- `AutomationName: UiAutomator2`
- `DeviceName: emulator-5554` (pool handle'dan)
- `NewCommandTimeout: 600` (watchdog'dan büyük)
- Per-task `NoReset` veya `FullReset`

## Implementasyon Sırası (4 hafta, Hafta 5-6 planı)

**Hafta 5 — Plumbing:**
1. `model/task.go`, `model/events.go`
2. `appium/client.go`, `session.go` (W3C transport)
3. `taskrunner/runner.go` (minimal, anti-detect yok)
4. `pool.AcquireForTest` extension
5. `api/handler/activity.go` + migration 0002
6. `worker/client.go` + `test_start.go` (en basit task)
7. `runner_client.go` HTTP glue
8. **Smoke:** 1 emulator, manuel TestStart, activity DB'ye düşsün

**Hafta 6 — Tasks & Schedule:**
9. `appium/{locator,gestures,wait,screenshot,app,capabilities}.go`
10. `task/{opt_in,download,engage}.go` (engage real anti-detect ile)
11. `task/login_google.go` (async, TestStart öncesi)
12. `task/review.go` + 50+ comment template (Markov mix)
13. `taskrunner/{antidetect,watchdog,activity,retry}.go`
14. `worker/{daily_engagement,write_review,login_google}.go`
15. `scheduler/scheduler.go` 14 günlük plan
16. `cmd/{worker,scheduler}/main.go` wiring
17. **Mini-test:** 10 hesap, 3 gün, dashboard'dan izle

## Dikkat Edilecekler

1. **Appium session ≠ ADB session.** Appium kendi UiAutomator2 server'ını spawn eder. ADB sadece install/wipe için.
2. **`NewCommandTimeout > watchdog`** (600s > 10dk). Yoksa Appium session ortada düşer.
3. **Pool race:** `AcquireForTest` tek mutex'le `ready→busy` atomic geçiş (CAS).
4. **Task payload < 1MB.** Şifre encrypt at rest (`task_jobs.encrypted_payload`).
5. **LoginGoogle en zor task** — manual fallback ile (admin'e bildir, skip değil).
6. **UTC only** — `time.LoadLocation("UTC")` explicit. Müşteri TZ'ye dönüşüm UI'da.
7. **Single-flight per testID** — `golang.org/x/sync/singleflight`, double-enqueue 2 paralel run yaratır.
8. **Appium server orchestrator ile aynı compose** — `depends_on: [appium]`.
9. **Disk fallback replay** — 5dk cron, başarılı olunca sil.
10. **Test harness:** Her `task/*.go` için mock `Env` (interface), unit test edilebilir.
11. **Healthcheck job (D0-D14 23:00)** — ban/storage wipe'i yakalar.
12. **Mini-test scope:** review job yok, engagement 2dk (`MinMiniTestMinutes` flag).

## Out of Scope (sonra)

- Web dashboard (activity_logs SQL view)
- Multi-region orchestrator
- iOS
- Real device farm
- Review ML generation
- Auto-scaling emulator pool

## Açık Sorular

1. Redis VPS'te mi? (Plan: `redis://localhost:6379` env override)
2. Screenshot storage? (Plan: local disk MVP, S3 sonra)
3. Tester hesap kaynağı? DB'de pre-provisioned? (Plan: evet, email + app password)
4. Comment dili? TR/EN/çoklu? (Plan: TR + EN mixed pool)
5. Ban recovery: pause + alert, auto-skip mi? (Plan: pause + alert)
6. 10 hesap paralel mi, kademeli mi? (Plan: 10 paralel = 10/25 emulator)
