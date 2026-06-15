// Tek marka kaynağı. Yeni isim değişikliği sadece burada.
// Domain ve diğer marka bilgileri buradan okunur.

export const BRAND = {
  name: 'PlayForge',
  tagline: 'Android app beta testing, 25 real testers in 14 days',
  shortDescription: '14-day closed beta testing service for Android apps on Google Play',
  domain: process.env.NEXT_PUBLIC_DOMAIN ?? 'playforge.app',
  // Tüm subdomainler aynı zone'da
  api: `https://api.${process.env.NEXT_PUBLIC_DOMAIN ?? 'playforge.app'}`,
  web: `https://${process.env.NEXT_PUBLIC_DOMAIN ?? 'playforge.app'}`,
  queue: `https://queue.${process.env.NEXT_PUBLIC_DOMAIN ?? 'playforge.app'}`,
  email: {
    support: 'support@playforge.app',
    legal: 'legal@playforge.app',
    privacy: 'privacy@playforge.app',
    refund: 'refund@playforge.app',
    noReply: 'no-reply@playforge.app',
  },
  social: {
    twitter: 'https://twitter.com/playforge',
    github: 'https://github.com/playforge',
  },
  legalName: 'PlayForge',
  legalAddress: '—',
  founded: 2026,
} as const

export const SUPPORTED_LOCALES = ['tr', 'en', 'es', 'de', 'fr', 'pt'] as const
export type Locale = (typeof SUPPORTED_LOCALES)[number]
export const DEFAULT_LOCALE: Locale = 'en'

export const LOCALE_LABELS: Record<Locale, string> = {
  tr: 'Türkçe',
  en: 'English',
  es: 'Español',
  de: 'Deutsch',
  fr: 'Français',
  pt: 'Português',
}

// Locale'e göre para birimi ve Intl locale
export const LOCALE_CONFIG: Record<Locale, { currency: 'TRY' | 'USD' | 'EUR' | 'BRL'; intl: string; country: string }> = {
  tr: { currency: 'TRY', intl: 'tr-TR', country: 'TR' },
  en: { currency: 'USD', intl: 'en-US', country: 'US' },
  es: { currency: 'EUR', intl: 'es-ES', country: 'ES' },
  de: { currency: 'EUR', intl: 'de-DE', country: 'DE' },
  fr: { currency: 'EUR', intl: 'fr-FR', country: 'FR' },
  pt: { currency: 'BRL', intl: 'pt-BR', country: 'BR' },
}
