---
tags: [code-graph, database, postgres]
---

# DB Tables

PostgreSQL şemasının repository katmanı ile birlikte kataloğu. Her tablo için: kolonlar, index'ler, repository method'ları.

## Users & Auth

### `users` — 0001

| Kolon | Tip | Not |
| --- | --- | --- |
| id | UUID PK | |
| email | VARCHAR UNIQUE | Login key |
| password_hash | VARCHAR | bcrypt |
| name | VARCHAR | |
| role | VARCHAR | 'customer' \| 'admin' |
| email_verified | BOOLEAN | |
| created_at, updated_at | TIMESTAMPTZ | trigger |

**Repository**: `apps/api/internal/repository/user.go` → `UserRepository`
- `Create(ctx, *User)`
- `GetByEmail(ctx, email) (*User, error)`
- `GetByID(ctx, uuid) (*User, error)`

### `sessions` — 0004

Refresh token storage. SHA256 hash + expiry + IP/UA.

| Kolon | Tip | Not |
| --- | --- | --- |
| refresh_token_hash | VARCHAR UNIQUE | SHA256(token) |
| user_agent, ip_address | TEXT, INET | Audit |
| expires_at | TIMESTAMPTZ | 30 gün |
| revoked_at | TIMESTAMPTZ NULL | Soft-revoke |

**Repository**: `repository/session.go` → `SessionRepository`

### `password_reset_tokens`, `email_verification_tokens` — 0004

Tek-use tokenler. `token_hash` SHA256, `expires_at`, `used_at` NULL.

## Test Domain

### `tests` — 0001

| Kolon | Tip | Not |
| --- | --- | --- |
| user_id | UUID FK → users | Owner |
| package_name | VARCHAR | com.example.app |
| test_link | TEXT | Play Store closed test |
| status | VARCHAR | pending\|active\|completed\|failed\|cancelled |
| star_preference | VARCHAR | all5\|mixed\|custom |
| google_group_id | UUID FK | |
| started_at, ends_at | TIMESTAMPTZ | |

**Repository**: `repository/test.go` → `TestRepository`
- `Create`, `GetByID`, `ListByUser`, `ListAll(status, limit)`

### `test_assignments` — 0001

Tester × test eşleme.

| Kolon | Tip | Not |
| --- | --- | --- |
| test_id, tester_id | UUID FK | UNIQUE(test_id, tester_id) |
| status | VARCHAR | pending\|in_progress\|completed\|failed\|skipped |
| opt_in_at, install_at, last_engagement_at | TIMESTAMPTZ | State markers |
| review_id | UUID FK → reviews | |
| error_message | TEXT | Last error |

**Repository**: `TestRepository.ListAssignments(ctx, testID)`

### `activity_logs` — 0001

Orchestrator'dan step event'leri.

| Kolon | Tip | Not |
| --- | --- | --- |
| test_assignment_id | UUID FK | |
| action | VARCHAR | opt_in\|download\|install\|open\|interact\|review\|error |
| performed_at | TIMESTAMPTZ | |
| success | BOOLEAN | |
| error_message, screenshot_path | TEXT | |
| metadata | JSONB | gesture_count, app_duration, vs. |

**Indexes**: `(test_assignment_id, performed_at DESC)`, `(test_id, performed_at DESC)` (partial 0002), GIN on metadata

**Repository**: `TestRepository.ListActivity(ctx, testID, limit)`

### `reviews` — 0001

| Kolon | Tip | Not |
| --- | --- | --- |
| test_assignment_id | UUID FK | |
| rating | INT | 1-5 |
| review_text | TEXT | TR/EN/DE |
| language | VARCHAR | 'tr' default |
| posted_at | TIMESTAMPTZ | |

**Repository**: `TestRepository.ListReviews(ctx, testID)`

## Account Management

### `testers` — 0001

25 Google hesabı.

