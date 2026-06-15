# Emulator Setup Otonom Altyapısı

## Genel Bakış

25 Android emulator'ü tamamen otonom yöneten sistem. Persistent çalışır (sürekli açık), her test öncesi fabrika ayarlarına sıfırlanır.

```
Orchestrator (Go binary, port 9000)
   ├── Container Manager: docker compose kontrol (up/down/restart)
   ├── Health Monitor: sürekli boot durumu kontrol
   ├── ADB Client: shell komutları (install, wipe, screenshot)
   ├── Emulator Pool: 25 emulator'ün durumu (in-memory)
   └── Lifecycle Manager: tüm akışı orkestre eder
        │
        ▼
Docker Compose (25 emulator)
   ├── emulator-01 ... emulator-25
   ├── Her biri: budtmo/docker-android, KVM, headless
   └── 1.5GB RAM × 25 = 37.5GB
```

## Mimari Kararlar

| Karar | Tercih | Neden |
|-------|--------|-------|
| Persistent | ✅ Sürekli açık | Test anında hazır, boot bekleme yok |
| Reset on test | ✅ Her test öncesi wipe | Temiz başlangıç, hesap kalıntısı yok |
| Tam otonom | ✅ Orchestrator yönetir | Tek API komutuyla tüm farm |
| KVM | ✅ Hardware accel | Zorunlu (KVM yoksa emulator çalışmaz) |

## Container Durumları (State Machine)

```
   ┌──────────┐
   │ unknown  │  (henüz kontrol edilmedi)
   └────┬─────┘
        │ container check
        ▼
   ┌──────────┐
   │ stopped  │  (container yok, veya durmuş)
   └────┬─────┘
        │ docker compose up
        ▼
   ┌──────────┐
   │ booting  │  (container ayakta, sys.boot_completed=0)
   └────┬─────┘
        │ getprop sys.boot_completed=1
        ▼
   ┌──────────┐
   │  ready   │  (test için uygun)
   └────┬─────┘
        │ test başladı
        ▼
   ┌──────────┐
   │  busy    │  (test çalışıyor)
   └────┬─────┘
        │ test bitti
        ▼
   ┌──────────┐
   │  ready   │  (yeni test için uygun)
   └──────────┘

Herhangi bir durumdan:
   ┌──────────┐
   │  error   │  (boot timeout, ADB kopması, vb.)
   └────┬─────┘
        │ otomatik recovery (max 3 deneme)
        ▼
   ┌──────────┐
   │ stopped  │
   └──────────┘
```

Wipe operasyonu sırasında: `ready` → `wiping` → `ready` (veya `booting` → `ready`)

## Orchestrator Bileşenleri

### 1. `internal/emulator/pool.go` — Pool (in-memory state)

**Sorumluluk:** 25 emulator'ün anlık durumunu tutar.

```go
type Emulator struct {
    Index       int
    Serial      string       // "emulator-5554", "emulator-5556"...
    ContainerID string       // Docker container ID
    Status      Status       // stopped, booting, ready, busy, wiping, error
    TesterID    string       // Hangi tester atanmış (busy iken)
    LastUsed    time.Time
    BootedAt    time.Time
    BootCount   int          // Kaç kez boot oldu
    ErrorMsg    string
}
```

**Method'lar:**
- `Acquire()` → Ready durumda bir emulator alır, busy yapar
- `Release(serial)` → Busy → Ready
- `SetStatus(serial, status, errMsg)` → Durum güncelle
- `Counts()` → Her status'ten kaç tane var (metric için)

### 2. `internal/container/manager.go` — Docker Compose Manager

**Sorumluluk:** `docker compose` komutlarını çalıştırır.

**Method'lar:**
- `UpAll(ctx)` → Tüm 25 emulator'ü başlat (ilk kurulum)
- `Up(ctx, service)` → Tek bir emulator başlat (örn: "emulator-05")
- `Stop(ctx, service)` → Tek bir emulator durdur
- `Restart(ctx, service)` → Tek bir emulator yeniden başlat
- `Down(ctx)` → Tümünü durdur (orchestrator kapanırken)
- `PS(ctx)` → Container listesi (durum senkronizasyonu için)
- `ContainerID(ctx, service)` → Tek bir container'ın ID'si

**Service naming:** `emulator-01`, `emulator-02`, ... `emulator-25` (zero-padded)

### 3. `internal/health/monitor.go` — ADB Health Monitor

**Sorumluluk:** Emulator'ün gerçekten hazır olup olmadığını kontrol eder.

**Method'lar:**
- `Check(ctx, serial)` → Anlık sağlık kontrolü
  - `adb connect` (gerekirse)
  - `getprop sys.boot_completed` → 1 mi?
  - `getprop ro.build.version.release` → Android version
  - `getprop ro.product.model` → Pixel 5
  - `uptime` → Çalışma süresi
