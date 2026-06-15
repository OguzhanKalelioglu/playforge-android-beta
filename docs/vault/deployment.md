---
tags: [deployment, infrastructure, vps, minipc]
---

# Deployment

İki ayrı deployment: **VPS** (control plane) + **Mini PC** (emulator farm).

## VPS (Hetzner CCX13 / eşdeğeri)

`infra/vps/docker-compose.yml` — 7 servis:

| Servis | Image | Port (internal) | Ağ |
| --- | --- | --- | --- |
| nginx | nginx:1.27-alpine | 80, 443 | proxy |
| postgres | postgres:16-alpine | 5432 | internal |
| redis | redis:7-alpine | 6379 | internal |
| api | custom (Go) | 8080 | proxy, internal |
| worker | custom (Go) | — | internal |
| scheduler | custom (Go) | — | internal |
| asynqmon | hibiken/asynqmon | 8080 | internal |
| web | custom (Next.js) | 3000 | proxy |

### Domain Yapısı

- `api.testerscomm.net` → nginx → api:8080
- `testerscomm.net` → nginx → web:3000
- `queue.testerscomm.net` → nginx → asynqmon:8080 (basic auth + IP whitelist)

### İlk Kurulum

```bash
# VPS'e SSH
ssh root@<VPS_IP>

# Docker + compose
apt install -y docker.io docker-compose-plugin

# Repo clone
git clone https://github.com/testerscommunity/testerscommunity.git
cd testerscommunity

# Env
cp infra/vps/.env.example infra/vps/.env
nano infra/vps/.env  # STRIPE_*, JWT_SECRET, DB_PASSWORD, vs.

# SSL (Let's Encrypt)
apt install -y certbot
certbot certonly --standalone -d api.testerscomm.net -d testerscomm.net -d queue.testerscomm.net
cp /etc/letsencrypt/live/testerscomm.net/fullchain.pem infra/vps/certs/
cp /etc/letsencrypt/live/testerscomm.net/privkey.pem   infra/vps/certs/

# Up
cd infra/vps
docker compose up -d

# Verify
docker compose ps
curl https://api.testerscomm.net/liveness
```

### Migrations

Postgres container ilk açılışta `apps/api/migrations/` altındaki SQL'leri otomatik çalıştırır (`/docker-entrypoint-initdb.d`). Sonradan eklenen migration'lar için:

```bash
docker compose exec postgres psql -U tester -d testers -f /docker-entrypoint-initdb.d/0005_stripe.sql
```

## Mini PC (Ubuntu Server 22.04, 64GB RAM)

`infra/minipc/deploy/docker-compose.yml`:

| Servis | Image | Port |
| --- | --- | --- |
| appium | appium/appium:2.11.0 | 4723 |
| orchestrator | custom (Go) | 9000 |
| emulator-0001 ... 0025 | budtmo/docker-android | ADB 5555-5579 |

### Kurulum

```bash
# Mini PC'ye SSH
ssh <user>@<MINIPC_IP>

# KVM (emulator için zorunlu)
sudo apt install -y qemu-kvm libvirt-clients libvirt-daemon-system
sudo usermod -aG kvm $USER

# Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Repo
git clone https://github.com/testerscommunity/testerscommunity.git
cd testerscommunity/infra/minipc/deploy

# Env
cp .env.example .env
nano .env  # VPS_HOST, ORCHESTRATOR_API_TOKEN, ACTIVITY_API_URL

# Up
docker compose up -d
docker compose logs -f orchestrator
```

### Emulator'lar

`infra/minipc/docker-compose.yml` 25 emulator tanımı içerir (`emulator-0001` ... `emulator-0025`).

İlk açılışta:
- Her emulator 30-60dk boot olur (system image pull)
- Orchestrator ADB connect + ready olmasını bekler
- 25/25 ready olduktan sonra pool "ready" durumuna geçer

### İlk Hesap Ekleme (Manuel)

```sql
INSERT INTO testers (email, password_encrypted, google_group_id, device_profile_id, status)
VALUES ('tester01@gmail.com', '\x...', '<group_uuid>', '<profile_uuid>', 'warming');
```

3 gün warming (günde 1-2 oturum, 5dk casual browse), sonra `active`.

## Backup

```bash
# VPS'te cron: günlük 03:00 UTC
0 3 * * * cd /opt/testerscommunity/infra/vps && docker compose exec -T postgres pg_dump -U tester testers | gzip > /backup/testers-$(date +\%F).sql.gz
```

## İlgili

- [[architecture]]
- [[services/api]] — VPS binary
- [[services/orchestrator]] — Mini PC binary
- `infra/vps/` — VPS compose + nginx
- `infra/minipc/deploy/` — Mini PC compose
- `infra/minipc/docker-compose.yml` — 25 emulator tanımı
