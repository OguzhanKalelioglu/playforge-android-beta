# Mimari Dokümanı

## Sistem Bileşenleri

```
┌─────────────────────────────────────────────────────────────┐
│              MÜŞTERİ (Web Tarayıcı)                         │
│              Next.js 14 Frontend (VPS)                      │
└──────────────────────────┬──────────────────────────────────┘
                           │ HTTPS (Caddy)
                           ▼
┌─────────────────────────────────────────────────────────────┐
│       HETZNER VPS CPX41 (16GB RAM, €25/ay)                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Caddy (reverse proxy + otomatik HTTPS)           │    │
│  │  - testerscomm.net → web:3000                     │    │
│  │  - api.testerscomm.net → api:8080                 │    │
│  │  - asynq.testerscomm.net → asynqmon:8080 (auth)   │    │
│  └────────────────────────────────────────────────────┘    │
│  ┌────────────────────────────────────────────────────┐    │
│  │  api (Go) - HTTP API                               │    │
│  │  worker (Go) - Asynq background jobs              │    │
│  │  scheduler (Go) - Asynq cron                       │    │
│  │  web (Next.js 14, standalone)                      │    │
│  └────────────────────────────────────────────────────┘    │
│  ┌────────────────────────────────────────────────────┐    │
│  │  PostgreSQL 15 + Redis 7 (kalıcı volume)          │    │
│  └────────────────────────────────────────────────────┘    │
│  ┌────────────────────────────────────────────────────┐    │
│  │  asynqmon (queue dashboard, basic auth)            │    │
│  └────────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────────┘
                           │ PostgreSQL connection (TLS)
                           │ (Hetzner internal network + firewall)
                           ▼
┌─────────────────────────────────────────────────────────────┐
│        MINI PC (Ev/Ofis - 64GB RAM) - Ubuntu 22.04         │
│  ┌────────────────────────────────────────────────────┐    │
│  │  orchestrator (Go binary)                          │    │
│  │  - HTTP API (kontrol için)                         │    │
│  │  - Merkezi ADB Server (port 5037)                  │    │
│  │  - Emulator pool yönetimi                          │    │
│  │  - Görev kuyruğu (opt-in, download, engage, review)│    │
│  └────────────────────────────────────────────────────┘    │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Android Emulator'ler (1-25 instance)              │    │
│  │  - budtmo/docker-android image                     │    │
│  │  - Her biri: 1.5GB RAM, 2 core, headless           │    │
│  │  - KVM üzerinde, /dev/kvm mount                    │    │
│  │  - ADB port: 5554, 5556, 5558... (2*n+5554)        │    │
│  └────────────────────────────────────────────────────┘    │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Appium Server (Phase 2+)                          │    │
│  │  - Node.js Appium + UiAutomator2 driver            │    │
│  │  - Test otomasyonu (tap, swipe, screenshot)        │    │
│  └────────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────────┘
                           │ Internet (ev IP'si)
                           ▼
              Google Play Store (25 Hesap)
```

## Veri Akışı

### 1. Müşteri Test Süreci Başlatma

```
1. Müşteri web sitesine giriş yapar
   ↓
2. /dashboard/tests/new sayfasında formu doldurur
   - package_name: com.example.app
   - test_link: https://play.google.com/apps/testing/com.example.app
   - notes, star_preference
   ↓
3. Next.js → API'ye POST /api/v1/tests
   ↓
4. Go API: tests tablosuna pending kayıt ekler
   ↓
5. Go API: Iyzico Checkout Form token alır
   ↓
6. Frontend: Iyzico iframe'i gösterir, müşteri ödeme yapar
   ↓
7. Iyzico webhook → /api/v1/webhooks/iyzico
   ↓
8. Go API: payments tablosunu günceller
   Go API: tests.status = 'active' yapar
   Go API: test_assignments tablosuna 25 kayıt ekler
   Go API: Google Groups oluşturur (admin SDK veya manuel)
   Go API: Asynq'ya "test_start" job ekler
```

### 2. Test Başlatma (Asynq Worker → Orchestrator)

