---
tags: [database, postgres, schema]
---

# PostgreSQL Şeması

PostgreSQL 16, VPS üzerinde Docker container olarak çalışır. Migration'lar `apps/api/migrations/` altında sıralı SQL dosyalarıdır ve container başlangıcında otomatik uygulanır.

## Tablolar (Domain Grupları)

### Kullanıcılar & Auth

| Tablo | Amaç | Migration |
| --- | --- | --- |
| `users` | Müşteri + admin hesapları | 0001 |
| `sessions` | Refresh token storage (SHA256 hash) | 0004 |
| `password_reset_tokens` | Şifre sıfırlama | 0004 |
| `email_verification_tokens` | E-posta doğrulama | 0004 |

### Hesap Yönetimi

| Tablo | Amaç | Migration |
| --- | --- | --- |
| `google_groups` | Test başına Google grubu | 0001 |
| `device_profiles` | Cihaz parmak izi (fingerprint) | 0001 |
| `testers` | 25 Google hesabı (warming/active/cooling) | 0001 |
| `tester_daily_usage` | Günlük görev/dakika sayacı | 0004 |

### Test & Aktivite

| Tablo | Amaç | Migration |
| --- | --- | --- |
| `tests` | Müşteri test işleri | 0001 |
| `test_assignments` | Tester-test eşlemesi | 0001 |
| `reviews` | Play Store yorumları | 0001 |
| `activity_logs` | Step event log (orchestrator ingest) | 0001 |

### Plan & Ödeme

| Tablo | Amaç | Migration |
| --- | --- | --- |
| `plan_tiers` | Basic/Pro/Enterprise ürün paketleri | 0004 |
| `orders` | Checkout öncesi/sonrası sipariş | 0004 |
| `payments` | Stripe PaymentIntent kayıtları | 0001 (genişletildi 0005) |

### Operasyon

| Tablo | Amaç | Migration |
| --- | --- | --- |
| `task_jobs` | Asynq job audit trail | 0003 |
| `system_events` | Operasyonel loglar (info/warn/error) | 0001 |

## State Machines

### `tests.status`

```
pending ──► active ──► completed
   │           │
   │           └──► failed
   └──► cancelled
```

### `test_assignments.status`

```
pending ──► in_progress ──► completed
   │             │              │
   │             └──► failed   (review_id NULL ise skip)
   └──► skipped
```

### `testers.status`

```
warming ──► active ──► cooling ──► disabled
   ▲          │            │
   └──────────┘            │
   (3-day warm-up)         │
   └─── tekrar warming'e dönmek için admin override
```

### `orders.status`

```
pending ──► awaiting_payment ──► paid
   │              │                 │
   │              └──► failed      refunded
   └──► cancelled
```

## Index Stratejisi

- `activity_logs`: `(test_id, performed_at DESC)` (brief'e uygun dashboard timeline)
- `tests`: `(user_id, created_at DESC)` (müşteri dashboard)
- `orders`: `(status, expires_at)` (stale order cleanup)
- `payments`: `(status, created_at DESC)` (admin panel)

## İlgili

- [[architecture]]
- [[services/api]] — `internal/repository/` SQL'leri
- `apps/api/migrations/0001_init.sql` — İlk şema
- `apps/api/migrations/0004_orders_plans.sql` — Plan/orders/sessions
- `apps/api/migrations/0005_stripe.sql` — Stripe alanları (yakında)
