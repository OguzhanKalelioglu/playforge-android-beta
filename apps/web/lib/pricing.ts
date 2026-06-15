// Pricing config — pazar eşdeğeri (PrimeTestLab karşılaştırması)
//
// Pazar benchmark (Mayıs 2026):
//   PrimeTestLab: $14.99 (12), $24.99 (20), $19.99 (25) | $40 (25 premium)
//   BetaTesting.com: enterprise tier
//   Prefinery: $399/mo+ subscription
//
// PlayForge plan yapısı: 3 tier × 3 tester sayısı
//   - Basic: minimum başlangıç (Google 12 eşiği)
//   - Pro: en popüler (yeterli güvenlik)
//   - Enterprise: 25 tester + extras (best value)

import { type Locale } from './brand'

export type PlanKey = 'basic' | 'pro' | 'enterprise'

export interface PlanConfig {
  slug: PlanKey
  name: string
  testerCount: 12 | 20 | 25
  durationDays: 14
  // Fiyatlar USD cinsinden (Stripe Checkout bu currency'de)
//   Tüm locale'lerde aynı USD fiyat gösterilir (vergili/komisyonlu, faturalandırma netliği için)
  priceUSD: number
  highlighted: boolean
  order: number
}

export const PLANS: PlanConfig[] = [
  {
    slug: 'basic',
    name: 'Basic',
    testerCount: 12,
    durationDays: 14,
    priceUSD: 19.99,
    highlighted: false,
    order: 1,
  },
  {
    slug: 'pro',
    name: 'Pro',
    testerCount: 20,
    durationDays: 14,
    priceUSD: 29.99,
    highlighted: true,
    order: 2,
  },
  {
    slug: 'enterprise',
    name: 'Enterprise',
    testerCount: 25,
    durationDays: 14,
    priceUSD: 49.99,
    highlighted: false,
    order: 3,
  },
]

// Locale'e göre ödeme para birimi (Stripe Checkout currency)
// USD globalde primary, TR pazarı için TL opsiyonel.
// Mevcut: sadece USD. TR/BRL için ayrı Stripe product gerekebilir.
export function getCurrencyForLocale(_locale: Locale): 'USD' {
  return 'USD'
}

export function getIntlLocale(locale: Locale): string {
  const map: Record<Locale, string> = {
    tr: 'tr-TR',
    en: 'en-US',
    es: 'es-ES',
    de: 'de-DE',
    fr: 'fr-FR',
    pt: 'pt-BR',
  }
  return map[locale]
}