```
9. Asynq worker, "test_start" job'ı alır
   ↓
10. Worker → Orchestrator'a POST /tests/:id/start
    (X-API-Token header ile auth)
    ↓
11. Orchestrator:
    a) 25 emulator'ü ADB'ye bağlar
    b) Her tester için:
       - Google hesabıyla giriş
       - Test link'e git
       - "Become a tester" tıkla
       - Play Store'da uygulamayı indir
       - İlk açılış + screenshot
    c) Her adımda activity_logs tablosuna yazar
    d) Test başlangıç zamanı + 14 gün = ends_at kaydeder
    ↓
12. Orchestrator → API'ye POST /tests/:id/assignments/:aid/activity
    (her adım için DB'ye yazma)
```

### 3. 14 Günlük Engagement (Asynq Scheduler)

```
13. Asynq scheduler, her test için 14 günlük cron'lar oluşturur:
    - Gün 1, 10:00 → opt_in (25 hesap)
    - Gün 1, 14:00 → download (25 hesap)
    - Gün 1, 18:00 → install (25 hesap)
    - Gün 1, 20:00 → first_open (25 hesap)
    - Gün 2-13, 09:00/15:00/21:00 → engage (5 hesap rotasyon)
    - Gün 14, 10:00 → final_open + review (10 hesap)
    ↓
14. Scheduler her cron tick'inde Asynq queue'ya job ekler
    ↓
15. Worker, "engage" job'ı alır
    ↓
16. Worker → Orchestrator'a POST /tests/:id/engage
    ↓
17. Orchestrator, uygun emulator'de:
    - Uygulamayı açar
    - 2-5 dakika rastgele etkileşim (swipe, tap, bekle)
    - Screenshot alır
    - activity_logs tablosuna yazar
    ↓
18. Müşteri dashboard'da gerçek zamanlı ilerleme görür
    (TanStack Query, 30 saniye polling)
```

## Bileşen Detayları

### Go API (apps/api)

**Modüller:**
- `cmd/server`: HTTP API (Gin framework)
- `cmd/worker`: Asynq worker (background jobs)
- `cmd/scheduler`: Asynq scheduler (cron)
- `internal/config`: Viper config loader
- `internal/db`: pgx pool
- `internal/lib`: logger, middleware, redis helper
- `internal/handler`: HTTP handlers
- `internal/service`: business logic (auth, payment, test)
- `internal/repository`: sqlc generated queries
- `internal/worker`: Asynq task handlers
- `internal/scheduler`: cron job definitions

**Endpoint'ler (Mevcut):**
- `GET /health` — DB + Redis sağlık kontrolü
- `GET /liveness` — Basit canlılık
- `GET /api/v1/ping` — API test

**Endpoint'ler (Hafta 7-8):**
- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `POST /api/v1/auth/forgot-password`
- `GET /api/v1/tests` (müşteri kendi testleri)
- `POST /api/v1/tests` (yeni test oluştur)
- `GET /api/v1/tests/:id`
- `POST /api/v1/tests/:id/checkout` (Iyzico token al)
- `GET /api/v1/tests/:id/activity` (timeline)
- `GET /api/v1/tests/:id/screenshots/:aid`
- `POST /api/v1/webhooks/iyzico`

**Endpoint'ler (Hafta 8 - admin):**
- `GET /api/v1/admin/tests` (tümü)
- `GET /api/v1/admin/testers`
- `GET /api/v1/admin/payments`
- `POST /api/v1/admin/payments/:id/refund`
- `GET /api/v1/admin/logs`

### Go Orchestrator (apps/orchestrator)

**Modüller:**
- `cmd/orchestrator`: Ana binary
- `internal/api`: HTTP API (kontrol)
- `internal/adb`: ADB client wrapper
- `internal/appium`: Appium HTTP client (Phase 2+)
- `internal/emulator`: Pool yönetimi (acquire/release)
- `internal/profile`: Device fingerprint
- `internal/task`: Görev implementasyonları
  - `opt_in.go`: Play Store opt-in
  - `download.go`: İndirme + kurulum
  - `engage.go`: Günlük engagement
  - `review.go`: Review yazma

