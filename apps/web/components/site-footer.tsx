import Link from 'next/link'

export function SiteFooter() {
  return (
    <footer className="border-t bg-card">
      <div className="container py-8">
        <div className="flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <div className="flex h-6 w-6 items-center justify-center rounded bg-primary text-primary-foreground text-xs font-semibold">
              T
            </div>
            <span>© 2026 TestersCommunity</span>
          </div>
          <nav className="flex flex-wrap gap-x-6 gap-y-2 text-sm text-muted-foreground">
            <Link href="/legal/terms" className="hover:text-foreground">
              Kullanım Şartları
            </Link>
            <Link href="/legal/privacy" className="hover:text-foreground">
              Gizlilik
            </Link>
            <Link href="/legal/refund" className="hover:text-foreground">
              İade Politikası
            </Link>
            <Link href="mailto:hello@testerscomm.net" className="hover:text-foreground">
              İletişim
            </Link>
          </nav>
        </div>
      </div>
    </footer>
  )
}
