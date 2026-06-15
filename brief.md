# Brief — TestersCommunity

> Design constitution. Single source of truth for visual + interaction decisions. Owned by the design system; read before any `/design` invocation.

---

## Register

**Product.** This is a paid service tool, not a marketing site. The audience lands on `/` for context, but the value lives inside authenticated product surfaces: customer dashboards watching live tests, an admin panel orchestrating emulator farms, and order/checkout flows that take real money.

Marketing copy is allowed to feel like a confident product. It must never feel like a SaaS brochure, a gray-market reseller, or a black-hat utility.

## Users and context

**Primary user: indie Android developer** running a Google Play closed test. They have a single APK, a tight budget, and 14 days of patience. They log in 2–3 times per day from a phone, glance at status, and want assurance their money is doing something.

**Secondary user: platform admin** (the operator). They watch 25 active emulators, balance load, triage failed reviews, and rotate banned accounts. They live in the admin panel for hours at a time on a real monitor.

**Tertiary: the customer checking in once.** Comes from a Google search, sees pricing, decides. Bounces if the page doesn't answer "is this legal and does it work" inside 5 seconds.

## Product purpose

Generate 14 days of organic-looking engagement for one Android package, across 25 managed Google accounts, end-to-end automated. Customer gets: live activity, screenshots, reviews, peace of mind. Operator gets: a fleet control surface, not a firefighting console.

## Voice

**Direct, technical, slightly dry.** Turkish, second-person, no exclamation marks. Says what is happening without selling the happening.

Examples of the register:
- "25 hesap, 14 gün, otomatik." not "HIZLI VE GÜÇLÜ!"
- "Opt-in tamamlandı, indirme başladı." not "Müthiş bir başarı!"
- "Yorum yazma gün 14'te tetiklenir." not "Süpriz bonus! 🎉"

The product is doing something the customer is nervous about (Play Store closed testing). Calm beats enthusiastic.

## Anti-references

- **Generic SaaS landing pages** — three-card features, big centered hero, pill buttons, "trusted by 10,000+ teams" badges. We are not a startup pitch deck. We are a tool with a single job.
- **Spam/bulk-reseller sites** — neon green "INSTANT DOWNLOAD", fake review counts, countdown timers, "limited offer". Ucuz, güven düşürücü.
- **Gray-market / hack-tool aesthetic** — terminal green-on-black, skull icons, edgy copy. Attracts the wrong customer and exposes the operator.

## Design principles

1. **Status over decoration.** Every authenticated surface must show live state within 2 seconds of load. Numbers, timestamps, status pills. No static "welcome" cards.
2. **One job per screen.** Customer dashboard = tests + activity. Admin = fleet + orders. New order = package + payment. No "and also here are blog posts".
3. **Numbers carry the trust.** A 25/25 progress bar beats any testimonial. Show counts, not opinions.
4. **Calm under uncertainty.** When something fails (banned account, payment error), the UI explains what happened and what is being done. Never red flashing alerts. Never vague "an error occurred".
5. **Desktop-first for admin, mobile-first for dashboard.** Admin is a fleet operator. Customer checks from a phone.
6. **Money flows in Turkish Lira, dates in Europe/Istanbul, language in TR.** This is not a global product yet. Do not optimize for hypothetical international customers.

## Accessibility expectations

- WCAG 2.1 AA minimum, AA-strong preferred.
- Color contrast ≥ 4.5:1 for text, ≥ 3:1 for UI elements and status pills.
- Status communicated by more than color: icons + text label for every state ("Aktif", "Beklemede", "Başarısız").
- `prefers-reduced-motion: reduce` honored by default. No auto-playing animations. Page transitions ≤ 200ms. Status updates fade in, never slide.
- Full keyboard navigation. Focus rings visible against both light and dark backgrounds.
- Screen reader semantics for all status badges, progress bars (with `aria-valuenow`), and live regions for activity timeline.
- Dark mode is a first-class citizen. Not a theme toggle afterthought. Customer will check at night from bed.

## Visual foundation

