# TestersCommunity

Ücretli Android uygulama test hizmeti platformu. Google Play Store kapalı beta test programı için 25 farklı Google hesabı kullanılarak 14 günlük otomatik test süreci.

> ⚠️ **Yasal Uyarı:** Bu hizmet Google Play Console Hizmet Şartları'nın sınırında hareket eder. Müşteriler bu riski kabul eder. Detaylar için `/legal/terms` ve `/legal/refund` sayfalarına bakın.

---

## Mimari

```
Müşteri (Browser)
   ↓ HTTPS
Hetzner VPS (€25/ay) - PostgreSQL + Redis + Go API + Next.js + Caddy
   ↓ PostgreSQL (TLS)
Mini PC (64GB RAM) - 25x Android Emulator + Orchestrator (Go)
   ↓ Ev IP'si
Google Play Store (25 Hesap)
```

Detaylı mimari için: [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md)

---

## Proje Yapısı

```
.
├── apps/
│   ├── api/                  # Go backend (VPS'te çalışır)
│   │   ├── cmd/server/       # HTTP API
│   │   ├── cmd/worker/       # Asynq worker (background jobs)
│   │   ├── cmd/scheduler/    # Asynq scheduler (cron)
│   │   ├── internal/         # handler, service, repository
│   │   └── migrations/       # SQL migrations
│   │
│   ├── orchestrator/         # Go binary (Mini PC'de çalışır)
│   │   ├── cmd/orchestrator/
│   │   └── internal/         # adb, emulator, appium
│   │
│   └── web/                  # Next.js 14 (frontend)
│       ├── app/              # App Router (landing, auth, dashboard, admin, legal)
│       ├── components/ui/    # shadcn-style primitives
│       └── lib/              # utils, api-client
│
├── infra/
│   ├── vps/                  # VPS deployment (docker-compose, Caddy, scripts)
│   └── minipc/               # Mini PC deployment (emulator farm)
│
├── docs/                     # Mimari, runbook, legal
├── go.work                   # Go workspaces
├── package.json              # pnpm workspace
└── .env.example              # Environment template
```

---

## Teknoloji Stack

**Backend:** Go 1.22, Gin, pgx (PostgreSQL), go-redis, Asynq, zap, golang-jwt, iyzico SDK

**Frontend:** Next.js 14 (App Router), TypeScript, Tailwind CSS, shadcn/ui, React Hook Form + Zod, TanStack Query, Zustand

**Veritabanı:** PostgreSQL 15 (pgcrypto, uuid-ossp)

**Queue/Cache:** Redis 7 + Asynq (background jobs + scheduler)

**Infra:** Docker Compose, Caddy 2 (reverse proxy + otomatik HTTPS), Hetzner VPS, budtmo/docker-android

**Ödeme:** Iyzico (Türkiye pazarı için optimize)

---

## Hızlı Başlangıç (Development)

### 1. Repo'yu klonla

```bash
git clone <repo-url> testerscommunity
cd testerscommunity
cp .env.example .env
# .env dosyasını düzenle (özellikle DB_PASSWORD, JWT_SECRET)
```

### 2. Go workspace setup

```bash
# apps/api ve apps/orchestrator için go mod tidy
cd apps/api && go mod tidy && cd ../..
cd apps/orchestrator && go mod tidy && cd ../..
```

### 3. PostgreSQL + Redis'i başlat (yalnızca development)

```bash
cd infra/vps
docker compose up -d postgres redis
```

### 4. Migration'ları çalıştır

```bash
# Container çalışınca otomatik çalışır (docker-entrypoint-initdb.d)
# Manuel çalıştırmak istersen:
docker exec -i testers-vps-postgres-1 psql -U tester -d testers < apps/api/migrations/0001_init.sql
```

### 5. Go API'yi çalıştır

```bash
# Terminal 1: API
cd apps/api
go run ./cmd/server
# → http://localhost:8080

# Terminal 2: Worker
go run ./cmd/worker

# Terminal 3: Scheduler
go run ./cmd/scheduler
```