- `WaitForBoot(ctx, serial)` → Boot olana kadar poll et (max 5 dakika)

### 4. `internal/lifecycle/manager.go` — Orchestrator (Ana Katman)

**Sorumluluk:** Tüm bileşenleri koordine eder.

**Method'lar:**
- `Start(ctx)` → Auto-start (25 emulator), health loop başlat
- `StartAllEmulators(ctx)` → İlk kurulumda veya down sonrası
- `StartEmulator(ctx, serial)` → Tek bir emulator başlat
- `StopEmulator(ctx, serial)` → Tek bir emulator durdur
- `RestartEmulator(ctx, serial)` → Yeniden başlat (boot fix için)
- `WipeEmulator(ctx, serial)` → Fabrika ayarlarına sıfırla (test öncesi)
- `ResetForTest(ctx, serial)` → Test öncesi wipe
- `StopAllEmulators(ctx)` → Tümünü durdur (graceful shutdown)

**Health Loop (60 saniyede bir):**
- Her emulator'ün boot durumunu kontrol et
- Hata varsa status'u "error" yap
- Booting → Ready transition'ı yakala
- Ready → Booting (reboot detected) uyarısı

## HTTP API Endpoints

**Public (no auth):**
```
GET  /health                  # Toplam sağlık (counts, status)
GET  /liveness                # Liveness check
GET  /emulators               # Tüm 25 emulator (detaylı)
GET  /emulators/status        # Sadece status dict
GET  /emulators/counts        # Status başına sayı
GET  /emulators/:serial       # Tek bir emulator (örn: emulator-5554)
```

**Auth required (X-API-Token):**
```
POST /emulators/start-all                  # Hepsini başlat (async)
POST /emulators/stop-all                   # Hepsini durdur (async)
POST /emulators/:serial/start              # Tek başlat
POST /emulators/:serial/stop               # Tek durdur
POST /emulators/:serial/restart            # Tek restart
POST /emulators/:serial/wipe               # Fabrika ayarı (factory reset)
POST /emulators/:serial/reset              # Test öncesi reset (wipe alias)
```

## Akış Diyagramları

### 1. Orchestrator İlk Başlatma

```
main()
  │
  ├─ config.Load() (env: EMULATOR_MAX_INSTANCES, AUTO_START_EMULATORS, ...)
  ├─ logger init
  ├─ adb.StartADBServer()
  ├─ emulator.Pool(25) oluştur
  ├─ container.Manager oluştur
  ├─ health.Monitor oluştur
  ├─ lifecycle.Manager oluştur
  │
  ├─ go manager.Start(ctx)  // async
  │   │
  │   ├─ if AutoStart: StartAllEmulators()
  │   │   │
  │   │   ├─ docker compose up -d (25 emulator)
  │   │   ├─ refreshAllStatuses() (container ID'leri al)
  │   │   └─ waitForAllReady() (5-10 dakika)
  │   │
  │   └─ go runHealthLoop() // 60s interval
  │
  └─ httpSrv.ListenAndServe() :9000
       │
       └─ /health, /emulators, /emulators/:serial, ...
```

### 2. Tek Emulator'ü Başlatma (API)

```
POST /emulators/emulator-5554/start  (X-API-Token)
  │
  ▼
Server.StartEmulator()
  │
  ├─ pool.Get("emulator-5554")
  ├─ service = "emulator-01" (from index)
  ├─ pool.SetStatus("emulator-5554", booting)
  ├─ container.Up(ctx, "emulator-01")
  │   └─ docker compose up -d emulator-01
  ├─ pool.SetContainerID() (container ID al)
  │
  └─ go healthMon.WaitForBoot()  // background
       │
       ├─ Poll: getprop sys.boot_completed
       │   └─ "1" → Boot tamam
       │
       └─ pool.SetStatus("emulator-5554", ready)
            └─ Log: "emulator ready"
```

### 3. Test Öncesi Reset (Worker → Orchestrator)

```
POST /emulators/emulator-5554/reset  (worker tarafından)
  │
  ▼
Server.ResetForTest()
  │
  ├─ pool.SetStatus(wiping)
  │
  ├─ adb.WipeData("emulator-5554")
  │   └─ adb shell pm clear --user 0
  │       (tüm app data, hesap cache, ayarlar silinir)
  │
  ├─ sleep 2s (Android reboot)
  │
  └─ Durum kontrolü:
      ├─ Bootluysa: pool.SetStatus(ready)
      └─ Bootlu değilse:
          └─ go healthMon.WaitForBoot()
              └─ pool.SetStatus(ready)
```

### 4. Otomatik Health Check (Her 60 saniye)