| Kolon | Tip | Not |
| --- | --- | --- |
| email | VARCHAR UNIQUE | |
| password_encrypted | BYTEA | AES-256-GCM |
| recovery_email, phone | VARCHAR | |
| google_group_id, device_profile_id | UUID FK | |
| status | VARCHAR | warming\|active\|cooling\|disabled |
| last_used_at | TIMESTAMPTZ | |
| notes | TEXT | Admin |

**Repository**: `repository/tester.go` → `TesterRepository`
- `ListAll`, `GetByID`, `UpdateStatus`, `MarkUsed`
- `ListAdmin` — 30d görev sayısı ile join

### `google_groups`, `device_profiles`, `tester_daily_usage` — 0001/0004

Gruplar (test başına 1 grup), cihaz profilleri, günlük rotasyon takibi.

## Plan & Ödeme

### `plan_tiers` — 0004

| Kolon | Tip | Not |
| --- | --- | --- |
| slug | VARCHAR UNIQUE | basic\|pro\|enterprise |
| name, description | VARCHAR, TEXT | |
| tester_count, duration_days | INT | |
| price_try, price_usd | DECIMAL | |
| features | JSONB | String array |
| is_active, sort_order | BOOLEAN, INT | |

**Default seed**: basic (₺4999), pro (₺7999, önerilen), enterprise (₺12999)

**Repository**: `repository/order.go` → `PlanRepository`
- `List`, `GetBySlug`, `GetByID`

### `orders` — 0004

| Kolon | Tip | Not |
| --- | --- | --- |
| user_id, plan_tier_id | UUID FK | |
| status | VARCHAR | pending\|awaiting_payment\|paid\|failed\|cancelled\|refunded |
| subtotal, tax_total, total | DECIMAL | TRY |
| stripe_checkout_session_id | VARCHAR | (0005) |
| stripe_payment_intent_id | VARCHAR | (0005) |
| stripe_customer_id | VARCHAR | (0005) |
| test_id | UUID FK | paid olduktan sonra set |
| billing_email, billing_name, billing_phone, billing_address | | Stripe checkout için |
| expires_at | TIMESTAMPTZ | 30dk default |
| paid_at, cancelled_at | TIMESTAMPTZ | |

**Repository**: `OrderRepository`
- `Create`, `GetByID`, `ListByUser`, `ListAll`
- `MarkPaid(ctx, id, testID)`, `SetStripeSession(ctx, id, sessionID, expires)`

### `payments` — 0001 (0005'te genişletilecek)

| Kolon | Tip | Not |
| --- | --- | --- |
| user_id, test_id | UUID FK | |
| amount, currency | DECIMAL, VARCHAR | |
| status | VARCHAR | pending\|completed\|refunded\|failed\|cancelled |
| stripe_payment_intent_id | VARCHAR | (0005) |
| stripe_session_id | VARCHAR | (0005) |
| stripe_charge_id | VARCHAR | (0005) |
| paid_at, refunded_at | TIMESTAMPTZ | |

**Repository**: `PaymentRepository`
- `Create`, `GetByID`, `MarkCompleted(ctx, id, chargeID)`

## Operasyon

### `task_jobs` — 0003

Asynq job audit trail.

| Kolon | Tip | Not |
| --- | --- | --- |
| job_id | VARCHAR UNIQUE | Asynq job ID |
| test_id, assignment_id | UUID FK | |
| task_type | VARCHAR | test_start\|engage\|review\|login_google |
| day | INT | 0-14 |
| status | VARCHAR | pending\|running\|completed\|failed\|retrying\|dead |
| attempts, last_error | INT, TEXT | |
| payload_encrypted | BYTEA | |
| started_at, completed_at | TIMESTAMPTZ | |

### `system_events` — 0001

Operational logs (info/warning/error/critical). JSONB metadata, severity-filtered index.

## İlgili

- [[database-schema]] — Yüksek seviye
- [[services/api]] — Repository kod
- `apps/api/migrations/` — SQL dosyaları
- [[code-graph/services]]
