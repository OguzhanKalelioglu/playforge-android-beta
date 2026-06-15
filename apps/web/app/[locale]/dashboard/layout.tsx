import { redirect } from 'next/navigation'
import { SiteHeader } from '@/components/site-header'
import { getCurrentUser } from '@/lib/auth-server'
import { type Locale } from '@/lib/brand'

export default async function DashboardLayout({
  children,
  params,
}: {
  children: React.ReactNode
  params: Promise<{ locale: Locale }>
}) {
  const { locale } = await params
  const user = await getCurrentUser()
  if (!user) redirect(`/${locale}/login?next=/${locale}/dashboard`)

  return (
    <div className="flex min-h-screen flex-col bg-muted/30">
      <SiteHeader variant="app" userEmail={user.email} locale={locale} />
      <main className="flex-1">{children}</main>
    </div>
  )
}
