'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Menu, X } from 'lucide-react'
import { useState } from 'react'
import { useTranslations } from 'next-intl'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { BRAND, type Locale } from '@/lib/brand'
import { LocaleSwitcher } from '@/components/locale-switcher'

interface SiteHeaderProps {
  variant?: 'marketing' | 'app'
  userEmail?: string
  locale: Locale
}

export function SiteHeader({ variant = 'marketing', userEmail, locale }: SiteHeaderProps) {
  const t = useTranslations('common')
  const tMarketing = useTranslations('marketing')
  const tFooter = useTranslations('footer')
  const tDashboard = useTranslations('dashboard')
  const pathname = usePathname()
  const [open, setOpen] = useState(false)

  const marketingNav = [
    { label: tFooter('productFeatures'), href: `/${locale}#features` },
    { label: tFooter('productHow'), href: `/${locale}#how` },
    { label: tFooter('productPricing'), href: '#pricing' },
    { label: tMarketing('faqTitle'), href: '#faq' },
  ]
  const appNav = [
    { label: tDashboard('title'), href: `/${locale}/dashboard` },
    { label: tDashboard('newTest'), href: `/${locale}/dashboard/new` },
    { label: t('settings'), href: `/${locale}/dashboard/settings` },
  ]
  const items = variant === 'marketing' ? marketingNav : appNav

  return (
    <header
      className={cn(
        'sticky top-0 z-40 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80',
        variant === 'app' && 'bg-card'
      )}
    >
      <div className="container flex h-14 items-center justify-between">
        <Link href={variant === 'app' ? `/${locale}/dashboard` : `/${locale}`} className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground text-sm font-semibold">
            P
          </div>
          <span className="text-sm font-semibold tracking-tightish">{BRAND.name}</span>
        </Link>

        <nav className="hidden items-center gap-1 md:flex">
          {items.map((item) => {
            const active = pathname === item.href
            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  'rounded-md px-3 py-1.5 text-sm transition-colors',
                  active
                    ? 'bg-muted text-foreground'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                )}
              >
                {item.label}
              </Link>
            )
          })}
        </nav>

        <div className="hidden items-center gap-2 md:flex">
          <LocaleSwitcher currentLocale={locale} />
          {variant === 'marketing' ? (
            <>
              <Button asChild variant="ghost" size="sm">
                <Link href={`/${locale}/login`}>{t('login')}</Link>
              </Button>
              <Button asChild size="sm">
                <Link href={`/${locale}/register`}>{t('register')}</Link>
              </Button>
            </>
          ) : (
            <div className="flex items-center gap-3">
              <span className="hidden text-xs text-muted-foreground lg:inline">{userEmail}</span>
              <form action="/api/auth/logout" method="post">
                <Button type="submit" variant="outline" size="sm">
                  {t('logout')}
                </Button>
              </form>
            </div>
          )}
        </div>

        <button
          type="button"
          className="rounded-md p-2 hover:bg-muted md:hidden"
          onClick={() => setOpen(!open)}
          aria-label="Menu"
          aria-expanded={open}
        >
          {open ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {open && (
        <div className="border-t bg-card md:hidden">
          <div className="container flex flex-col gap-1 py-3">
            {items.map((item) => {
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  onClick={() => setOpen(false)}
                  className="rounded-md px-3 py-2 text-sm text-foreground hover:bg-muted"
                >
                  {item.label}
                </Link>
              )
            })}
            <div className="mt-2 flex flex-col gap-2 border-t pt-3">
              <LocaleSwitcher currentLocale={locale} mobile />
              {variant === 'marketing' ? (
                <>
                  <Button asChild variant="outline" size="sm">
                    <Link href={`/${locale}/login`}>{t('login')}</Link>
                  </Button>
                  <Button asChild size="sm">
                    <Link href={`/${locale}/register`}>{t('register')}</Link>
                  </Button>
                </>
              ) : (
                <form action="/api/auth/logout" method="post">
                  <Button type="submit" variant="outline" size="sm" className="w-full">
                    {t('logout')}
                  </Button>
                </form>
              )}
            </div>
          </div>
        </div>
      )}
    </header>
  )
}
