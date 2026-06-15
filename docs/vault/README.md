---
tags: [moc, index, project]
---

# TestersCommunity — Project Vault

> Bu vault, LLM'ler ve insanlar için projenin tek kaynak doğruluğudur (single source of truth). Her yeni özellik burada wikilink ile bağlanır, her karar `decisions/` altında ADR olarak saklanır.

## Quick Links

- [[architecture|Mimari Genel Bakış]] — Servisler, veri akışı, dağıtım topolojisi
- [[services/api|API Service]] — Gin REST API (auth, orders, tests, admin)
- [[services/orchestrator|Orchestrator]] — Emulator farm + Appium otomasyon
- [[services/worker|Worker]] — Asynq job consumer
- [[services/scheduler|Scheduler]] — 14-günlük planlayıcı
- [[services/web|Web (Next.js)]] — Müşteri + Admin dashboard
- [[database-schema|PostgreSQL Şeması]]
- [[payment-flow|Stripe Ödeme Akışı]]
- [[task-runner|Task Runner Pipeline]]
- [[deployment|VPS + Mini PC Deploy]]

## Code Graph

- [[code-graph/services|Services Index]] — Her servisin dosya haritası
- [[code-graph/endpoints|HTTP Endpoints]] — Tüm route'lar
- [[code-graph/db-tables|DB Tables]] — Tablolar ve ilişkiler
- [[code-graph/jobs|Asynq Jobs]] — Background job kataloğu

## Decisions (ADRs)

- [[decisions/0001-stripe-over-iyzico|ADR-0001: Stripe over Iyzico]]
- [[decisions/0002-mini-pc-25-emulators|ADR-0002: Mini PC for 25 emulators]]
- [[decisions/0003-appium-orchestrator|ADR-0003: Appium for UI automation]]
- [[decisions/0004-brief-design|ADR-0004: brief.md design constitution]]

## Daily Notes

`daily/` klasöründe günlük ilerleme notları. Yeni iş gününde `YYYY-MM-DD.md` oluşturulur.

## Status

| Servis | Build | Deploy | Health |
| --- | --- | --- | --- |
| [[services/api\|API]] | ✅ | ⏳ VPS bekleniyor | — |
| [[services/orchestrator\|Orchestrator]] | ✅ | ⏳ Mini PC IP | — |
| [[services/worker\|Worker]] | ✅ | ⏳ VPS | — |
| [[services/scheduler\|Scheduler]] | ✅ | ⏳ VPS | — |
| [[services/web\|Web]] | ✅ (18 route) | ⏳ VPS | — |
