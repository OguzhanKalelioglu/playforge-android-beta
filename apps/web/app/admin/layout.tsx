import { redirect } from 'next/navigation'
import Link from 'next/link'
import { getCurrentUser } from '@/lib/auth-server'
import { SiteHeader } from '@/components/site-header'
import { cn } from '@/lib/utils'

const tabs = [
  { href: '/admin', label: 'Genel Bakış' },
  { href: '/admin/orders', label: 'Siparişler' },
  { href: '/admin/tests', label: 'Testler' },
  { href: '/admin/testers', label: 'Testerlar' },
  { href: '/admin/payments', label: 'Ödemeler' },
]

export default async function AdminLayout({ children }: { children: React.ReactNode }) {
  const user = await getCurrentUser()
  if (!user) redirect('/login?next=/admin')
  if (user.role !== 'admin') redirect('/dashboard')

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <SiteHeader variant="app" userEmail={user.email} />
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
