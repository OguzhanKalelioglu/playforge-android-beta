---
tags: [adr, brand, rebrand, playforge]
status: accepted
date: 2026-06-15
---

# ADR-0005: Rebrand TestersCommunity → PlayForge

## Context

"TestersCommunity" zaten App Store'da var olan bir isim (örn. testerscommunity.net, testers.community). Yeni marka için:
- Aynı isimle global nişte çakışma riski
- SEO açısından "play" + "test" kelimeleri kritik
- "Forge" ile "inşa etme" metaforu (25 hesapla test "inşa" etme)

## Decision

`lib/brand.ts` (apps/web) tek marka kaynağı:
- **Marka adı:** PlayForge
- **Domain:** playforge.app (hedef, şu an testerscomm.net altında)
- **Logo:** "P" kare primary
- **Emailler:** support@playforge.app, legal@playforge.app, ...

## Implementation

- `lib/brand.ts` — BRAND, SUPPORTED_LOCALES, LOCALE_CONFIG
- Tüm page/layout/component brand'i `BRAND.name` üzerinden okur
- Email/domain hardcoded değil
- Domain rename: testerscomm.net → playforge.app (operasyonel, sonra)

## Consequences

### Olumlu

- Tek marka kaynağı
- Yeni isim çakışma riski sıfır
- "Play" + "Test" SEO'ya uygun
- Global ölçeklenebilirlik

### Olumsuz

- Marka değişikliği marketing materyali gerektirir (sonra)
- Go modül adı (`testerscommunity/api`) teknik debt — kullanıcı görmüyor

## Status

**Accepted** — 2026-06-15. Code tarafı tamamlandı, domain henüz yenilenmedi.
