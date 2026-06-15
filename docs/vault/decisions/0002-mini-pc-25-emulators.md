---
tags: [adr, infrastructure, mini-pc, emulator]
status: accepted
date: 2026-05-15
---

# ADR-0002: Mini PC for 25 emulators

## Context

Google Play closed test 14 günlük süreçte 25 farklı Google hesabı gerektirir. Her hesap eşzamanlı, gerçek cihaz görünümünde davranmalı.

**Alternatifler:**

1. **Cloud emulator farm** (BrowserStack, Sauce Labs, Firebase Test Lab)
   - Maliyet: ayda $200-500 (25 cihaz × 14 gün)
   - Vendor lock-in
   - Tespit riski (datacenter IP'leri, ortak cihaz fingerprint)
   - Hetzner'da zaten VPS var, neden ek cost?

2. **VPS'te emulator** (Hetzner CCX13)
   - 4 vCPU, 16GB RAM → max 3-4 emulator
   - Nested virtualization gerekli
   - KVM yok

3. **Raspberry Pi cluster** (4-5 board)
   - ARM, Android emulator desteği zayıf
   - Stability sorunu
   - 64GB RAM yok

4. **Eski PC** (kullanıcının elinde olan)
   - 64GB RAM Mini PC mevcut
   - Ubuntu Server kurulu
   - Maliyet: $0 (zaten var)
   - Lokal kontrol, düşük latency

## Decision

Kullanıcının 64GB RAM Ubuntu Server Mini PC'si kullanılacak. 25 emulator konteyner olarak çalışacak (budtmo/docker-android).

## Consequences

### Olumlu

- Maliyet: 0 (mevcut donanım)
- Lokasyon bağımsızlığı (data center IP sorunu yok, ev broadband IP'si)
- Tam kontrol (kernel seviyesi, networking, ADB)
- Test edilebilirlik: 25 hesap için yeterli RAM (~2GB/emulator × 25 = 50GB, 64GB'de sığar)
- Aynı IP üzerinden 25 hesap (gerçekçi)

### Olumsuz

- 7/24 çalışma gerekir (elektrik, soğutma, internet uptime)
- 30-60dk cold boot (25 emulator ilk açılışta)
- 25 emulator aynı anda → aynı IP → Google risk artışı
- Donanım arızası = 25 hesap çöker

### Mitigations

- UPS bağlantısı (plan)
- Otomatik restart (lifecycle manager 60s health check)
- Hesap rotasyonu (tester_daily_usage tablosu)
- Enterprise pakette farklı IP/VPN (post-launch)

## Implementation Notes

- `infra/minipc/docker-compose.yml` — 25 emulator tanımı (`emulator-0001` ... `emulator-0025`)
- `infra/minipc/deploy/docker-compose.yml` — orchestrator + appium
- `apps/orchestrator/internal/emulator/pool.go` — 25 entry
- `apps/orchestrator/internal/lifecycle/` — boot + health monitor
- KVM gerekli (`/dev/kvm` mount)

## Status

**Accepted** — 2026-05-15. Hardware mevcut, deploy beklemede (IP henüz paylaşılmadı).