```
HealthLoop tick
  │
  ├─ Her emulator için:
  │   │
  │   ├─ if Status == stopped: skip
  │   │
  │   ├─ healthMon.Check(serial) (15s timeout)
  │   │   │
  │   │   ├─ Success + boot_completed=1:
  │   │   │   └─ if Status == booting → ready (transition)
  │   │   │   └─ if Status == ready + !boot_completed → booting (reboot)
  │   │   │
  │   │   └─ Error:
  │   │       └─ if Status != error → SetStatus(error, msg) + Log.Warn
  │   │
  │   └─ Update last_check
  │
  └─ 60 saniye bekle
```

## Configuration

`.env` (orchestrator tarafında):
```bash
PORT=9000
APP_ENV=production
LOG_LEVEL=info

EMULATOR_MAX_INSTANCES=25
AUTO_START_EMULATORS=true
EMULATOR_CHECK_INTERVAL=60  # saniye

ADB_HOST=127.0.0.1
ADB_SERVER_PORT=5037

COMPOSE_PROJECT=testers-minipc
COMPOSE_PATH=../../infra/minipc/docker-compose.yml
SERVICE_PREFIX=emulator

ORCHESTRATOR_API_TOKEN=<32+ random karakter>
```

## Kurulum

### Mini PC'de İlk Kurulum

```bash
# 1. Repo klonla
git clone <repo> ~/app
cd ~/app

# 2. .env oluştur
cp .env.example .env
nano .env  # ORCHESTRATOR_API_TOKEN ayarla

# 3. Docker Compose'u başlat (otomatik olarak 25 emulator ayağa kalkar)
cd infra/minipc
docker compose up -d
# 5-10 dakika bekle (hepsi boot olsun)

# 4. Orchestrator'ı başlat
cd ~/app/apps/orchestrator
go run ./cmd/orchestrator
# veya production'da:
# ./bin/orchestrator
```

### VPS'ten Test Etme

```bash
# VPS'ten Mini PC'ye API çağrısı
curl -X POST http://<MINIPC_IP>:9000/health

# Auth gerekli endpoint:
curl -X POST http://<MINIPC_IP>:9000/emulators/emulator-5554/start \
  -H "X-API-Token: <token>"

# Tüm emülatörleri başlat:
curl -X POST http://<MINIPC_IP>:9000/emulators/start-all \
  -H "X-API-Token: <token>"

# Durum:
curl http://<MINIPC_IP>:9000/emulators/counts
# {"total":25,"counts":{"ready":25,"stopped":0,"error":0,...}}
```

## Operasyon

### Günlük Kontrol

```bash
# Sağlık
curl http://<MINIPC_IP>:9000/health | jq .

# Counts
curl http://<minipc>:9000/emulators/counts | jq .

# Hata varsa detay
curl http://<minipc>:9000/emulators | jq '.emulators[] | select(.status=="error")'
```

### Emulator Crash Olduğunda

Otomatik recovery var ama manuel müdahale:

```bash
# 1. Tek emulator'ü restart et
curl -X POST http://<minipc>:9000/emulators/emulator-5554/restart \
  -H "X-API-Token: $TOKEN"

# 2. Hala hata varsa wipe + restart
curl -X POST http://<minipc>:9000/emulators/emulator-5554/wipe \
  -H "X-API-Token: $TOKEN"

# 3. Tamamen çöktüyse container'ı sil ve yeniden başlat
docker compose -p testers-minipc -f infra/minipc/docker-compose.yml \
  up -d --force-recreate emulator-15
```

### RAM/CPU Monitoring

```bash
# Container resource kullanımı
docker stats --no-stream | grep tc-emulator

# Orchestrator log
docker logs tc-orchestrator --tail 100 -f
```

## Sınırlamalar ve Dikkat Edilecekler

1. **KVM zorunlu:** `/dev/kvm` yoksa hiçbir emulator çalışmaz. BIOS'tan VT-x/AMD-V açılmalı.

2. **Boot süresi:** İlk açılışta her emulator 3-5 dakika. Persistent modda sonraki restart'lar 1-2 dakika.

3. **ADB port çakışması:** 25 emulator = 50 port (her biri 2 port: console + adb). Firewall'da 5554-5603 açık olmalı.

4. **Wipe gecikmesi:** `pm clear` 5-10 saniye sürer, ardından Android reboot eder (30-60 saniye). Test başlatmadan önce ready state'i bekle.

5. **Disk I/O:** 25 emulator aynı anda wipe = yüksek I/O. SSD olmazsa sorunlu.

6. **Snapshot önerilmez:** Şu an her wipe = fresh state. Persistent state istiyorsan snapshot ekle (ileride).

## Sonraki Adımlar (Hafta 5+)

- [ ] `internal/task/` — Asynq worker'ların çağıracağı görevler
  - `opt_in.go` — Play Store opt-in (UI automation)
  - `download.go` — İndirme + kurulum
  - `engage.go` — Günlük engagement
  - `review.go` — Review yazma
- [ ] Appium client (`internal/appium/`) — UI automation için
- [ ] Real device integration (1 fiziksel telefon)
- [ ] Prometheus metrics export