### 6. Next.js'i çalıştır

```bash
cd apps/web
pnpm install
pnpm dev
# → http://localhost:3000
```

---

## Production Deployment

### VPS (Hetzner CPX41)

```bash
# İlk kurulum
ssh root@<VPS_IP>
bash infra/vps/scripts/setup.sh

# Repo'yu klonla
ssh testops@<VPS_IP>
git clone <repo-url> ~/app
cd ~/app
cp .env.example .env
nano .env  # DB_PASSWORD, JWT_SECRET, IYZICO_API_KEY, vb. doldur

# Servisleri başlat
cd infra/vps
docker compose up -d
docker compose ps
docker compose logs -f api
```

**Doğrulama:**
- `curl https://testerscomm.net` → landing page
- `curl https://api.testerscomm.net/health` → 200 OK
- `https://asynq.testerscomm.net` → queue dashboard (basic auth)

### Mini PC (64GB RAM, Ubuntu 22.04)

```bash
# İlk kurulum
sudo bash infra/minipc/scripts/setup.sh

# Repo'yu klonla
git clone <repo-url> ~/app
cd ~/app

# 1 emulator başlat (Phase 1)
cd infra/minipc
docker compose up -d

# Test et
adb connect localhost:5554
adb devices
adb -s emulator-5554 shell getprop ro.build.version.release  # 11
```

---

## Geliştirme Yol Haritası

Detaylı plan için: `.commandcode/plans/testers-community-clone.md`

**Hafta 1-2 (Şu an):** Altyapı kurulumu ✅
- [x] Go monorepo
- [x] PostgreSQL şeması
- [x] VPS Docker Compose
- [x] Caddy reverse proxy
- [x] Next.js landing page
- [x] Health check endpoint'leri
- [x] Setup/backup scriptleri

**Hafta 3-4:** Hesap yönetimi + emulator setup
**Hafta 5-6:** Test pipeline (Asynq, opt-in, engagement, review)
**Hafta 7-8:** Web + ödeme (Iyzico)
**Hafta 9-10:** Ölçeklendirme (25 hesap + 25 emulator)
**Hafta 11-12:** Lansman + beta

---

## Maliyet

| Kalem | Aylık |
|-------|-------|
| Hetzner CPX41 (16GB) | €25 |
| Domain | ~$1 |
| Backblaze B2 (50GB) | $0.25 |
| SMS servisi (5 hesap) | $5 |
| **Toplam sabit** | **~$32/ay** |

**Gelir:** Test başına ₺2.999 (komisyon sonrası ~₺2.894)
- Günde 1 test: ~₺87.000/ay
- Günde 3 test: ~₺260.000/ay

---

## API Endpoints (Mevcut)

```
GET  /health                  # Tam sağlık kontrolü (DB + Redis)
GET  /liveness                # Basit canlılık kontrolü
GET  /api/v1/ping             # API test
```

Orchestrator:
```
GET  /health                  # Orchestrator sağlık
GET  /emulators               # 25 emulator listesi
GET  /emulators/status        # Boş/dolu sayıları
POST /tests/:id/start         # Test başlat (auth gerekli)
POST /tests/:id/stop          # Test durdur (auth gerekli)
GET  /tests/:id/status        # Test ilerleme (auth gerekli)
```

---

## Yasal Uyarı

Bu proje Google Play Store Hizmet Şartları'nın sınırında hareket eder. Kullanmadan önce:

- Müşterilere riskleri açıklayan sözleşme sunulmalı
- Hukuk danışmanı desteği alınmalı
- Google'ın ToS değişiklikleri yakından takip edilmeli
- Hesap kayıplarına karşı yedek plan hazır olmalı

Detaylar: [`/legal/terms`](apps/web/app/legal/terms/page.tsx), [`/legal/refund`](apps/web/app/legal/refund/page.tsx)

---

## Lisans

Özel mülk. Tüm hakları saklıdır.
