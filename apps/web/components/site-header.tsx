'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { Menu, X } from 'lucide-react'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface NavItem {
  label: string
  href: string
}

interface SiteHeaderProps {
  variant?: 'marketing' | 'app'
  userEmail?: string
}

const marketingNav: NavItem[] = [
  { label: 'Özellikler', href: '/#features' },
  { label: 'Nasıl Çalışır', href: '/#how' },
  { label: 'Fiyat', href: '/#pricing' },
  { label: 'SSS', href: '/#faq' },
]

const appNav: NavItem[] = [
  { label: 'Testlerim', href: '/dashboard' },
  { label: 'Yeni Test', href: '/dashboard/new' },
  { label: 'Hesap', href: '/dashboard/settings' },
]

export function SiteHeader({ variant = 'marketing', userEmail }: SiteHeaderProps) {
  const pathname = usePathname()
  const [open, setOpen] = useState(false)
  const items = variant === 'marketing' ? marketingNav : appNav

  return (
    <header
      className={cn(
        'sticky top-0 z-40 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/80',
        variant === 'app' && 'bg-card'
      )}
    >
      <div className="container flex h-14 items-center justify-between">
        <Link href={variant === 'app' ? '/dashboard' : '/'} className="flex items-center gap-2">
          <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground text-sm font-semibold">
            T
          </div>
          <span className="text-sm font-semibold tracking-tightish">TestersCommunity</span>
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
          {variant === 'marketing' ? (
            <>
              <Button asChild variant="ghost" size="sm">
                <Link href="/login">Giriş Yap</Link>
              </Button>
              <Button asChild size="sm">
                <Link href="/register">Başla</Link>
              </Button>
            </>
          ) : (
            <div className="flex items-center gap-3">
              <span className="hidden text-xs text-muted-foreground lg:inline">{userEmail}</span>
              <form action="/api/auth/logout" method="post">
                <Button type="submit" variant="outline" size="sm">
                  Çıkış
                </Button>
              </form>
            </div>
          )}
        </div>

        <button
          type="button"
          className="rounded-md p-2 hover:bg-muted md:hidden"
          onClick={() => setOpen(!open)}
          aria-label="Menüyü aç/kapat"
          aria-expanded={open}
        >
          {open ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {open && (
        <div className="border-t bg-card md:hidden">
          <div className="container flex flex-col gap-1 py-3">
            {items.map((item) => (
              <Link
                key={item.href}
                href={item.href}
                onClick={() => setOpen(false)}
                className="rounded-md px-3 py-2 text-sm text-foreground hover:bg-muted"
              >
                {item.label}
              </Link>
            ))}
            <div className="mt-2 flex flex-col gap-2 border-t pt-3">
              {variant === 'marketing' ? (
                <>
                  <Button asChild variant="outline" size="sm">
                    <Link href="/login">Giriş Yap</Link>
                  </Button>
                  <Button asChild size="sm">
                    <Link href="/register">Başla</Link>
                  </Button>
                </>
              ) : (
                <form action="/api/auth/logout" method="post">
                  <Button type="submit" variant="outline" size="sm" className="w-full">
                    Çıkış
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
