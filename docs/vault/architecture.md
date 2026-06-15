---
tags: [architecture, moc]
---

# Mimari Genel Bakış

TestersCommunity iki ana deployment'a yayılmıştır:

```
┌──────────────────────────┐                ┌──────────────────────────┐
│  VPS (Hetzner CCX13)     │                │  Mini PC (Ubuntu Server) │
│                          │                │                          │
│  Nginx (SSL + proxy)    │                │  Orchestrator (Go)       │
│   ├─ api.tc.net:443     │◄────REST──────►│   └─ 25 emulators (ADB)  │
│   └─ tc.net:443 (web)   │                │  Appium server           │
│                          │                │  25 × budtmo/android     │
│  PostgreSQL 16           │                │                          │
│  Redis 7 (Asynq)        │                │                          │
│  Asynqmon (queue UI)     │                │                          │
│                          │                │                          │
│  Go binaries:            │                │                          │
│   ├─ api (Gin)          │                │                          │
│   ├─ worker (Asynq)     │                │                          │
│   └─ scheduler          │                │                          │
│                          │                │                          │
│  Next.js 14 standalone  │                │                          │
└──────────────────────────┘                └──────────────────────────┘
```

## Veri Akışı (Mutlu Yol)

1. Müşteri landing'ten plan seçer → `/dashboard/new`
2. `POST /api/v1/orders` → Stripe Checkout Session oluşturur → `payment_url` döner
3. Müşteri Stripe hosted page'de kart bilgisi girer → 3D Secure
4. Stripe webhook → `POST /api/v1/payments/stripe/webhook` → signature verify → order `paid` işaretlenir
5. Scheduler 14-günlük planı Asynq'ya ekler: gün 0 = test_start, gün 1-13 = daily_engagement, gün 14 = write_review
6. Worker Asynq job'ı consume eder → orchestrator'a `POST /v1/tasks/{type}/start`
7. Orchestrator pool'dan ready emulator alır (CAS), Appium session açar, 2-5dk engagement yapar
8. Her step'te orchestrator → API'ye `POST /v1/activity` ile event yollar
9. Müşteri dashboard'dan canlı izler; admin panel'den fleet yönetilir

## İlgili

- [[services/api]] — REST API
- [[services/orchestrator]] — Emulator farm
- [[services/worker]] — Job consumer
- [[services/scheduler]] — Planlayıcı
- [[services/web]] — Dashboard
- [[database-schema]] — DB şeması
- [[payment-flow]] — Stripe Checkout
- [[task-runner]] — Pipeline detayı
- [[deployment]] — Deploy rehberi