**Şu anki durum:** İskelet + emulator pool + ADB bağlantısı hazır. Görev implementasyonları Hafta 5-6'da eklenecek.

### Next.js Web (apps/web)

**Routes:**
- `(marketing)/`: Landing page (`/`)
- `(auth)/login`, `(auth)/register`: Auth sayfaları
- `(dashboard)/`: Müşteri paneli
  - `/dashboard` (ana sayfa)
  - `/dashboard/tests/new`
  - `/dashboard/tests/[id]`
  - `/dashboard/earnings`
  - `/dashboard/profile`
- `(admin)/`: Admin paneli (Hafta 8)
- `legal/terms`, `legal/privacy`, `legal/refund`

**State management:**
- TanStack Query (server state, 30s polling)
- Zustand (client state, hafif)
- React Hook Form + Zod (form + validation)
- NextAuth.js v5 (auth)

**UI:**
- shadcn/ui pattern (Button, Card, Badge, Input, Label)
- Tailwind CSS
- Lucide React icons
- Mobile-first responsive

### PostgreSQL Şeması

**Ana tablolar:**

- `users`: Müşteri + admin hesapları
- `testers`: 25 Google hesabı (şifreler pgcrypto ile encrypted)
- `device_profiles`: Her hesap için cihaz kimliği (Android ID, IMEI, MAC, model, locale, timezone)
- `google_groups`: Her test için oluşturulan grup
- `tests`: Müşteri test işleri
- `test_assignments`: Hangi tester hangi teste atandı (25 × N)
- `activity_logs`: Günlük aktivite (opt_in, download, install, open, interact, review)
- `reviews`: Yazılan yorumlar
- `payments`: Iyzico ödeme kayıtları
- `system_events`: Operasyonel loglar (error, warning)

**İndeksler:**
- `tests(status, ends_at)` — aktif testleri bul
- `test_assignments(tester_id, status)` — tester bazlı sorgu
- `activity_logs(test_assignment_id, performed_at DESC)` — timeline
- `payments(user_id, created_at DESC)` — müşteri geçmişi

### Redis Kullanımı

**Asynq tarafından:**
- Queue (FIFO görev kuyruğu)
- Scheduler (cron kalıcılığı)
- Retry/backoff state

