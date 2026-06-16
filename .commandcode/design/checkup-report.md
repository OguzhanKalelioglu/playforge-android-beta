# Checkup Report — PlayForge Web

**Date:** 2026-06-16
**Mode:** checkup
**Scope:** PlayForge web — landing, auth, legal, dashboard redirects, 6 locales, mobile 375px + desktop 1280px

---

## Verdict

**35 / 60 — Watch.** Six vitals measured. Two healthy, three watch, one unverified. Interface is shippable, but the brand color collapses to "AI default" and the auth screens have a real accessibility hole (no semantic h1). No critical block, no broken task.

## TL;DR

- **Brand color reads generic.** `--primary: 222 89% 48%` is HSL, lands in the same neighborhood as Vercel/Tailwind blue-500. Brief explicitly calls for a cobalt with "not Stripe-purple, not Vercel-pink, not Tailwind-blue-500" distinctness. Needs OKLCH, slightly cooler, more chroma, distinct from the field.
- **Auth screens have no h1.** `/tr/register` and `/tr/login` render with `h1: None`. Headings use h2 inside cards. Screen reader, SEO, and the squint test all suffer. Fix in same pass.
- **Hero stat bar is timid.** `25 / 14 / ~18h` is the strongest single argument the product has, but it sits behind `border-y py-6` like a footnote. Brief: "A 25/25 progress bar beats any testimonial." Promote it.
- **Reduced motion and focus ring are already honored.** Good.
- **Speed unverified in dev mode.** Will be measured at next deploy, not now.

## Composition Vital Sign

| Surface | Pattern | Verdict |
|---|---|---|
| `/[locale]` landing | Decide + Reassure | Match. Hero → 3-step how → pricing → FAQ → legal warning → CTA. |
| `/[locale]/(auth)/{login,register}` | Learn + Decide | Match. Single card, one button. |
| `/[locale]/legal/terms` | Learn | Match. Long-form readable measure, mono on package names. |
| `/[locale]/dashboard` (gated) | Monitor | Verified via redirect to `/login?next=…`. Cannot judge rendered surface from current session. |
| `/[locale]/admin/*` | Operate + Compare | Out of scope for this pass (not authed). |

## Prompt Fidelity Vital Sign

- **Name:** "PlayForge" rendered, no TestersCommunity leak. ✓
- **Category:** Closed beta testing service for Android, visible in hero and pricing copy. ✓
- **Artifact:** "25 hesap, 14 gün, otomatik test." / "25 accounts, 14 days, automated testing." Exact brief phrasing. ✓
- **Proof:** 25/14/~18h stat row, plan tester count, 14-day engagement. ✓
- **Drift to refuse:** The 3-card feature grid at `/features` is the genre reflex brief warned about. It survives only because the section is small (6 cards in a 3-col grid) and the hero is honest. Marginal. Worth noting.
- **Hue check:** Primary is the first thing tested. Current cobalt is not distinct from the field. This is the most generic tell on the page.

## Vital Score Matrix

| Vital | Status | Score | Evidence |
|---|---|---|---|
| Intentionality | Watch | 5/10 | Theme system is in place, dark mode is first-class, but the cobalt hue blends with the field. |
| Readability | Healthy | 10/10 | `tracking-tight` on display, `tabular-nums` on money, `font-mono` on package names, dark mode 1:1 token parity. |
| Usability | Watch | 5/10 | Anchor fix landed. CTA clear. Auth pages lack h1, hurts scan path. |
| Responsiveness | Watch | 5/10 | 375px viewport renders cleanly, mobile menu present. Container queries and thumb-zone placement are not done. |
| Speed | Unverified | 5/10 | Dev server, no production build measured. Marked Watch by default. |
| Accessibility | Watch | 5/10 | Focus-visible ring, reduced-motion, semantic links. Missing: h1 on auth, focus order verification, color contrast in `bg-muted/30` cards. |

**Total: 35 / 60.**

## Priority Findings

### P1 — Brand color is in the field

`globals.css` line 13: `--primary: 222 89% 48%;`

This is the same neighborhood as `tailwindcss-blue-500` (`217 91% 60%` in dark, `221 83% 53%` in light). Brief said "cobalt blue, not Stripe-purple, not Vercel-pink, not Tailwind-blue-500" — currently indistinguishable at a glance.

Fix: Convert primary to OKLCH, push hue slightly cooler (`250` → `245`), raise chroma, drop lightness a touch in light mode to claim "cobalt" rather than "indigo/blue". Status colors follow a similar path. HSL stays in the token names, but values become OKLCH.

### P2 — Auth pages have no semantic h1

`/tr/register` and `/tr/login` report `h1: None`. Headings inside are h2 (or no heading at all). Screen reader announces the page by `<title>` only, no landmark. SEO misses the keyword. Squint test fails — first viewport reads as a form with no orientation.

Fix: Add an h1 to the auth card header in both `login/page.tsx` and `register/page.tsx`. The visual can stay close to current; the heading is the missing piece.

### P3 — Hero stat bar is not doing its job

The `25 / 14 / ~18h` row is the brief's "A 25/25 progress bar beats any testimonial." It is currently 3 text columns inside `border-y py-6`, the same visual weight as a footnote. It is the proof object, and it is being treated as decoration.

Fix: Give the stat row a panel surface (subtle background or inset border), larger numerical type, and a sentence that names the artifact in plain language.

## Watch-List (not blocking)

- 3-card feature grid in `/features` is the genre reflex. Not removed because brief gave the OK to "small features" — but the section is the most generic in the product.
- The legal warning bar uses `bg-warning/5` with no icon pairing, only the title prefix. The "show, don't tell" rule says: never red flashing alerts. Current is correct in restraint, but could carry a status pill instead of relying on color alone.
- Mobile thumb-zone: primary CTAs are at the bottom of the hero on mobile (correct), but the dashboard is untested. Logged for next pass.
- Speed will be measured in the next deploy pass (Lighthouse on production build, not dev).

## Prescriptions

| # | Issue | Mode to fix | Why now |
|---|---|---|---|
| 1 | Primary hue blends with field | `recolor` | One-shot system change, every page benefits. |
| 2 | Auth screens missing h1 | `surface` | 5 lines of code, accessibility win, ship blocker avoided. |
| 3 | Hero stat bar underweighted | `relayout` | Composition change, requires the recolor to land first. |
| 4 | Feature grid genre reflex | `deslop` | Optional follow-up; brief allowed small features. |

The first three are addressed in this pass. The fourth is logged for a future pass.

## Evidence

- Primary value read from `getComputedStyle(document.documentElement).getPropertyValue('--primary')` on `/tr`, `/en`, `/tr/register`, `/tr/login`, `/tr/legal/terms` — all reported `222 89% 48%`.
- h1 absence confirmed via Playwright `locator('h1').count()` on `/tr/register` and `/tr/login`.
- 31 interactive elements (buttons + links) on landing — many are nested locale switcher + legal links. Not a usability problem, but a refactor opportunity.
- 375px viewport rendered without overflow on hero, register, login.
- `prefers-reduced-motion: reduce` honored via `globals.css` line 145.
- Focus-visible ring styled via `globals.css` line 64.

## Artifacts

- `.commandcode/design/checkup-report.md` (this file)
- `.commandcode/design/checkup-report.html` (visual report, dark, structured)