**Color system:** Two-state (light/dark) with a single accent for action. Status colors follow a fixed semantic palette that survives both modes without losing meaning.

- **Primary accent:** saturated cobalt blue. Trust + technical. Not Stripe-purple, not Vercel-pink, not Tailwind-blue-500. Distinct.
- **Status palette:** green (active), amber (pending/warning), red (failed), slate (idle/disabled). All four must pass AA contrast in both modes.
- **Neutral scale:** warm slate (not cool gray). Slight blue undertone to harmonize with the accent.
- **No gradient backgrounds on landing or product.** Flat fills. The only gradient allowed: subtle 2-stop on the primary CTA, used once.

**Typography:** One sans-serif for everything. System font stack acceptable, but pair with a humanist sans (Inter, Geist Sans, or similar) as the webfont fallback. No second font for "elegance". One weight axis is wasted energy.

- Display: 600 weight, tight tracking (-0.02em).
- Body: 400 weight, 1.6 line height for reading, 1.4 for UI labels.
- Mono: for IDs, package names, ASIN-style codes. Same family, monospaced variant.

**Spacing & shape:** 4px base. Generous vertical rhythm in product surfaces (status boards breathe), tight in admin (density is a feature). Radii: 6px for inputs/buttons, 10px for cards, 14px for hero panels. No fully rounded pills except status badges (which use full rounding because their job is readability at a glance).

**Iconography:** Lucide (already in use). Stroke 1.75. Never filled. Color comes from current text, never a hardcoded hue.

## Composition lanes (what each surface IS)

- **Customer dashboard** → **Monitor.** Status boards, live activity timeline, progress bars, screenshot grid. The customer is watching a system work. Density is medium, freshness is high.
- **Admin panel** → **Operate + Compare.** Fleet table (sortable, filterable), order queue, account health matrix. Right inspector panel for selected entity. Density is high; power-user territory.
- **New order form** → **Configure + Decide.** Plan tier selector (3 cards, one highlighted), billing form, order summary sidebar. One dominant action ("Ödemeyi Tamamla"), sticky to the right on desktop.
- **Auth pages (login/register)** → **Learn + Decide.** Minimal. Logo, one card, one button. Don't distract. Don't cross-sell.
- **Landing page** → **Decide + Reassure.** Hero with one number ("25 hesap, 14 gün"), three-step "how it works" (not a feature grid), pricing block, FAQ. The FAQ answers the legal nerves. The legal warning bar is visible but not aggressive.

## Component rules

- **Buttons:** Primary is the only filled button. Secondary is outlined. Destructive uses a specific red, never the primary blue. Disabled state has 50% opacity, not a separate color.
- **Status badges:** Small, rounded-full, paired with an icon. Background 10% opacity of the status color, foreground the full status color. Text label always present.
- **Tables (admin):** Sticky header, row hover state, sortable columns indicated by arrow icon. Right-aligned for numbers. 8px row padding minimum; don't squish.
- **Cards:** 10px radius, 1px border, no shadow in light mode. Subtle shadow on dark mode.
- **Empty states:** Always include a single next action. "Henüz test yok → Yeni test oluştur". Never just an icon and silence.
- **Errors:** Inline, specific, recoverable. "Bu email zaten kayıtlı. Şifreni mi unuttun?" not "Error 409".

## Domain-specific notes

- **Package names** (e.g. `com.example.app`) are first-class text. Render in mono font. Truncate with ellipsis from the middle, not the end, so the TLD stays visible.
- **Test status flow:** `pending → active → completed` (or `failed`, `cancelled`). Always show the previous status as a small subtitle when relevant.
- **Money:** Always show currency symbol (₺). Two decimals. Right-aligned in tables. Don't say "TL", say "₺".
- **Dates:** Europe/Istanbul timezone. Format: `15 Haz 2026, 14:32`. Always show the year in archive views.
- **Tone around "testing":** legal copy lives in `/legal/*` and the bottom-of-page warning. Product copy can use the word "test" freely. Hero is honest about the ToS gray area in one sentence, not in scare quotes.
