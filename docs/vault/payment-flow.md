---
tags: [payment, stripe, integration]
---

# Stripe Ödeme Akışı

> Global ödeme altyapısı. Iyzico yerine Stripe kullanılıyor (bkz. [[decisions/0001-stripe-over-iyzico]]).

## API Seçimi

- **Checkout Sessions** (`checkout.sessions.create`) — hosted page
- `payment_method_types` ASLA parametre olarak gönderilmez (dynamic payment methods)
- Webhook signature doğrulaması zorunlu (`whsec_...` secret ile)

## Akış

```
[Müşteri] ─► /dashboard/new
   │
   ▼
[Web] POST /api/v1/orders {plan_slug, package_name, test_link}
   │
   ▼
[API] OrderService.Create()
   │  ├─ orders tablosuna 'pending' insert
   │  └─ StripeService.CreateCheckoutSession()
   │       └─ stripe.checkout.sessions.create({
   │             mode: 'payment',
   │             line_items: [{ price_data: ... }],
   │             success_url, cancel_url,
   │             metadata: { order_id, user_id, package_name }
   │          })
   │
   ▼
[API] ─► { order_id, payment_url, expires_at }
   │
   ▼
[Web] window.location = payment_url
   │
   ▼
[Stripe Hosted] Müşteri kart girer → 3D Secure
   │
   ▼
[Stripe Webhook] ─► POST /api/v1/payments/stripe/webhook
   │  Header: Stripe-Signature
   │  Event: checkout.session.completed | payment_intent.succeeded | payment_intent.payment_failed
   │
   ▼
[API] WebhookHandler
   │  ├─ stripe.webhooks.ConstructEvent(payload, sig, webhookSecret)
   │  ├─ event.type == 'checkout.session.completed' →
   │  │   order.MarkPaid() + payments tablosuna insert
   │  │   Asynq: test_start job enqueue
   │  └─ 200 OK
   │
   ▼
[Web] /dashboard/orders/{id}/success
```

## API Endpoints

| Method | Path | Auth | Amaç |
| --- | --- | --- | --- |
| `POST` | `/api/v1/orders` | JWT | Yeni sipariş, Stripe session |
| `GET` | `/api/v1/orders` | JWT | Kullanıcının siparişleri |
| `GET` | `/api/v1/orders/:id` | JWT | Sipariş detayı |
| `POST` | `/api/v1/payments/stripe/webhook` | Stripe-Sig | Webhook alıcı |

## Env Değişkenleri

```
STRIPE_SECRET_KEY=sk_live_...         # Production
STRIPE_SECRET_KEY=sk_test_...         # Sandbox (stripe sandbox create)
STRIPE_PUBLISHABLE_KEY=pk_...
STRIPE_WEBHOOK_SECRET=whsec_...       # Dashboard > Developers > Webhooks
```

## Restricted API Key (RAK)

Production'da `sk_` yerine `rk_` kullanılır. Permissions:
- `checkout.sessions: write`
- `checkout.sessions: read`
- `payment_intents: read`
- `webhooks: read`

## Test

Stripe CLI ile sandbox keys:
```bash
brew install stripe/stripe-cli/stripe
stripe login
stripe sandbox create              # 7 günlük claimable sandbox
stripe sandbox claim               # kalıcı yap
stripe listen --forward-to localhost:8080/api/v1/payments/stripe/webhook
```

Stripe olmadan lokal test (dev stub):
- `STRIPE_SECRET_KEY` boşsa `stripeClient.IsConfigured()` false döner
- Fake session ID ve URL üretir, order `pending` kalır
- Web tarafı `localhost:3000/dashboard/orders/{id}/success`e direkt yönlenir

## İlgili

- [[services/api]] — `internal/service/stripe_client.go`, `internal/handler/order.go`
- [[decisions/0001-stripe-over-iyzico]]
- `.commandcode/skills/stripe-best-practices/SKILL.md` — Best practices