**Uygulama tarafından:**
- Rate limiting (IP bazlı)
- Distributed lock (tester bazlı, aynı anda 2 test çalışmasın)
- Session cache (opsiyonel, MVP'de JWT yeterli)

### Caddy Reverse Proxy

**Sorumlulukları:**
- Otomatik HTTPS (Let's Encrypt)
- HTTP/3 desteği
- Domain → service routing
- Güvenlik header'ları (HSTS, X-Content-Type-Options)
- Basic auth (Asynqmon için)

## Veri Akış Diyagramları

### 14 Günlük Test Yaşam Döngüsü

```
                    ┌──────────┐
                    │ pending  │  (müşteri ödeme yaptı, test başlamadı)
                    └────┬─────┘
                         │ webhook payment completed
                         ▼
                    ┌──────────┐
                    │ active   │  (25 hesap opt-in + download yapıyor)
                    └────┬─────┘
                         │ günlük engagement (gün 2-13)
                         │ final review (gün 14)
                         ▼
                    ┌──────────┐
                    │completed │  (14 gün doldu, rapor hazır)
                    └──────────┘
                    veya
                    ┌──────────┐
                    │ failed   │  (paket adı yanlış, vb.)
                    └──────────┘
                    veya
                    ┌──────────┐
                    │cancelled │  (admin iptal etti / refund)
                    └──────────┘
```

### Tester Yaşam Döngüsü

```
                    ┌──────────┐
                    │ warming  │  (yeni hesap, 0-3 gün)
                    └────┬─────┘
                         │ 3 gün sonra cron
                         ▼
                    ┌──────────┐
                    │ active   │  (test atanabilir)
                    └────┬─────┘
                         │ test bitti
                         ▼
                    ┌──────────┐
                    │ cooling  │  (3 gün bekleme)
                    └────┬─────┘
                         │ 3 gün sonra cron
                         ▼
                    ┌──────────┐
                    │ active   │  (tekrar kullanılabilir)
                    └──────────┘

Herhangi bir durumdan:
                    ┌──────────┐
                    │disabled  │  (ban, manuel kapatma)
                    └──────────┘
```

## Güvenlik

**Network:**
- VPS firewall (UFW): sadece 22, 80, 443
- Docker internal network: servisler arası iletişim izole
- Caddy: TLS 1.3, HSTS, güvenlik header'ları

**Auth:**
- JWT (HS256, 32+ secret) — httpOnly cookie
- Bcrypt (12 round) — şifre hash
- API token (orchestrator → API) — random 64 char

**Veri:**
- Tester şifreleri: pgcrypto symmetric encryption (AES-256)
- HTTPS: tüm dış trafiğ
- Backup: pg_dump + Backblaze B2 (encrypted at rest)

**Rate limiting (Hafta 10):**
- API: 60 req/min per IP (sliding window, Redis)
- Auth: 5 req/min per IP (brute force koruması)
- Webhook: IP whitelist (Iyzico)

## Ölçeklendirme

**Şu an (Phase 1):** 1 emulator
**Hafta 4 (Phase 2):** 5 emulator
**Hafta 9 (Phase 3):** 25 emulator
**İleri (Phase 4):** 50+ emulator (2. Mini PC veya Hetzner 64GB)

**Yatay ölçeklendirme:**
- 1 API sunucusu (Hetzner CPX41 yeterli, 16GB)
- 1-2 Worker (API ile aynı sunucuda)
- 1 Scheduler (cron için 1 instance yeterli)
- 1-3 Mini PC (her biri 25 emulator)

**Dikey ölçeklendirme:**
- Mini PC RAM: 128GB'a çıkarılabilir (45 emulator)
- VPS: CPX51 (32GB) veya CCX63 (48GB dedicated vCPU)

## Monitoring (Hafta 10)

**Prometheus metrikleri:**
- API: request rate, latency, error rate
- Worker: queue length, processing time, failure rate
- Emulator: uptime, boot time, connection status
- PostgreSQL: connection count, slow queries
- Redis: memory, hit rate

**Grafana dashboard'ları:**
- Sistem overview (CPU, RAM, disk, network)
- API performance (RPS, P50/P95/P99 latency)
- Queue health (job/sec, retry rate, DLQ)
- Business metrics (günlük test, gelir, aktif hesap)

**Alert (Telegram bot):**
- API 5xx > 5/dakika
- Queue length > 1000
- Emulator offline > 3
- Disk usage > 80%
- Backup failed

## Disaster Recovery

**Senaryolar:**

1. **VPS çöktü:**
   - Hetzner snapshot'tan 5 dakikada geri yükleme
   - PostgreSQL backup → B2'den (son 24 saat max data loss)

2. **Mini PC çöktü:**
   - Yedek Mini PC'de aynı compose çalıştır
   - Emulator data (tester cache) kaybolur → yeni hesap aç

3. **Google toplu ban:**
   - 25 hesabın tamamı kapanır
   - Yeni 25 hesap aç (günde 1-2, 12-15 gün)
   - Bu sırada yeni test alımı durdurulur

4. **Database corruption:**
   - B2'den backup → yeni PostgreSQL'e restore
   - RPO (Recovery Point Objective): 24 saat
   - RTO (Recovery Time Objective): 1 saat

## Limitler ve Kısıtlamalar

- **Eşzamanlı test:** 1 (1 müşteri testi = 25 hesap = 25 emulator, tümü dolu)
- **Günlük yeni test:** 1-2 (ban riski)
- **Aylık yeni hesap:** 30-60 (günde 1-2 hesap, 25'e tamamla)
- **API rate limit:** 60 req/min per IP
- **Webhook timeout:** 30 saniye (Iyzico)
- **Screenshot saklama:** 30 gün (sonra otomatik sil)
- **Backup retention:** 30 gün (B2'de)
