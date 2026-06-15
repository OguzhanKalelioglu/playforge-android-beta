---
tags: [adr, payment, stripe]
status: accepted
date: 2026-06-15
---

# ADR-0001: Stripe over Iyzico

## Context

Proje Türkiye pazarını hedefliyor, ancak tek bir ödeme sağlayıcısıyla sınırlandığımızda:

- **Müşteri tabanı sınırlı** — Iyzico yalnızca Türkiye'de faaliyet gösterir
- **Para birimi** — TRY dışında native ödeme zor
- **Subscription/recurring** — İleride önemli olabilir
- **Developer experience** — Stripe'ın SDK'ları, dökümanları, sandbox'ı üstün
- **Stripe skills** — Commander Coder CLI ile entegrasyon için AI skill'leri mevcut (best-practices, projects, directory, upgrade)

## Decision

Stripe Checkout Sessions API kullanılacak. Iyzico kaldırılacak.

## Consequences

### Olumlu

- Global ölçeklenebilirlik (US, EU, TR hepsi Stripe destekler)
- TRY, USD, EUR multi-currency doğal
- Webhook signature verification built-in
- Stripe CLI ile sandbox test (CLI ile `stripe sandbox create`)
- Stripe MCP server ile AI agent integration
- `restricted API keys` ile fine-grained permissions
- PCI compliance otomatik (hosted checkout)

### Olumlu (taste-aligned)

- `brief.md` rate 0.90: "Use Stripe (Global) instead of Iyzico"
- `brief.md` rate 0.75: "Use Commander Coder CLI with Stripe skills"

### Olumsuz

- Migration kodu (Iyzico client → Stripe client)
- Stripe webhook için HTTPS zorunlu (lokal'de ngrok / dev stub)
- İlk setup daha karmaşık (Stripe Dashboard, API keys)
- TR'de Stripe fee yüksek olabilir (2.4% + 0.30 TRY vs Iyzico 1.49%)

## Implementation Notes

- `internal/service/stripe_client.go` (yeni) — Stripe SDK v82+
- `internal/service/order.go` — `CreateCheckoutSession` + `RetrieveCheckoutSession`
- `internal/handler/order.go` — `StripeWebhook` handler, signature verification
- `apps/api/migrations/0005_stripe.sql` — `stripe_checkout_session_id`, `stripe_payment_intent_id`, `stripe_customer_id` kolonları
- `internal/service/payment_iyzico.go` — silinecek (veya archive)
- Env: `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET` (Iyzico env'leri kaldırılır)
- Dev stub: `STRIPE_SECRET_KEY` boşsa fake session ID üretir, gerçek API çağrısı yapmaz

## References

- https://docs.stripe.com/building-with-ai
- https://docs.stripe.com/payments/checkout
- https://docs.stripe.com/api/checkout/sessions
- https://docs.stripe.com/webhooks#verify-events
- `.commandcode/skills/stripe-best-practices/SKILL.md`
- `.commandcode/skills/stripe-best-practices/references/payments.md`
- `brief.md` (taste preferences)

## Status

**Accepted** — 2026-06-15. Migration devam ediyor, hedef: aynı hafta.
