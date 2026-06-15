---
tags: [service, web, nextjs, frontend]
---

# Web (Next.js)

> Müşteri + Admin dashboard. VPS'te çalışır. 18 route.

## Sorumluluk

- Landing (public)
- Auth: login, register
- Customer dashboard: tests list, test detail, new order, payment, settings
- Admin panel: overview, orders, tests, testers, payments
- Legal: terms, privacy, refund

## Stack

- Next.js 14.2.15 (App Router, standalone output)
- React 18.3.1
- TypeScript 5.6
- Tailwind CSS 3.4 (Brief-driven tokens)
- React Hook Form + Zod (validation)
- TanStack Query (data fetching)
- Zustand (client state, planned)
- Lucide React (icons)

## Design System

`brief.md` (root) tek kaynak doğruluğu:
- Cobalt accent (`hsl(222 89% 48%)`)
- Status palette (success/warning/info/destructive)
- Warm slate neutral
- Dark mode first-class
- WCAG 2.1 AA + reduced-motion

Component library (`apps/web/components/ui/`):
- `button`, `card`, `badge`, `input`, `label` (shadcn-style)
- `status-badge` (10 status, color+icon+label)
- `progress` (a11y, aria-valuenow)
- `empty-state` (icon + title + desc + CTA)
- `package-name` (mono, RTL truncate)
- `money` (₺ TRY, tabular-nums)
- `site-header` (marketing/app variant)
- `site-footer`

## Routes (18)

| Path | Type | Lane | Amaç |
| --- | --- | --- | --- |
| `/` | static | Decide | Landing |
| `/login` | static | Learn+Decide | Auth |
| `/register` | static | Learn+Decide | Auth |
| `/dashboard` | dynamic | Monitor | Test listesi |
| `/dashboard/new` | dynamic | Configure+Decide | Yeni test |
| `/dashboard/[testId]` | dynamic | Monitor | Test detay |
| `/dashboard/orders/[id]/pay` | dynamic | — | Stripe redirect |
| `/dashboard/orders/[id]/success` | dynamic | — | Ödeme onayı |
| `/dashboard/settings` | dynamic | — | Profil |
| `/admin` | dynamic | Operate+Compare | Overview |
| `/admin/orders` | dynamic | Compare | Tablo |
| `/admin/tests` | dynamic | Compare | Tablo |
| `/admin/testers` | dynamic | Operate | Tablo |
| `/admin/payments` | dynamic | Compare | Tablo |
| `/legal/terms` | static | Learn | |
| `/legal/privacy` | static | Learn | |
| `/legal/refund` | static | Learn | |

## API Client

`apps/web/lib/api.ts` — type-safe:

```ts
api.plans()                               // GET /api/v1/plans
api.tests(token)                          // GET /api/v1/tests
api.test(id, token)                       // GET /api/v1/tests/:id
api.testActivity(id, token)               // GET /api/v1/tests/:id/activity
api.testReviews(id, token)                // GET /api/v1/tests/:id/reviews
api.orders(token)                         // GET /api/v1/orders
api.createOrder(data, token)              // POST /api/v1/orders
api.order(id, token)                      // GET /api/v1/orders/:id
```

## Auth Server-Side

`lib/auth-server.ts`:
- `getCurrentUser()` — cookie'den access_token alır, API'ye `GET /auth/me` çağrısı
- Server Component'lerde redirect için

## Form Pattern

Login/Register:
- `useForm({ resolver: zodResolver(schema) })`
- Suspense wrap (useSearchParams için)
- Inline error messages (field-level)
- Submit button loading state

## Build

```
pnpm build
  → .next/standalone
  → Docker: COPY --from=builder .next/standalone ./
  → CMD ["node", "server.js"]
```

**First load JS**: 87KB (shared), 87-128KB per route

## İlgili

- [[architecture]]
- `brief.md` (root)
- `apps/web/lib/api.ts`
- [[services/api]] — Kontrat
- [[deployment]] — VPS deploy
