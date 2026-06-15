import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// ============================================
// Money — TR ₺ always, 2 decimals
// ============================================
export function formatCurrency(amount: number, currency = 'TRY'): string {
  return new Intl.NumberFormat('tr-TR', {
    style: 'currency',
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(amount)
}

export function formatMoneyShort(amount: number, currency = 'TRY'): string {
  if (amount >= 1000) {
    return `${(amount / 1000).toFixed(1).replace('.0', '')}K ₺`
  }
  return `${amount.toFixed(0)} ₺`
}

// ============================================
// Dates — Europe/Istanbul, "15 Haz 2026, 14:32"
// ============================================
const TR_LOCALE = 'tr-TR'
const TR_TIMEZONE = 'Europe/Istanbul'

export function formatDate(date: string | Date): string {
  return new Intl.DateTimeFormat(TR_LOCALE, {
    day: '2-digit',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    timeZone: TR_TIMEZONE,
  }).format(new Date(date))
}

export function formatDateShort(date: string | Date): string {
  return new Intl.DateTimeFormat(TR_LOCALE, {
    day: '2-digit',
    month: 'short',
    timeZone: TR_TIMEZONE,
  }).format(new Date(date))
}

export function formatTime(date: string | Date): string {
  return new Intl.DateTimeFormat(TR_LOCALE, {
    hour: '2-digit',
    minute: '2-digit',
    timeZone: TR_TIMEZONE,
  }).format(new Date(date))
}

export function relativeTime(date: string | Date): string {
  const d = new Date(date)
  const diffMs = d.getTime() - Date.now()
  const diffMin = Math.round(diffMs / 60000)
  const rtf = new Intl.RelativeTimeFormat(TR_LOCALE, { numeric: 'auto' })
  if (Math.abs(diffMin) < 60) return rtf.format(diffMin, 'minute')
  const diffH = Math.round(diffMin / 60)
  if (Math.abs(diffH) < 24) return rtf.format(diffH, 'hour')
  const diffD = Math.round(diffH / 24)
  return rtf.format(diffD, 'day')
}

// ============================================
// Package names — mono font, RTL truncate (keep TLD)
// ============================================
export function formatPackageName(pkg: string, max = 28): string {
  if (!pkg) return '—'
  if (pkg.length <= max) return pkg
  // Keep ".com" / ".net" / ".io" tail
  const lastDot = pkg.lastIndexOf('.')
  if (lastDot === -1) {
    return pkg.slice(0, max - 1) + '…'
  }
  const tld = pkg.slice(lastDot)
  const head = pkg.slice(0, max - tld.length - 1)
  return head + '…' + tld
}

// ============================================
// Email
// ============================================
export function maskEmail(email: string): string {
  const [local, domain] = email.split('@')
  if (!local || !domain) return email
  const visible = Math.min(2, Math.floor(local.length / 3))
  return `${local.slice(0, visible)}${'*'.repeat(Math.max(3, local.length - visible))}@${domain}`
}

// ============================================
// Percent
// ============================================
export function formatPercent(value: number, total: number): number {
  if (total === 0) return 0
  return Math.round((value / total) * 100)
}

// ============================================
// Duration (seconds → "2dk 35s")
// ============================================
export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  if (m < 60) return s > 0 ? `${m}dk ${s}s` : `${m}dk`
  const h = Math.floor(m / 60)
  const rm = m % 60
  return rm > 0 ? `${h}sa ${rm}dk` : `${h}sa`
}
