'use client'

import { usePathname, useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { useTransition } from 'react'
import { Globe } from 'lucide-react'
import { LOCALE_LABELS, SUPPORTED_LOCALES, type Locale } from '@/lib/brand'
import { cn } from '@/lib/utils'

interface LocaleSwitcherProps {
  currentLocale: Locale
  mobile?: boolean
}

export function LocaleSwitcher({ currentLocale, mobile }: LocaleSwitcherProps) {
  const pathname = usePathname()
  const router = useRouter()
  const t = useTranslations('common')
  const [isPending, startTransition] = useTransition()

  const switchTo = (next: Locale) => {
    if (next === currentLocale) return
    // pathname: /tr/dashboard, /en/...
    const segments = pathname.split('/').filter(Boolean)
    if (SUPPORTED_LOCALES.includes(segments[0] as Locale)) {
      segments[0] = next
    } else {
      segments.unshift(next)
    }
    const newPath = '/' + segments.join('/')
    startTransition(() => {
      router.push(newPath)
    })
  }

  if (mobile) {
    return (
      <div className="flex flex-wrap gap-2">
        {SUPPORTED_LOCALES.map((loc) => (
          <button
            key={loc}
            onClick={() => switchTo(loc)}
            disabled={isPending}
            className={cn(
              'rounded-md border px-3 py-1.5 text-xs font-medium transition-colors',
              loc === currentLocale
                ? 'border-primary bg-primary text-primary-foreground'
                : 'border-input bg-background hover:bg-muted'
            )}
          >
            {LOCALE_LABELS[loc]}
          </button>
        ))}
      </div>
    )
  }

  return (
    <div className="relative">
      <label className="sr-only" htmlFor="locale-select">
        {t('language')}
      </label>
      <div className="flex items-center gap-1 rounded-md border bg-background px-2 py-1 text-sm">
        <Globe className="h-3.5 w-3.5 text-muted-foreground" aria-hidden />
        <select
          id="locale-select"
          value={currentLocale}
          onChange={(e) => switchTo(e.target.value as Locale)}
          disabled={isPending}
          className="cursor-pointer appearance-none bg-transparent pr-1 text-xs font-medium focus:outline-none"
        >
          {SUPPORTED_LOCALES.map((loc) => (
            <option key={loc} value={loc}>
              {LOCALE_LABELS[loc]}
            </option>
          ))}
        </select>
      </div>
    </div>
  )
}
