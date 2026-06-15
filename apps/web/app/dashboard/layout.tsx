import { redirect } from 'next/navigation'
import { SiteHeader } from '@/components/site-header'
import { getCurrentUser } from '@/lib/auth-server'

export default async function DashboardLayout({ children }: { children: React.ReactNode }) {
  const user = await getCurrentUser()
  if (!user) redirect('/login?next=/dashboard')

  return (
    <div className="flex min-h-screen flex-col bg-muted/30">
      <SiteHeader variant="app" userEmail={user.email} />
      <main className="flex-1">{children}</main>
    </div>
  )
}
