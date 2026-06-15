---
tags: [adr, design, brief, constitution]
status: accepted
date: 2026-06-15
---

# ADR-0004: brief.md as design constitution

## Context

Frontend tasarım kararları her sayfa, her component için tekrarlanıyor:

- Renk paleti (primary, status, neutral)
- Tipografi (sans family, weight, size)
- Spacing (4px base, card radius)
- Composition lanes (Monitor, Operate, Compare, Configure, Learn, Decide)
- Accessibility (WCAG AA, reduced-motion, focus rings)
- Anti-references (generic SaaS, spam, gray-market)

Bu kararlar `apps/web/components/ui/` içinde parçalanmış olmadan **tek dosyada, tek otorite** olmalı.

## Decision

Repo root'unda `brief.md` dosyası oluşturulacak. `/design` skill'i ve gelecek tüm design işleri bu dosyayı okuyacak.

## Structure

`brief.md` 12 bölüm:
1. Register (product vs brand)
2. Users and context
3. Product purpose
4. Voice (direct, technical, dry)
5. Anti-references (generic SaaS, spam, gray-market)
6. Design principles (6 kural)
7. Accessibility expectations (WCAG 2.1 AA, dark mode)
8. Visual foundation (cobalt, warm slate, 4px base)
9. Composition lanes (5 lane tanımı)
10. Component rules (button, badge, table, card, empty, error)
11. Domain-specific notes (package names, money, dates, ToS)
12. (token economics)

## Consequences

### Olumlu

- Single source of truth
- Yeni sayfa eklerken karar tekrarı yok
- `/design setup` her zaman aynı temele döner
- Taste ile uyumlu (taste.md ↔ brief.md cross-link)
- LLM'ler için kısa, yapılandırılmış (maks 200 satır)

### Olumsuz

- Brief güncellenmezse eski kararlar geçerli kalır
- Brief'i ignore eden biri tutarsız UI yapar

### Mitigations

- PR review'da brief uyumu kontrol
- `tailwind.config.ts` ve `globals.css` brief'ten derive edildi
- Her component brief referansı ile başlar (code comment)

## Implementation Notes

- `brief.md` (root) — 103 satır
- `apps/web/tailwind.config.ts` — Brief'teki color/font/radius
- `apps/web/app/globals.css` — CSS variables (light + dark)
- `apps/web/components/ui/*` — Component rules uygulanmış
- `apps/web/app/page.tsx` — Decide lane (landing)
- `apps/web/app/dashboard/*` — Monitor lane
- `apps/web/app/admin/*` — Operate+Compare lane
- `apps/web/app/(auth)/*` — Learn+Decide lane
- `apps/web/app/dashboard/new` — Configure+Decide lane

## References

- `/design setup` skill: `.nvm/.../command-code/skills/design/SKILL.md`
- `brief.md` (root)
- `.commandcode/taste/taste.md` (cross-link)

## Status

**Accepted** — 2026-06-15. Brief yazıldı, design tokens uygulandı, 18 route brief uyumlu.
