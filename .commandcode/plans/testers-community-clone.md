# Testers Community Benzeri Android Uygulama Test Hizmeti Platformu

## İçindekiler
1. [Proje Özeti](#1-proje-özeti)
2. [Mimari Genel Bakış](#2-mimari-genel-bakış)
3. [Altyapı Kurulumu](#3-altyapı-kurulumu)
4. [Hesap Yönetim Sistemi](#4-hesap-yönetim-sistemi)
5. [Emulator Farm](#5-emulator-farm)
6. [Test Otomasyon Pipeline](#6-test-otomasyon-pipeline)
7. [Web Sitesi & Ödeme](#7-web-sitesi--ödeme)
8. [Operasyon & Yasal Uyarı](#8-operasyon--yasal-uyarı)
9. [Zaman Çizelgesi](#9-zaman-çizelgesi)
10. [Dosya Yapısı](#10-dosya-yapısı)
11. [İlk Adımlar (Bugün)](#11-ilk-adımlar-bugün)
12. [Kritik Kararlar Özeti](#12-kritik-kararlar-özeti)

---

## 1. Proje Özeti

**İş Modeli:** Ücretli ajans hizmeti. Müşteri (Android geliştirici) platforma gelir, paket adı + test linkini girer, ödeme yapar. Sistem otomatik olarak 25 Google hesabıyla 14 günlük kapalı beta test sürecini başlatır, yönetir ve raporlar.

**Gelir Hedefi:** Test başına ₺2.999 (müşteri fiyatı). Iyzico komisyonu sonrası ~₺2.894 net. Günde 1 test ile ~₺87.000/ay.

**Mevcut Kaynaklar:**
- 64GB RAM Mini PC (Ubuntu 22.04 kurulacak)
- 24GB RAM Mac Mini (Apple Silicon - yardımcı)
- 1 fiziksel Android telefon
- 3 Gmail hesabı (karışık durumda, kontrol edilecek)
- Hetzner VPS (satın alınacak)
- Go + Next.js geliştirme bilgisi

---

## 2. Mimari Genel Bakış

```
┌─────────────────────────────────────────────────────────────┐
│              MÜŞTERİ (Web Tarayıcı)                         │
│              Next.js 14 Frontend (VPS)                      │
└──────────────────────────┬──────────────────────────────────┘
                           │ HTTPS
                           ▼
┌─────────────────────────────────────────────────────────────┐
│       HETZNER VPS CPX41 (16GB RAM, €25/ay)                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Go Backend API (Gin) - Port 8080                 │    │
│  │  - REST API + NextAuth uyumlu JWT                 │    │
│  │  - Asynq Worker (background queue)                │    │
│  │  - Asynq Scheduler (cron)                         │    │
│  └────────────────────────────────────────────────────┘    │
│  ┌────────────────────────────────────────────────────┐    │
│  │  PostgreSQL 15  │  Redis 7  │  Caddy             │    │
│  └────────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────────┘
                           │ PostgreSQL connection (TLS)
                           ▼
┌─────────────────────────────────────────────────────────────┐
│        MINI PC (Ev/Ofis - 64GB RAM) - Ubuntu 22.04         │
│  ┌────────────────────────────────────────────────────┐    │
│  │  Docker Compose Stack                             │    │
│  │  - 25x Android Emulator (headless, 1.5GB RAM)   │    │
│  │  - 1x Orchestrator (Go + Appium)                 │    │
│  │  - 1x ADB Server (central, port 5037)            │    │
│  │  - 1x Postgres client (yazma: activity_logs)     │    │
│  └────────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────────┘
                           │ Internet (ev IP'si)
                           ▼
              Google Play Store (25 Hesap)
```

**Neden bu yapı?**
- VPS: Kontrol paneli, müşteri arayüzü, ödeme, veritabanı. Sürekli uptime şart.
- Mini PC: 25 emulator çalıştırmak için. 64GB RAM yeterli (25 × 1.5GB = 37.5GB + OS + ADB).
- Ev IP'si: Datacenter IP'si Google tarafından şüpheli bulunur. Ev IP'si organik görünür.
- ADB Server merkezi: Tek yerden 25 emulator yönetimi.

---

## 3. Altyapı Kurulumu

### 3.1 Hetzner VPS Satın Alma

**Konfigürasyon:**
- Image: Ubuntu 22.04 LTS
- Type: CPX41 (16GB RAM, 4 vCPU, 240GB SSD) - €25/ay
- Location: Falkenstein (FSN1) - Türkiye'ye yakın ping
- SSH Key: `~/.ssh/id_ed25519.pub` ekle
- Firewall: 22, 80, 443 açık (Cloudflare arkasında olacaksa sadece 22)

**Hetzner Cloud Console'dan** sipariş verdikten sonra SSH ile bağlan:
```bash
ssh root@<SUNUCU_IP>
```

### 3.2 VPS İlk Konfigürasyon

**Kullanıcı oluşturma ve temel güvenlik:**
```bash
# 1. Yeni kullanıcı
adduser testops
usermod -aG sudo testops
mkdir -p /home/testops/.ssh
cp ~/.ssh/authorized_keys /home/testops/.ssh/
chown -R testops:testops /home/testops/.ssh

# 2. Firewall
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable

# 3. Fail2ban
apt install -y fail2ban
systemctl enable fail2ban
systemctl start fail2ban

# 4. Otomatik güvenlik güncellemeleri
apt install -y unattended-upgrades
dpkg-reconfigure -plow unattended-upgrades

# 5. Docker + Docker Compose
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
usermod -aG docker testops

# 6. Caddy (reverse proxy + otomatik HTTPS)
apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf "https://dl.cloudsmith.io/public/caddy/stable/gpg.key" | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf "https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt" | tee /etc/apt/sources.list.d/caddy-stable.list
apt update
apt install caddy
systemctl enable caddy
```

**Çıkış, testops olarak gir:**
```bash
exit
ssh testops@<SUNUCU_IP>
```

### 3.3 Proje Yapısı (Monorepo)

**Dizin yapısı:**
```
testerscommunity/
├── apps/
│   ├── api/                 # Go HTTP API (VPS'te çalışır)
│   │   ├── cmd/
│   │   │   ├── server/      # main.go - HTTP API
│   │   │   ├── worker/      # main.go - Asynq worker
│   │   │   └── scheduler/   # main.go - Asynq scheduler (cron)
│   │   ├── internal/
│   │   │   ├── config/      # viper config loader
│   │   │   ├── db/          # pgx pool
│   │   │   ├── handler/     # HTTP handlers
│   │   │   │   ├── auth.go
│   │   │   │   ├── test.go
│   │   │   │   ├── payment.go
│   │   │   │   ├── iyzico_webhook.go
│   │   │   │   └── admin.go
│   │   │   ├── service/     # business logic
│   │   │   │   ├── auth_service.go
│   │   │   │   ├── test_service.go
│   │   │   │   └── payment_service.go
│   │   │   ├── repository/  # sqlc generated
│   │   │   ├── worker/      # Asynq task handlers
│   │   │   │   ├── test_start.go
│   │   │   │   ├── daily_engagement.go
│   │   │   │   └── write_review.go
│   │   │   ├── scheduler/   # cron jobs
│   │   │   ├── middleware/  # auth, rate limit, CORS
│   │   │   ├── model/       # DTOs
│   │   │   └── lib/         # utilities
│   │   ├── migrations/      # SQL migrations (golang-migrate)
│   │   ├── queries/         # sqlc input
│   │   ├── sqlc.yaml
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── orchestrator/        # Mini PC'de çalışan Go binary
│   │   ├── cmd/orchestrator/main.go
│   │   ├── internal/
│   │   │   ├── api/         # kontrol HTTP API
│   │   │   ├── adb/         # ADB client wrapper
│   │   │   ├── appium/      # Appium HTTP client
│   │   │   ├── emulator/    # emulator manager
│   │   │   ├── profile/     # device profile
│   │   │   ├── task/
│   │   │   │   ├── opt_in.go
│   │   │   │   ├── download.go
│   │   │   │   ├── engage.go
│   │   │   │   └── review.go
│   │   │   ├── screenshot/  # local storage
│   │   │   └── config/
│   │   ├── templates/
│   │   │   └── reviews.json
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   └── web/                 # Next.js 14
│       ├── app/
│       │   ├── (marketing)/
│       │   │   └── page.tsx
│       │   ├── (auth)/
│       │   │   ├── login/page.tsx
│       │   │   └── register/page.tsx
│       │   ├── (dashboard)/
│       │   │   ├── layout.tsx
│       │   │   ├── page.tsx
│       │   │   ├── tests/
│       │   │   │   ├── new/page.tsx
│       │   │   │   └── [id]/page.tsx
│       │   │   ├── earnings/page.tsx
│       │   │   └── profile/page.tsx
│       │   ├── (admin)/
│       │   │   ├── layout.tsx
│       │   │   ├── page.tsx
│       │   │   ├── tests/page.tsx
│       │   │   ├── testers/page.tsx
│       │   │   ├── payments/page.tsx
│       │   │   └── logs/page.tsx
│       │   ├── legal/
│       │   │   ├── terms/page.tsx
│       │   │   ├── privacy/page.tsx
│       │   │   └── refund/page.tsx
│       │   └── api/
│       │       ├── auth/[...nextauth]/route.ts
│       │       └── webhooks/iyzico/route.ts
│       ├── components/
│       │   ├── ui/          # shadcn
│       │   ├── landing/
│       │   ├── dashboard/
│       │   └── admin/
│       ├── lib/
│       │   ├── api-client.ts
│       │   ├── auth.ts
│       │   └── utils.ts
│       ├── package.json
│       ├── next.config.js
│       ├── tailwind.config.js
│       └── Dockerfile
│
├── packages/
│   ├── shared/              # ortak Go types
│   │   ├── go.mod
│   │   └── types/
│   │       ├── test.go
│   │       ├── task.go
│   │       └── events.go
│   └── db/                  # sqlc generated (shared)
│       ├── go.mod
│       └── db.go
│
├── infra/
│   ├── vps/
│   │   ├── docker-compose.yml
│   │   ├── caddy/
│   │   │   └── Caddyfile
│   │   └── scripts/
│   │       ├── setup.sh
│   │       ├── deploy.sh
│   │       └── backup.sh
│   └── minipc/
│       ├── docker-compose.yml
│       └── scripts/
│           ├── setup.sh
│           ├── start-all.sh
│           └── stop-all.sh
│
├── scripts/
│   ├── account-warmup.md    # Manuel checklist
│   ├── google-groups-helper.md
│   └── generate-fingerprints.go
│
├── docs/
│   ├── ARCHITECTURE.md
│   ├── RUNBOOK.md
│   ├── LEGAL.md
│   └── DEPLOYMENT.md
│
├── .env.example
├── .gitignore
├── Makefile
├── go.work
├── package.json
├── pnpm-workspace.yaml
└── README.md
```

**Go workspaces (`go.work`):**
```go
go 1.22

use (
    ./apps/api
    ./apps/orchestrator
    ./packages/shared
    ./packages/db
)
```

**pnpm workspaces (`pnpm-workspace.yaml`):**
```yaml
packages:
  - "apps/web"
```

### 3.4 Docker Compose - VPS

**Dosya:** `infra/vps/docker-compose.yml`

```yaml
version: '3.9'

services:
  postgres:
    image: postgres:15-alpine
    restart: unless-stopped
    environment:
      POSTGRES_USER: tester
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: testers
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U tester"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    restart: unless-stopped
    command: redis-server --appendonly yes
    volumes:
      - redisdata:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s

  api:
    build: ../../apps/api
    restart: unless-stopped
    environment:
      DATABASE_URL: postgres://tester:${DB_PASSWORD}@postgres:5432/testers
      REDIS_URL: redis://redis:6379/0
      JWT_SECRET: ${JWT_SECRET}
      IYZICO_API_KEY: ${IYZICO_API_KEY}
      IYZICO_SECRET_KEY: ${IYZICO_SECRET_KEY}
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    command: ["./api", "serve"]

  worker:
    build: ../../apps/api
    restart: unless-stopped
    environment:
      DATABASE_URL: postgres://tester:${DB_PASSWORD}@postgres:5432/testers
      REDIS_URL: redis://redis:6379/0
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    command: ["./api", "worker"]

  scheduler:
    build: ../../apps/api
    restart: unless-stopped
    environment:
      DATABASE_URL: postgres://tester:${DB_PASSWORD}@postgres:5432/testers
      REDIS_URL: redis://redis:6379/0
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    command: ["./api", "scheduler"]

  asynqmon:
    image: hibiken/asynqmon:latest
    restart: unless-stopped
    environment:
      REDIS_URL: redis://redis:6379
    ports:
      - "8081:8080"
    depends_on:
      - redis

  caddy:
    image: caddy:2-alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
      - caddy_config:/config
    depends_on:
      - api
      - web

  web:
    build: ../../apps/web
    restart: unless-stopped
    environment:
      NEXT_PUBLIC_API_URL: https://api.testerscomm.net
      INTERNAL_API_URL: http://api:8080
      NEXTAUTH_URL: https://testerscomm.net
      NEXTAUTH_SECRET: ${NEXTAUTH_SECRET}
    depends_on:
      - api

volumes:
  pgdata:
  redisdata:
  caddy_data:
  caddy_config:
```

**Caddyfile (reverse proxy + otomatik HTTPS):**
```
testerscomm.net, www.testerscomm.net {
    reverse_proxy web:3000
}

api.testerscomm.net {
    reverse_proxy api:8080
}

asynq.testerscomm.net {
    basicauth {
        admin ${ASYNQ_ADMIN_PASSWORD}
    }
    reverse_proxy asynqmon:8080
}
```

### 3.5 PostgreSQL Şema Tasarımı

**Dosya:** `apps/api/migrations/0001_init.sql`

```sql
-- Extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users (müşteri + admin)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'customer', -- customer|admin
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Testers (25 Google hesabı - master listesi)
CREATE TABLE testers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_encrypted BYTEA NOT NULL, -- pgcrypto ile şifrelenmiş
    recovery_email VARCHAR(255),
    phone VARCHAR(50),
    google_group_id UUID REFERENCES google_groups(id),
    status VARCHAR(20) NOT NULL DEFAULT 'warming', -- warming|active|cooling|disabled
    device_profile_id UUID REFERENCES device_profiles(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    notes TEXT
);
CREATE INDEX idx_testers_status ON testers(status);

-- Device profiles (her hesap için cihaz kimliği)
CREATE TABLE device_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tester_id UUID REFERENCES testers(id),
    android_id VARCHAR(50) NOT NULL,
    imei VARCHAR(20) NOT NULL,
    mac_address VARCHAR(20) NOT NULL,
    model VARCHAR(50) NOT NULL,
    manufacturer VARCHAR(50) NOT NULL,
    android_version VARCHAR(10) NOT NULL,
    screen_resolution VARCHAR(20) NOT NULL,
    user_agent TEXT NOT NULL,
    locale VARCHAR(10) NOT NULL,
    timezone VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Google Groups (her test için ayrı grup)
CREATE TABLE google_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'active' -- active|archived
);

-- Tests (müşteri test işleri)
CREATE TABLE tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    package_name VARCHAR(255) NOT NULL,
    test_link TEXT,
    notes TEXT,
    star_preference VARCHAR(20) NOT NULL DEFAULT 'mixed', -- all5|mixed|custom
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending|active|completed|failed|cancelled
    google_group_id UUID REFERENCES google_groups(id),
    started_at TIMESTAMPTZ,
    ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_tests_status_ends ON tests(status, ends_at);

-- Test assignments (hangi tester hangi teste atandı)
CREATE TABLE test_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_id UUID NOT NULL REFERENCES tests(id),
    tester_id UUID NOT NULL REFERENCES testers(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending|in_progress|completed|failed
    opt_in_at TIMESTAMPTZ,
    install_at TIMESTAMPTZ,
    last_engagement_at TIMESTAMPTZ,
    review_id UUID REFERENCES reviews(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(test_id, tester_id)
);
CREATE INDEX idx_assignments_tester ON test_assignments(tester_id, status);

-- Activity logs (günlük aktivite kayıtları)
CREATE TABLE activity_logs (
    id BIGSERIAL PRIMARY KEY,
    test_assignment_id UUID NOT NULL REFERENCES test_assignments(id),
    action VARCHAR(50) NOT NULL, -- opt_in|download|install|open|interact|review
    performed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    success BOOLEAN NOT NULL,
    error_message TEXT,
    screenshot_path TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'
);
CREATE INDEX idx_logs_assignment ON activity_logs(test_assignment_id, performed_at DESC);

-- Reviews
CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    test_assignment_id UUID NOT NULL REFERENCES test_assignments(id),
    rating INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    review_text TEXT NOT NULL,
    language VARCHAR(10) NOT NULL DEFAULT 'tr',
    posted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Payments
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    test_id UUID REFERENCES tests(id),
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'TRY',
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending|completed|refunded|failed
    iyzico_token VARCHAR(255),
    iyzico_payment_id VARCHAR(255),
    paid_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_payments_user ON payments(user_id, created_at DESC);

-- System events (operasyonel loglar)
CREATE TABLE system_events (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL, -- info|warning|error|critical
    message TEXT NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_events_type_time ON system_events(event_type, created_at DESC);
CREATE INDEX idx_events_severity_time ON system_events(severity, created_at DESC) WHERE severity IN ('error', 'critical');
```

### 3.6 Ortam Değişkenleri (`.env.example`)

```bash
# Database
DB_PASSWORD=change-me-strong-password
DATABASE_URL=postgres://tester:change-me-strong-password@localhost:5432/testers
REDIS_URL=redis://localhost:6379/0

# JWT
JWT_SECRET=change-me-32-characters-minimum-length-secret
NEXTAUTH_SECRET=change-me-32-characters-minimum-length-secret

# Caddy / Domain
DOMAIN=testerscomm.net
ACME_EMAIL=admin@testerscomm.net

# Asynqmon
ASYNQ_ADMIN_PASSWORD=change-me

# Iyzico
IYZICO_API_KEY=sandbox-your-key
IYZICO_SECRET_KEY=sandbox-your-secret
IYZICO_BASE_URL=https://sandbox-api.iyzico.com

# Google Groups API (opsiyonel, ilerde)
GOOGLE_SERVICE_ACCOUNT_JSON=

# Email (SMTP)
SMTP_HOST=smtp.eu.mailgun.org
SMTP_PORT=587
SMTP_USER=postmaster@mg.testerscomm.net
SMTP_PASSWORD=change-me

# Mini PC erişim (orchestrator -> API yazma)
ORCHESTRATOR_API_TOKEN=change-me-long-random-token

# Logging
LOG_LEVEL=info
SENTRY_DSN=

# Emulator farm
EMULATOR_MAX_INSTANCES=25
ADB_SERVER_PORT=5037
```

---

## 4. Hesap Yönetim Sistemi

### 4.1 Gmail Hesap Açma Prosedürü

**Yasal Uyarı:** Google ToS'a göre otomatik hesap açma yasaktır. **Manuel** açacağız, ama script'ler süreci hızlandıracak.

**Hedef:** 25 hesap, 12-15 günde tamamlanır (günde 1-2 hesap).

**Öncelik: 3 mevcut Gmail'i değerlendir**

İlk adım mevcut 3 Gmail'i durumuna göre ayır:
- **Temiz olanlar** (yeni, kişisel aktivite yok): Direkt test için kullanılabilir
- **Kişisel olanlar** (kendi aktif hesapların): Ban riski var, **kullanma**, yeni hesap aç
- **Karışık olanlar**: Tek tek login dene, durumu kontrol et

**Karar: 3'ü de test için ayrılmamışsa 25 yeni hesap açılacak.**

**Her yeni hesap için adımlar:**

**Adım 1: Hazırlık (Mac Mini'de Safari/Firefox)**

Kişilik oluştur ve not al:
```markdown
# Tester #1
- Email: tester001.realname@gmail.com
- Password: <unique-32-char>
- Name: [Türkçe isim]
- Birthdate: 1990-05-15 (farklı gün/ay)
- Gender: Erkek/Kadın (karışık)
- Recovery email: protonmail001@proton.me
- Phone: kendi numaran (son 2 hane değiştir) VEYA SMS servisi
- Profile photo: thispersondoesnotexist.com'dan indir
```

**Adım 2: Hesap açma (accounts.google.com/signup)**

1. `accounts.google.com/signup` adresine git
2. Bilgileri gir (yukarıdaki kişiliğe göre)
3. Telefon doğrulama:
   - **İlk 5 hesap:** Kendi numaranı kullan, son 2 haneyi değiştir (Google bazen kabul eder)
   - **Sonraki 20 hesap:** SMS servisi (aşağıda detay)
4. Gizlilik: "İfadeleri görmesine izin ver" varsayılan kalsın
5. **2FA AÇMA** (test sürecini zorlaştırır)
6. Recovery email: ProtonMail adresi (5 tane yeter, döngüsel kullan)

**Adım 3: Hesap açıldıktan sonra**

1. Profil fotoğrafı yükle (thispersondoesnotexist.com)
2. Play Store'a giriş yap, **ödeme yöntemi ekleme**
3. İlk 24-72 saat: **warming** dönemi (hiçbir şey yapma)

**Adım 4: Manuel warming (3 gün)**

```bash
# Her gün 5-10 dakika:
# 1. Chrome'da gezin
# 2. Gmail'de 2-3 email oku (spam'e düşenler dahil)
# 3. YouTube'da 1-2 video izle
# 4. Google Search'te birkaç arama
# Bu, "bot değilim" sinyali verir
```

**Adım 5: PostgreSQL'e kaydet (manuel SQL)**

```sql
-- Encryption key'i .env'den al
INSERT INTO testers (email, password_encrypted, recovery_email, phone, status)
VALUES (
    'tester001.realname@gmail.com',
    pgp_sym_encrypt('password-here', 'ENCRYPTION_KEY'),
    'protonmail001@proton.me',
    '+90xxxxxxxxx',
    'warming'
);
```

### 4.2 SMS Doğrulama Servisleri

**İlk 5 hesap:** Kendi numaran + son 2 hane değişiklik (Google bazen kabul eder)
**Sonraki 20 hesap:** SMS-Activate.org veya 5sim.net (Türkiye, $0.10-0.50/numara)

**SMS-Activate API pattern (referans):**
```
GET https://api.sms-activate.org/stubs/handler_api.php?api_key=KEY&action=getNumber&service=go&country=0
→ ACCESS_NUMBER:12345:+79001234567

GET ...&action=getStatus&id=12345
→ SMS_STATUS_WAIT_CODE (henüz gelmedi) veya STATUS_FINISH:code (kod geldi)
```

**MVP'de SMS API entegrasyonu gerekmez:** Manuel olarak numara al, SMS'i oku, hesabı aç.

### 4.3 Google Groups Yönetimi

**Neden Google Groups?**
- Workspace pahalı ($6/kullanıcı/ay × 25 = $150/ay)
- Groups ücretsiz, limitsiz üye
- Play Console test kullanıcıları için yeterli

**Her test için yeni grup:**

1. `groups.google.com` adresine git
2. Yeni grup: `test-<test_id>-<customer_slug>@googlegroups.com`
3. Ayarlar:
   - Üyelik: "Yalnızca davet edilen kullanıcılar"
   - Üyelerin birbirini görmesi: Hayır (gizlilik)
4. 25 tester email'ini yapıştır
5. **Önemli:** Müşterinin geliştirici email'ini de gruba ekle (bilgilendirme)

**Notlar:**
- Google bazen tek seferde 25 üyeyi kabul etmeyebilir
- Günde 5-10 hesap ekleyerek ilerle
- Hata olursa 24 saat sonra tekrar dene

**PostgreSQL'e kayıt:**
```sql
INSERT INTO google_groups (group_email, name)
VALUES ('test-abc123-myapp@googlegroups.com', 'Test for MyApp - 2026-06');

UPDATE tests SET google_group_id = (SELECT id FROM google_groups WHERE group_email = '...')
WHERE id = 'test-uuid';
```

### 4.4 Hesap Durumu State Machine

**State'ler:**
- `warming`: Yeni açıldı, 0-3 gün warming dönemi
- `active`: Kullanılabilir
- `cooling`: Test bitti, 3 gün bekleme
- `disabled`: Ban yendi veya artık kullanılmaz

**Geçişler:**
- warming → active: 3 gün sonra (otomatik cron)
- active → cooling: test bittikten sonra (otomatik)
- cooling → active: 3 gün sonra (otomatik cron)
- * → disabled: **manuel** karar (admin panelden)

---

## 5. Emulator Farm

### 5.1 Neden Mini PC (64GB RAM)?

**RAM hesaplama:**
- 25 emulator × 1.5GB RAM = 37.5GB
- + Orchestrator: 1GB
- + ADB Server: 0.5GB
- + Ubuntu OS: 2GB
- + Docker overhead: 3GB
- = ~44GB toplam
- 64GB RAM'de 20GB headroom (yeterli)

**VPS alternatifi karşılaştırması:**
- Hetzner 64GB sunucu: ~€100/ay
- Mini PC: €500 bir kerelik
- 5 ayda amorti edilir

**Karar: Mini PC**

### 5.2 Mini PC İşletim Sistemi

**Ubuntu 22.04 LTS Server (headless, masaüstü yok)**

Kurulum:
1. Ubuntu Server ISO indir: ubuntu.com/download/server
2. Balena Etcher ile USB'ye yaz
3. Mini PC'de boot (BIOS'ta USB seç)
4. "Install Ubuntu Server"
5. Disk: tamamını kullan, LVM
6. OpenSSH server: ✓
7. Kullanıcı: `testops`

**İlk giriş:**
```bash
ssh testops@<MINIPC_IP>
```

**Docker + KVM kurulumu:**
```bash
# Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
sudo usermod -aG docker testops
# Logout/login

# KVM kontrol (emulator için gerekli)
egrep -c '(vmx|svm)' /proc/cpuinfo
# 0'dan büyük olmalı (sanallaştırma desteği)
ls /dev/kvm
# Var olmalı

# Eğer KVM yok: BIOS'a gir, Intel VT-x / AMD SVM'yi aç
```

### 5.3 Emulator Docker Compose

**Dosya:** `infra/minipc/docker-compose.yml`

**Karar: `budtmo/docker-android` kullan** (hazır, stabil, headless, ADB pre-configured)

**İlk test: 1 emulator ayağa kaldır, sonra ölçekle**

**Phase 1 (Hafta 3): 1 emulator test**
```yaml
version: '3.9'

services:
  emulator-01:
    image: budtmo/docker-android:emulator_11.0
    privileged: true
    devices:
      - /dev/kvm:/dev/kvm
    ports:
      - "5554:5554"
      - "5555:5555"
    environment:
      - ANDROID_EMULATOR_DEVICE=Pixel_5
      - WEB_VNC=false
      - EMULATOR_PARAMS=-no-window -no-audio -no-boot-anim -gpu swiftshader_indirect
    shm_size: '2gb'
    mem_limit: 3gb
    volumes:
      - ./emulator-data/01:/root/.android
    restart: unless-stopped
```

**Phase 2 (Hafta 4): 5 emulator + Appium**
```yaml
version: '3.9'

services:
  # ... emulator-01 ... emulator-05 (aynı template)

  appium:
    image: appium/appium:latest
    restart: unless-stopped
    ports:
      - "4723:4723"
    environment:
      - ANDROID_HOME=/opt/android-sdk
    volumes:
      - /opt/android-sdk:/opt/android-sdk
    depends_on:
      - emulator-01
      - emulator-02
      - emulator-03
      - emulator-04
      - emulator-05

  orchestrator:
    build: ../../apps/orchestrator
    restart: unless-stopped
    environment:
      DATABASE_URL: postgres://tester:${DB_PASSWORD}@<VPS_IP>:5432/testers
      ADB_SERVER_PORT: 5037
      APPIUM_URL: http://appium:4723
    depends_on:
      - appium
```

**Phase 3 (Hafta 9): 25 emulator (tam kapasite)**

Aynı template'i 25 kez kopyala (port: 5554 + 2*n). **NOT: RAM'i aşarsa 20'ye düşür.**

### 5.4 ADB Topolojisi

**Karar: Merkezi ADB Server** (orchestrator üzerinde)

```
Orchestrator (Go)
  └─> ADB Client (Go - electricbubble/go-to-adb)
       └─> ADB Server (port 5037) - orchestrator içinde
            ├─> emulator-01 (port 5554)
            ├─> emulator-02 (port 5556)
            ...
            └─> emulator-25 (port 5602)
```

**Orchestrator başlarken:**
```go
// 25 emulator'ü ADB'ye bağla
for i := 0; i < 25; i++ {
    port := 5554 + 2*i
    adb.Connect(fmt.Sprintf("emulator-%d", port))
}
```

### 5.5 Device Fingerprint Randomization

**Her tester için benzersiz profil oluştur:**

**Script:** `scripts/generate-fingerprints.go`

Çıktı: `infra/minipc/emulator-data/fingerprints.json`
```json
[
  {
    "tester_id": 1,
    "android_id": "a3f7b9c2d8e1f4a5",
    "imei": "354012345678901",
    "mac_address": "9C:5A:44:1B:78:3D",
    "model": "Pixel 5",
    "manufacturer": "Google",
    "android_version": "11",
    "screen_resolution": "2340x1080",
    "user_agent": "Mozilla/5.0 (Linux; Android 11; Pixel 5) AppleWebKit/537.36 Chrome/91.0.4472.114 Mobile Safari/537.36",
    "locale": "en-US",
    "timezone": "Europe/Istanbul"
  },
  ... 24 tane daha
]
```

**Varyasyon kuralları:**
- Modeller: Pixel 5, Pixel 6, Samsung S21, OnePlus 9 (karışık)
- Android: 11, 12, 13 (karışık)
- Locale: en-US, en-GB, de-DE, tr-TR (karışık)
- Timezone: Europe/Istanbul, Europe/London, Europe/Berlin, America/New_York (karışık)

### 5.6 Orchestrator API (Mini PC'de)

**Endpoint'ler:**

| Method | Path | Açıklama |
|--------|------|----------|
| GET | `/health` | Liveness check |
| GET | `/testers` | 25 tester'ın durumu |
| POST | `/tests/:id/start` | Test başlat |
| POST | `/tests/:id/stop` | Test durdur |
| GET | `/tests/:id/status` | İlerleme (JSON) |

**Test başlatma akışı (Go - pseudo):**
```go
// POST /tests/:id/start
func StartTest(testID string) error {
    // 1. DB'den test bilgilerini al (VPS'ten)
    test := db.GetTest(testID)
    assignments := db.GetAssignments(testID) // 25 tester

    // 2. Her tester için:
    for _, a := range assignments {
        // 2a. Uygun emulator bul (boş olan)
        emulator := pool.GetAvailableEmulator()

        // 2b. ADB: Google hesabıyla giriş
        adb.LoginGoogle(emulator, a.TesterEmail, a.TesterPassword)

        // 2c. Opt-in: Play Store link'e git
        adb.OptIn(emulator, test.TestLink)

        // 2d. İndir + kur
        adb.InstallApp(emulator, test.PackageName)

        // 2e. Aç + screenshot
        adb.LaunchApp(emulator, test.PackageName)
        adb.Screenshot(emulator, "/screenshots/test-{id}/tester-{tid}-day-1.png")

        // 2f. DB'ye yaz
        db.LogActivity(a.ID, "install", true, nil, screenshotPath)

        // 2g. Emulator'ü serbest bırak
        pool.ReleaseEmulator(emulator)
    }

    return nil
}
```

---

## 6. Test Otomasyon Pipeline

### 6.1 14 Günlük Schedule

**Asynq Scheduler görevleri (her test için dinamik):**

| Gün | Saat (UTC) | Aksiyon | Hesap Sayısı |
|-----|-----------|---------|--------------|
| 1 | 10:00 | Opt-in (Play Store link) | 25/25 |
| 1 | 14:00 | İndirme başlat | 25/25 |
| 1 | 18:00 | Kurulum tamamla | 25/25 |
| 1 | 20:00 | İlk açılış (30 sn bekle) | 25/25 |
| 2-13 | 09:00 | Uygulama aç, 2-5 dk etkileşim | 5/gün (rotasyon) |
| 2-13 | 15:00 | Uygulama aç, 2-5 dk etkileşim | 5/gün (rotasyon) |
| 2-13 | 21:00 | Uygulama aç, 2-5 dk etkileşim | 5/gün (rotasyon) |
| 14 | 10:00 | Son açılış + review yaz | 10/25 |
| 14 | 18:00 | Test sonu, cleanup | - |

**Neden 5 hesap/gün?**
- 25 hesap × 3 açılış = 75 session/gün (çok fazla)
- Daha azı: organik görünür, ban riski düşük
- Daha fazlası: Google flag'ler

### 6.2 Opt-in & İndirme Akışı

**Appium komutları (Go - pseudo):**

```go
// Task: OptInAndDownload
func OptInAndDownload(testID, testerID string) error {
    session := appium.NewSession(emulatorPort)
    defer session.Quit()

    // 1. Play Store link'e git (Chrome)
    session.Get(test.TestLink)
    session.WaitForElement("Become a tester", 30*time.Second)
    session.Click("Become a tester")
    session.Sleep(5 * time.Second)

    // 2. Play Store uygulamasını aç
    session.ActivateApp("com.android.vending")
    session.Sleep(3 * time.Second)

    // 3. Arama yap veya URL'ye git
    searchURL := fmt.Sprintf("market://details?id=%s", test.PackageName)
    session.Get(searchURL)
    session.Sleep(5 * time.Second)

    // 4. Install butonuna tıkla
    session.WaitForElement("Install", 30*time.Second)
    session.Click("Install")

    // 5. İndirme tamamlanmasını bekle (max 5 dakika)
    session.WaitForElement("Open", 5*time.Minute)
    session.Sleep(5 * time.Second) // Install complete

    // 6. Uygulamayı aç
    session.Click("Open")
    session.Sleep(30 * time.Second) // İlk açılış

    // 7. Screenshot al
    session.Screenshot("/screenshots/test-{id}/tester-{tid}-install.png")

    // 8. DB'ye log
    db.LogActivity(assignmentID, "install", true, nil, screenshotPath)

    return nil
}
```

**Hata yönetimi:**
- İndirme başarısız → 5 dakika sonra tekrar dene (max 3 deneme)
- "App not available" → müşteriye email gönder (paket adı yanlış olabilir)
- 3 deneme başarısız → o hesabı skip et, log'a yaz

### 6.3 Günlük Engagement (Organik Görünüm)

**Davranış pattern'leri:**

```go
// Task: DailyEngagement
func DailyEngagement(testID, testerID string) error {
    session := appium.NewSession(emulatorPort)
    defer session.Quit()

    // 1. Uygulamayı aç
    session.ActivateApp(test.PackageName)
    session.Sleep(random.Int(3, 10)) // 3-10 sn

    // 2. Rastgele etkileşim (2-5 dakika)
    duration := random.Int(120, 300) // saniye
    end := time.Now().Add(time.Duration(duration) * time.Second)

    for time.Now().Before(end) {
        action := random.Int(0, 4)
        switch action {
        case 0: // Scroll down
            session.Swipe(540, 1500, 540, 500, random.Int(300, 800))
        case 1: // Scroll up
            session.Swipe(540, 500, 540, 1500, random.Int(300, 800))
        case 2: // Tap (rastgele)
            x := random.Int(100, 980)
            y := random.Int(300, 1800)
            session.Tap(x, y)
        case 3: // Geri tuşu
            session.PressBack()
        case 4: // Bekle
            session.Sleep(random.Int(1, 5))
        }
        session.Sleep(random.Int(1, 4)) // Aksiyonlar arası bekleme
    }

    // 3. Screenshot
    session.Screenshot("/screenshots/test-{id}/tester-{tid}-day-{n}.png")

    // 4. DB'ye log
    db.LogActivity(assignmentID, "engage", true, nil, screenshotPath)

    return nil
}
```

**Önemli detaylar:**
- Her hesap farklı saatlerde çalışsın (rastgele dağılım)
- 2-5 dakika oturum süresi (çok kısa = bot, çok uzun = şüpheli)
- Rastgele swipe + tap + bekle pattern'i

### 6.4 Review Yazma

**Strateji:**
- 10/25 hesaptan review (her zaman hepsinden değil)
- Yıldız dağılımı: %70 5 yıldız, %20 4 yıldız, %10 3 yıldız
- Hazır şablonlar (15-20 adet, farklı)

**Dosya:** `apps/orchestrator/templates/reviews.json`

```json
[
  {
    "rating": 5,
    "language": "tr",
    "templates": [
      "Güzel uygulama, kullanışlı. Birkaç küçük bug var ama genel olarak başarılı.",
      "Beklediğimden iyi çıktı. Arayüz sade ve hızlı. Tavsiye ederim.",
      "Hızlı ve sorunsuz çalışıyor. Özellikler yeterli. Devamını bekliyorum."
    ]
  },
  {
    "rating": 4,
    "language": "tr",
    "templates": [
      "Güzel ama geliştirilebilir. Birkaç küçük sorun var.",
      "İyi tasarım, bazı yerlerde kasma var. Güncellemeyle düzelir umarım."
    ]
  },
  {
    "rating": 3,
    "language": "tr",
    "templates": [
      "Fena değil ama beklediğim kadar iyi değil. Bazı eksikler var."
    ]
  }
]
```

**Zamanlama:** 14. gün, 10:00-18:00 arası, günde max 2-3 review.

### 6.5 Ban Önleme Stratejileri

**Google'ın flag'leyebileceği durumlar ve önlemler:**

| Risk | Önlem |
|------|-------|
| Çok hızlı indirme (25 aynı anda) | 5'erli gruplar halinde, 30 dk arayla |
| Aynı IP (datacenter) | Ev IP'si (Mini PC) |
| Aynı cihaz fingerprint | Her hesap farklı profile |
| Ani aktivite spike | Yumuşak warming (3 gün) |
| Mükemmel pattern | Rastgele saat + süre |
| Çok fazla hesap aynı uygulamayı indiriyor | Aynı testte max 25, farklı testlerde rotasyon |

**Hata monitoring:**
- "App not available" → paket adı yanlış olabilir
- "Too many requests" → rate limit, durakla
- Login başarısız → şifre değişmiş olabilir
- 5+ hesap aynı hatayı verirse → operatöre Telegram bildirimi

---

## 7. Web Sitesi & Ödeme

### 7.1 Landing Page

**Dosya:** `apps/web/app/page.tsx`

**Bölümler:**

1. **Hero:**
   - Başlık: "Uygulamanızı 25 Gerçek Kullanıcıyla Test Edin"
   - Alt başlık: "14 günlük kapsamlı test, detaylı rapor"
   - CTA: "Hemen Başla" → /register

2. **Özellikler (3 kolon):**
   - 25 Hesap
   - 14 Gün
   - Detaylı Rapor

3. **Nasıl Çalışır (3 adım):**
   - Paket adınızı girin
   - Ödeme yapın
   - Sonuçları izleyin

4. **Fiyatlandırma:**
   - ₺2.999
   - Neler dahil: 25 hesap, 14 gün, günlük rapor, 10 review
   - CTA: "Satın Al"

5. **SSS (accordion):**
   - Yasal mı? (Risk belirt)
   - Ne kadar sürede başlar? (24 saat)
   - İade var mı? (Refund policy)
   - Hangi ülkeler? (Türkiye, ABD, Avrupa)

6. **Footer:** Yasal linkler, iletişim

### 7.2 Müşteri Kayıt & Giriş

**Auth:** NextAuth.js (Credentials provider)
- Email + password (bcrypt, 12 round)
- JWT session (httpOnly cookie, 7 gün)
- Email verification (opsiyonel MVP, sonra ekle)
- Password reset: token + email link

**Gerekli paketler:**
```bash
pnpm add next-auth bcrypt jose
```

**Auth config:** `apps/web/lib/auth.ts` (NextAuth v5)
- `apps/api` JWT'ı RS256 ile imzalar, public key `/.well-known/jwks.json` üzerinden paylaşır
- Veya: Next.js içinde imzalar, API `POST /auth/verify` ile doğrular (basit)

**MVP karar:** Next.js imzalar, API shared secret ile doğrular (HMAC).

### 7.3 Yeni Test Oluşturma Formu

**Dosya:** `apps/web/app/dashboard/tests/new/page.tsx`

**Form alanları (React Hook Form + Zod):**

```typescript
const testSchema = z.object({
  package_name: z.string()
    .regex(/^[a-z][a-z0-9_]*(\.[a-z0-9_]+)+$/, 'Geçerli bir paket adı girin (örn: com.example.app)'),
  test_link: z.string().url().optional(),
  notes: z.string().max(500).optional(),
  star_preference: z.enum(['all5', 'mixed', 'custom']).default('mixed'),
});
```

**Submit akışı:**
1. Form validasyonu (Zod)
2. API'ye POST `/api/tests` (test kaydı oluştur, status: pending)
3. Backend: Iyzico checkout form token al
4. Frontend: Iyzico iframe göster
5. Müşteri ödeme yapar
6. Iyzico webhook → API → test status: active
7. Asynq'ya "test başlat" job ekle
8. Redirect: `/dashboard/tests/[id]`

### 7.4 Müşteri Dashboard

**Dosya:** `apps/web/app/dashboard/page.tsx`

**Bölümler:**
- Aktif testler kartları (paket adı, ilerleme çubuğu gün 7/14, tamamlanan hesap 18/25)
- Geçmiş testler tablosu
- Hesap bilgileri sidebar (email, toplam harcama)

**Test detay sayfası** (`apps/web/app/dashboard/tests/[id]/page.tsx`):
- Üst bilgi (paket adı, tarih, kalan süre)
- İlerleme dashboard (chart + tablo)
- Hesap bazlı durum tablosu (maskelenmiş email, opt-in tarihi, son aktivite, review yazıldı mı)
- Aktivite timeline (gün gün ne yaptı)
- Screenshot viewer (lightbox)
- Rapor indir (PDF, test tamamlandığında)

**Real-time updates:** TanStack Query `refetchInterval: 30000` (polling yeterli MVP, WebSocket overkill).

### 7.5 Admin Panel

**Yetkilendirme:** `users.role === 'admin'`, middleware ile korunur. İlk admin SQL ile manuel.

**Sayfalar:**

1. **Dashboard (`/admin`):**
   - Aktif test sayısı
   - Bugünkü gelir
   - Aktif hesap sayısı (25 üzerinden)
   - Son hatalar (alert)

2. **Testler (`/admin/tests`):**
   - Tablo: tüm testler (filtreleme: status, tarih)
   - Detay: müşteri email, paket adı, ödeme, ilerleme
   - Aksiyonlar: durdur, manuel başlat, refund

3. **Hesaplar (`/admin/testers`):**
   - Tablo: 25 hesap (durum, son kullanım, toplam test)
   - Detay: her hesabın yaptığı tüm testler
   - Aksiyonlar: warming başlat, cooling, disable

4. **Ödemeler (`/admin/payments`):**
   - Iyzico'dan gelen tüm ödemeler
   - Eşleşen testler
   - Refund işlemleri

5. **Sistem Logları (`/admin/logs`):**
   - Tüm hatalar
   - Audit log

### 7.6 Ödeme Entegrasyonu (Iyzico)

**Neden Iyzico?**
- Türkiye pazarı için en uygun
- Taksit desteği
- TRY desteği
- Kolay entegrasyon

**Akış:**

1. **Checkout Form (Hosted):**
   - Backend: Iyzico API'ye istek → `checkoutFormContent` (HTML/iframe)
   - Frontend: iframe'i göster
   - Müşteri ödeme yapar
   - Iyzico webhook gönderir

2. **Webhook handler:** `apps/api/internal/handler/iyzico_webhook.go`
   - IP whitelist: Iyzico IP'leri
   - Signature verification
   - DB güncelle: payment status
   - Test'i active yap
   - Asynq'ya "test başlat" job ekle

3. **Refund:**
   - Admin panelden
   - Iyzico API: `POST /payment/refund`
   - DB'de payment status: refunded
   - Test durdurulur

**Test modu:**
- Iyzico sandbox: `sandbox-api.iyzico.com`
- Test kartları Iyzico docs'ta

**Go SDK:** `github.com/iyzico/iyzipay-go`

**Komisyon:** %3.49 + ₺0.25 (Iyzico)
- Müşteriye: ₺2.999
- Maliyet: ~₺105
- Net: ~₺2.894 (test başına)

---

## 8. Operasyon & Yasal Uyarı

### 8.1 Yasal Riskler (ÖNEMLİ)

**UYARI: Bu hizmet Google Play ToS'unun bazı maddelerini ihlal edebilir.**

**İlgili ToS maddeleri:**
- Play Console Distribution Agreement, Section 4: "You may not... artificially increase downloads, ratings, or reviews"
- Google Terms of Service: "You may not use multiple accounts to... bypass restrictions"
- Automated queries: Google'ın otomatik trafik tespit sistemleri var

**Olası sonuçlar:**

1. **Hesap ban:** Gmail, Google Play hesapları kalıcı kapatılabilir (**en olası**)
2. **IP ban:** Ev IP'si Google servislerinden engellenebilir (Chrome dahil)
3. **Yasal işlem:** Google nadiren dava açar, ama olabilir (özellikle büyük ölçekte)
4. **Müşteri riski:** Müşterinin uygulaması Play Store'dan kaldırılabilir (reviews spam)

**Risk azaltma:**

1. **Yasal danışmanlık (ilk 6 ay):** Türkiye'de bilişim hukuku avukatı
2. **Müşteri sözleşmesi:** Detaylı ToS, müşteri riski kabul etsin
3. **Anonimlik:** Müşteriye hesap email'lerini verme (sadece masked)
4. **Ölçek:** 25 hesap, günde max 1-2 yeni test (agresif büyüme yapma)
5. **Yedek plan:** Hesaplar banlenirse yenisini aç (döngüsel)

**Sizin sorumluluğunuz:**
- Bu riski anlıyorsunuz
- Müşterileriniz de anlamalı (ToS'ta belirtin)
- Hukuk danışmanı alın
- "İyi niyet" savunması zayıf (kar amacı)

### 8.2 Günlük Operasyon Checklist

**Sabah (10:00):**
- [ ] Asynq dashboard kontrol (kuyrukta bekleyen job var mı)
- [ ] Dünkü hata logları incele
- [ ] Aktif test sayısı (beklenenden fazla mı)
- [ ] Yeni müşteri var mı (email kontrol)
- [ ] VPS & Mini PC uptime (UptimeRobot veya basit cron)

**Öğlen (14:00):**
- [ ] Günlük engagement job'ları çalışıyor mu
- [ ] Screenshot'lar yükleniyor mu
- [ ] 1-2 hesabı manuel kontrol (Play Store'dan bak)

**Akşam (20:00):**
- [ ] Günlük review (yazılacak review var mı)
- [ ] Yarınki job'lar hazır mı (Asynq scheduler)
- [ ] Backup alındı mı (otomatik ama kontrol)

**Haftalık:**
- [ ] Ban yemiş hesap var mı kontrol (login dene)
- [ ] Yeni hesap açma planı (günde 1-2)
- [ ] Müşteri feedback'leri oku
- [ ] Gelir/gider tablosu güncelle

### 8.3 Monitoring & Alerting

**Minimal stack (ücretsiz):**

1. **Uptime monitoring:** UptimeRobot (50 monitor, ücretsiz)
2. **Error tracking:** Sentry (free tier: 5K events/ay)
3. **Log aggregation:** Dosya logları + SSH (MVP), Loki+Grafana (ilerde)
4. **Alert kanalı:** Telegram bot (ücretsiz, anlık)

### 8.4 Backup Stratejisi

**PostgreSQL:**
- Günlük `pg_dump` (cron: her gece 03:00)
- S3-compatible storage (Backblaze B2: $0.005/GB/ay)
- Retention: 30 gün

**Konfigürasyon:**
- Tüm `docker-compose.yml`, `.env` Git'te (private repo, şifreler .env.example'da)
- VPS: SSH key ile erişim (key Git'te commit etme!)

**Backup script:** `infra/vps/scripts/backup.sh`
```bash
#!/bin/bash
# VPS'te cron ile çalışır (her gece 03:00)
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="/backups/testers_$TIMESTAMP.sql.gz"

docker exec postgres pg_dump -U tester testers | gzip > $BACKUP_FILE

# B2'ye yükle (rclone ile)
rclone copy $BACKUP_FILE b2:testers-backups/

# 30 günden eski dosyaları sil
find /backups -name "testers_*.sql.gz" -mtime +30 -delete
```

### 8.5 Maliyet Tablosu (Aylık)

| Kalem | Maliyet |
|-------|---------|
| Hetzner CPX41 (16GB) | €25 (~$27) |
| Domain (testerscomm.net) | $1/ay (yıllık $12) |
| Backblaze B2 (50GB) | $0.25 |
| SMS servisi (5 hesap/ay) | $5 |
| Sentry | $0 (free tier) |
| Iyzico komisyon (%3.49) | Değişken |
| **TOPLAM sabit** | **~$35/ay** |
| **Gelir (1 test/gün × ₺2.999)** | **~₺90.000/ay** |
| **Net (1 test/gün)** | **~₺88.000/ay** |

**Not:** İlk 3 ay 0-5 test/ay (beklenen), 6. ay 30 test/ay, 12. ay 60 test/ay.

---

## 9. Zaman Çizelgesi

### 9.1 Haftalık Plan (12 Hafta)

#### **Hafta 1-2: Altyapı & Temel Kurulum**

**Hafta 1:**
- [ ] Hetzner VPS satın al, Ubuntu 22.04 kur
- [ ] VPS ilk konfigürasyon (firewall, fail2ban, Docker, Caddy)
- [ ] Mini PC'ye Ubuntu 22.04 kur
- [ ] Mini PC'de Docker + KVM kur
- [ ] Go monorepo oluştur, `go.work` ayarla
- [ ] Next.js 14 projesi oluştur
- [ ] PostgreSQL + Redis docker-compose

**Hafta 2:**
- [ ] Veritabanı şeması, migrations
- [ ] sqlc generate
- [ ] Go API: temel CRUD (users, tests)
- [ ] NextAuth.js entegrasyonu
- [ ] Landing page (hero, fiyat)
- [ ] Domain bağla, SSL (Caddy otomatik)

#### **Hafta 3-4: Hesap Yönetimi & Emulator Setup**

**Hafta 3:**
- [ ] 3 mevcut Gmail'i durumuna göre değerlendir
- [ ] Karışık/kişisel olanlar yerine yeni hesap planla
- [ ] Mac Mini'de yeni Gmail hesapları açmaya başla (günde 1-2)
- [ ] Google Groups: ilk test grubu oluştur
- [ ] Mini PC'de 1 Android emulator test (budtmo/docker-android)
- [ ] ADB bağlantısı doğrula

**Hafta 4:**
- [ ] 7 hesap daha aç (toplam 10)
- [ ] Her hesap için device profile generate et
- [ ] Multi-emulator test: 5 instance aynı anda
- [ ] Appium server kurulumu
- [ ] Orchestrator: temel HTTP API
- [ ] Orchestrator: ADB → Appium session

#### **Hafta 5-6: Test Pipeline MVP**

**Hafta 5:**
- [ ] Asynq entegrasyonu (queue + scheduler)
- [ ] Task: opt-in (Play Store link)
- [ ] Task: download & install
- [ ] Task: open app + screenshot
- [ ] Activity log repository
- [ ] Basit admin paneli: test listesi

**Hafta 6:**
- [ ] 10 hesap ile 3 günlük mini test (kendi uygulaman)
- [ ] Günlük engagement task
- [ ] Review template'leri (15 adet)
- [ ] Review yazma task
- [ ] 14 günlük schedule (Asynq scheduler)
- [ ] Hata yönetimi & retry

#### **Hafta 7-8: Web Sitesi & Ödeme**

**Hafta 7:**
- [ ] Müşteri dashboard
- [ ] Test detay sayfası (timeline)
- [ ] Screenshot viewer
- [ ] Iyzico entegrasyonu (sandbox)
- [ ] Webhook handler
- [ ] Email bildirimleri (test başladı, bitti)

**Hafta 8:**
- [ ] Admin paneli (testler, hesaplar, ödemeler)
- [ ] Refund akışı
- [ ] Yasal sayfalar (ToS, Privacy, Refund)
- [ ] Email: kayıt onayı, şifre sıfırlama
- [ ] KVKK aydınlatma metni, checkbox

#### **Hafta 9-10: Ölçeklendirme & Güvenlik**

**Hafta 9:**
- [ ] 25 hesabı tamamla (günde 1-2)
- [ ] 25 emulator aynı anda çalıştırma testi
- [ ] Performans tuning (RAM, CPU)
- [ ] Warming script (yeni hesap için)
- [ ] Cooling script (test sonrası)
- [ ] Backup otomasyonu (pg_dump + S3)

**Hafta 10:**
- [ ] Security audit (SQL injection, XSS, CSRF)
- [ ] Rate limiting (API)
- [ ] DDoS koruması (Cloudflare free tier)
- [ ] Sentry entegrasyonu
- [ ] Telegram alert bot
- [ ] Monitoring (UptimeRobot)

#### **Hafta 11-12: Lansman & İyileştirme**

**Hafta 11:**
- [ ] Beta lansman: 5 arkadaş/forum davet
- [ ] 3 gerçek test al, çalıştır
- [ ] Feedback topla, bug fix
- [ ] SEO (basit): meta tags, sitemap
- [ ] Blog/SSS güncelle

**Hafta 12:**
- [ ] Halka açık lansman (Twitter, Reddit, indie hacker)
- [ ] Ödeme → production (Iyzico live)
- [ ] İlk gerçek müşteri
- [ ] Günlük operasyon rutini otur
- [ ] 1. ay raporu, retrospect

### 9.2 Kritik Kilometre Taşları

| Tarih | Kilometre Taşı |
|-------|----------------|
| Hafta 2 sonu | VPS + Mini PC hazır, "Hello World" deploy |
| Hafta 4 sonu | 10 hesap + 5 emulator çalışıyor |
| Hafta 6 sonu | 3 günlük mini test başarılı (10 hesap) |
| Hafta 8 sonu | Web + ödeme hazır, admin paneli çalışıyor |
| Hafta 9 sonu | 25 hesap + 25 emulator stabil |
| Hafta 10 sonu | 14 günlük tam test (kendi uygulaman) başarılı |
| Hafta 11 sonu | Beta: 3 gerçek müşteri, feedback |
| Hafta 12 sonu | Public launch |

### 9.3 Risk Yönetimi

**En büyük 5 risk:**

1. **Google toplu ban (herhangi bir zaman):**
   - Önlem: Yavaş warming, günde max 2 yeni hesap
   - Plan B: Workspace'e geç (ücretli ama daha güvenli)
   - Plan C: Farklı model (sadece test grupları, kullanıcı araştırması)

2. **Mini PC RAM yetersiz (hafta 4-5):**
   - Önlem: 25 emulator test etmeden önce 10 ile dene
   - Plan B: 2. Mini PC al (yine bir kerelik €500)
   - Plan C: Hetzner'de 64GB sunucu (€100/ay, daha pahalı)

3. **Iyzico entegrasyon sorunları (hafta 7-8):**
   - Önlem: Sandbox'ta tüm akışı test et
   - Plan B: Stripe (Türkiye'de yok, yurtdışı müşteri için)
   - Plan C: Manuel ödeme (banka havalesi, yavaş)

4. **Müşteri bulamama (hafta 11+):**
   - Önlem: Beta'da 3 müşteri garantile (arkadaş çevresi)
   - Plan B: Fiverr/Upwork'te hizmet sat
   - Plan C: Indie hacker topluluğu, blog içerik

5. **Yasal sorun (herhangi bir zaman):**
   - Önlem: Hukuk danışmanı (ilk 3 ay)
   - Plan B: Model değiştir (kullanıcı araştırması platformu)
   - Plan C: Kapat, paralı model (subscription)

---

## 10. Dosya Yapısı (Referans)

```
testerscommunity/
├── apps/
│   ├── api/                          # Go backend (VPS)
│   │   ├── cmd/
│   │   │   ├── server/main.go
│   │   │   ├── worker/main.go
│   │   │   └── scheduler/main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── db/
│   │   │   ├── handler/
│   │   │   │   ├── auth.go
│   │   │   │   ├── test.go
│   │   │   │   ├── payment.go
│   │   │   │   ├── iyzico_webhook.go
│   │   │   │   └── admin.go
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── worker/
│   │   │   │   ├── test_start.go
│   │   │   │   ├── daily_engagement.go
│   │   │   │   └── write_review.go
│   │   │   ├── scheduler/
│   │   │   ├── middleware/
│   │   │   ├── model/
│   │   │   └── lib/
│   │   ├── migrations/               # SQL migrations
│   │   ├── queries/                  # sqlc input
│   │   ├── sqlc.yaml
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── orchestrator/                 # Mini PC'de çalışan binary
│   │   ├── cmd/orchestrator/main.go
│   │   ├── internal/
│   │   │   ├── api/
│   │   │   ├── adb/
│   │   │   ├── appium/
│   │   │   ├── emulator/
│   │   │   ├── profile/
│   │   │   ├── task/
│   │   │   │   ├── opt_in.go
│   │   │   │   ├── download.go
│   │   │   │   ├── engage.go
│   │   │   │   └── review.go
│   │   │   ├── screenshot/
│   │   │   └── config/
│   │   ├── templates/
│   │   │   └── reviews.json
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   └── web/                          # Next.js 14
│       ├── app/
│       ├── components/
│       ├── lib/
│       ├── public/
│       ├── package.json
│       ├── next.config.js
│       ├── tailwind.config.js
│       └── Dockerfile
│
├── packages/
│   ├── shared/                       # Ortak Go types
│   └── db/                           # sqlc generated
│
├── infra/
│   ├── vps/
│   │   ├── docker-compose.yml
│   │   ├── caddy/Caddyfile
│   │   └── scripts/
│   │       ├── setup.sh
│   │       ├── deploy.sh
│   │       └── backup.sh
│   └── minipc/
│       ├── docker-compose.yml
│       └── scripts/
│           ├── setup.sh
│           ├── start-all.sh
│           └── stop-all.sh
│
├── scripts/
│   ├── account-warmup.md             # Manuel checklist
│   ├── google-groups-helper.md
│   └── generate-fingerprints.go
│
├── docs/
│   ├── ARCHITECTURE.md
│   ├── RUNBOOK.md                    # Operasyon kılavuzu
│   ├── LEGAL.md                      # Yasal uyarılar
│   └── DEPLOYMENT.md
│
├── .env.example
├── .gitignore
├── Makefile
├── go.work
└── README.md
```

---

## 11. İlk Adımlar (Bugün)

**Plan onaylandıktan sonra ilk 3 gün yapılacaklar:**

### Gün 1: Hetzner VPS
1. Hetzner Cloud Console'dan VPS sipariş et (CPX41, Ubuntu 22.04, FSN1)
2. SSH key ekle, sunucuya bağlan
3. `testops` kullanıcısı oluştur
4. Firewall + fail2ban kur
5. Docker + Caddy kur
6. Domain kaydet (Cloudflare Registrar veya Namecheap, ~$12/yıl)

### Gün 2: Mini PC
1. Ubuntu Server 22.04 ISO indir, USB'ye yaz
2. Mini PC'de kur
3. Docker + KVM kur
4. SSH bağlantısını test et

### Gün 3: Repo
1. `testerscommunity/` dizinini oluştur
2. `go.work` + `package.json` + `pnpm-workspace.yaml` oluştur
3. `apps/api`, `apps/orchestrator`, `apps/web` iskeletlerini oluştur
4. `infra/vps/docker-compose.yml` (postgres + redis + caddy)
5. `apps/api/migrations/0001_init.sql` (PostgreSQL şeması)
6. İlk commit: `git init && git add . && git commit -m "Initial commit"`

**Sonraki 2 hafta:** Hafta 1-2 planına göre devam et.

---

## 12. Kritik Kararlar Özeti

| Konu | Karar | Neden |
|------|-------|-------|
| Stack | Go + Next.js 14 | Performans, type safety, ekosistem |
| Veritabanı | PostgreSQL 15 | Güvenilir, JSON desteği |
| Queue | Asynq (Redis) | Go-native, basit, scheduler dahil |
| Emulator | budtmo/docker-android | Hazır, stabil, headless |
| VPS | Hetzner CPX41 | Ucuz, AB lokasyonu, iyi performans |
| Hesap yönetimi | Google Groups (ücretsiz) | Workspace pahalı |
| Fingerprint | Her hesap farklı profile | Ban riski azaltma |
| Warming | Manuel + 3 gün | Google "warming" dönemi |
| Ödeme | Iyzico | Türkiye pazarı, TRY desteği |
| Orchestrator | Go binary, HTTP API | Performans, kontrol |
| Logging | Dosya + Sentry free | Düşük maliyet |
| Monitoring | UptimeRobot + Telegram | Ücretsiz, yeterli |
| Backup | pg_dump + B2 | Ucuz, güvenilir |
| İlk 25 hesap | 12-15 günde | Günde 1-2, organik |
| Test süresi | 14 gün | Endüstri standardı |
| Auth | NextAuth.js (HMAC) | Basit, hızlı |
| Frontend state | TanStack Query + Zustand | Modern standart |
| Forms | React Hook Form + Zod | Type-safe validation |
| UI | shadcn/ui + Tailwind | Hızlı geliştirme |
| Email | SMTP (Mailgun/Resend) | Kolay entegrasyon |
| OS (VPS) | Ubuntu 22.04 LTS | Stabil, geniş destek |
| OS (Mini PC) | Ubuntu 22.04 LTS Server | Docker + KVM uyumu |
| Reverse proxy | Caddy | Otomatik HTTPS, basit |

---

## 13. Doğrulama (Verification)

**Her hafta sonunda kontrol edilecekler:**

**Hafta 2 sonu:**
- [ ] VPS'te `docker compose ps` ile 5 servis ayakta mı
- [ ] `curl http://localhost:8080/health` → 200 OK
- [ ] Web `pnpm dev` → http://localhost:3000 açılıyor mu
- [ ] Mini PC'de `docker run hello-world` çalışıyor mu

**Hafta 4 sonu:**
- [ ] 10 Gmail hesabı açık mı (login dene)
- [ ] 1 emulator'da Play Store'a giriş yapabiliyor mu
- [ ] Orchestrator `GET /testers` → 10 tester dönüyor mu
- [ ] Appium session oluşturulabiliyor mu

**Hafta 6 sonu:**
- [ ] 3 günlük test tamamlandı mı (10 hesap)
- [ ] Activity log'lar DB'ye yazıldı mı
- [ ] 1 review yazıldı mı
- [ ] Screenshot'lar `/screenshots/` dizininde mi

**Hafta 8 sonu:**
- [ ] Landing page yayında mı (HTTPS)
- [ ] Müşteri kayıt + login çalışıyor mu
- [ ] Iyzico sandbox ödeme geçti mi
- [ ] Webhook test edildi mi

**Hafta 9 sonu:**
- [ ] 25 hesap + 25 emulator aynı anda çalışıyor mu
- [ ] RAM kullanımı 50GB altında mı
- [ ] 1 tam test (14 gün) kendi uygulamanla başarıyla tamamlandı mı

**Hafta 10 sonu:**
- [ ] Sentry alert geldi mi (hata simülasyonu)
- [ ] Backup otomatik alınıyor mu
- [ ] Telegram alert çalışıyor mu

**Hafta 12 sonu:**
- [ ] 3 gerçek müşteri testi başarıyla tamamlandı mı
- [ ] Müşteri dashboard'undan ilerleme canlı izlenebiliyor mu
- [ ] Iyzico live ödeme alındı mı

---

## 14. Yasal Uyarı (Tekrar)

**Bu proje Google Play Store ToS'unun bazı maddelerini ihlal edebilir:**

- ❌ "Automated queries" - otomatik hesap açma, otomatik review
- ❌ "Artificially increase downloads" - 25 hesaptan aynı anda indirme
- ❌ "Bypass restrictions" - 25 hesap, aynı kişi
- ❌ "Multiple accounts" - bir kişi, çok hesap

**Sonuçlar:**
- Google hesapları toplu banleyebilir (en olası)
- Google Play Store'dan müşteri uygulaması kaldırılabilir
- Nadir de olsa yasal işlem (dava)
- IP ban (Chrome dahil tüm Google servisleri)

**Sizin sorumluluğunuz:**
- Bu riski anlıyorsunuz
- Müşterileriniz de anlamalı (ToS'ta belirtin)
- Hukuk danışmanı alın
- "İyi niyet" savunması zayıf (kar amacı)

**Alternatif yasal modeller (ilerde değerlendirin):**
- Kullanıcı araştırması platformu (gerçek kullanıcılar, test grubu değil)
- Sadece "test grupları" (Google'ın izin verdiği resmi program)
- Beta tester topluluğu (üyelik bazlı, doğrudan müşteriye bağlı)

---

**Bu plan, kod yazmadan önce yapılacak tüm hazırlıkları içeriyor. Her adım somut ve uygulanabilir.**

**İlk 3 gün:** Hetzner VPS + Mini PC Ubuntu + Repo iskeleti
**İlk 30 gün:** Altyapı + 10 hesap + basit test pipeline + ilk mini test
**İlk 60 gün:** Web + ödeme + ilk gerçek müşteri (beta)
**İlk 90 gün:** Public launch + 25 hesap stabil + günlük operasyon rutini

Bu planı takip ederek 3 ayda çalışan bir MVP'ye sahip olacaksınız. Başarılar!
