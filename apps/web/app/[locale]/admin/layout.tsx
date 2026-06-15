import { redirect } from 'next/navigation'
import Link from 'next/link'
import { getCurrentUser } from '@/lib/auth-server'
import { SiteHeader } from '@/components/site-header'
import { cn } from '@/lib/utils'
import { type Locale } from '@/lib/brand'

export default async function AdminLayout({
  children,
  params,
}: {
  children: React.ReactNode
  params: Promise<{ locale: Locale }>
}) {
  const { locale } = await params
  const user = await getCurrentUser()
  if (!user) redirect(`/${locale}/login?next=/${locale}/admin`)
  if (user.role !== 'admin') redirect(`/${locale}/dashboard`)

  const tabs = [
    { href: `/${locale}/admin`, label: 'Overview' },
    { href: `/${locale}/admin/orders`, label: 'Orders' },
    { href: `/${locale}/admin/tests`, label: 'Tests' },
    { href: `/${locale}/admin/testers`, label: 'Testers' },
    { href: `/${locale}/admin/payments`, label: 'Payments' },
  ]

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <SiteHeader variant="app" userEmail={user.email} locale={locale} />
      <div className="border-b bg-card">
        <div className="container">
          <nav className="-mb-px flex gap-1 overflow-x-auto" aria-label="Admin">
            {tabs.map((t) => (
              <AdminTab key={t.href} href={t.href}>
                {t.label}
              </AdminTab>
            ))}
          </nav>
        </div>
      </div>
      <main className="flex-1">{children}</main>
    </div>
  )
}

function AdminTab({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <Link
      href={href}
      className={cn(
        'border-b-2 border-transparent px-3 py-3 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground',
        'aria-[current=page]:border-primary aria-[current=page]:text-foreground'
      )}
    >
      {children}
    </Link>
  )
}
