import { cookies } from 'next/headers'
import { Card } from '@/components/ui/card'
import { StatusBadge } from '@/components/ui/status-badge'
import { formatDate } from '@/lib/format'
import { maskEmail } from '@/lib/format'

const API_BASE = process.env.INTERNAL_API_URL ?? 'http://127.0.0.1:8080'

interface AdminTester {
  id: string
  email: string
  status: 'warming' | 'active' | 'cooling' | 'disabled'
  group_email: string | null
  last_used_at: string | null
  device_model: string
  tasks_completed_30d: number
}

async function getTesters(token: string): Promise<AdminTester[]> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/admin/testers`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: 'no-store',
    })
    if (!res.ok) return []
    return res.json()
  } catch {
    return []
  }
}

export const metadata = { title: 'Admin — Testerlar' }

export default async function AdminTestersPage() {
  const token = (await cookies()).get('access_token')?.value ?? ''
  const testers = await getTesters(token)

  const grouped = {
    active: testers.filter((t) => t.status === 'active').length,
    warming: testers.filter((t) => t.status === 'warming').length,
    cooling: testers.filter((t) => t.status === 'cooling').length,
    disabled: testers.filter((t) => t.status === 'disabled').length,
  }

  return (
    <div className="container py-8">
      <header className="mb-6">
        <h1 className="text-2xl font-semibold tracking-tightish">Tester havuzu</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          25 yönetilen Google hesabı. Durum, son kullanım, cihaz profili.
        </p>
      </header>

      <div className="mb-4 flex flex-wrap gap-2 text-xs">
        <span className="rounded-full bg-success/10 px-2.5 py-1 font-medium text-success">
          {grouped.active} aktif
        </span>
        <span className="rounded-full bg-warning/10 px-2.5 py-1 font-medium text-warning">
          {grouped.warming} warming
        </span>
        <span className="rounded-full bg-muted px-2.5 py-1 font-medium text-muted-foreground">
          {grouped.cooling} cooling
        </span>
        <span className="rounded-full bg-muted px-2.5 py-1 font-medium text-muted-foreground">
          {grouped.disabled} devre dışı
        </span>
      </div>

      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="border-b bg-muted/40 text-left text-xs uppercase tracking-wider text-muted-foreground">
              <tr>
                <th scope="col" className="px-4 py-3 font-medium">Email</th>
                <th scope="col" className="px-4 py-3 font-medium">Durum</th>
                <th scope="col" className="px-4 py-3 font-medium">Cihaz</th>
                <th scope="col" className="px-4 py-3 font-medium">Grup</th>
                <th scope="col" className="px-4 py-3 font-medium text-right">30g görev</th>
                <th scope="col" className="px-4 py-3 font-medium">Son kullanım</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {testers.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-12 text-center text-sm text-muted-foreground">
                    Henüz tester eklenmemiş. Sıcaklık için 3 günlük warming başlat.
                  </td>
                </tr>
              ) : (
                testers.map((t) => (
                  <tr key={t.id} className="hover:bg-muted/30">
                    <td className="px-4 py-3 font-mono text-xs">{maskEmail(t.email)}</td>
                    <td className="px-4 py-3">
                      <StatusBadge status={t.status} />
                    </td>
                    <td className="px-4 py-3 text-xs text-muted-foreground">
                      {t.device_model}
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-muted-foreground">
                      {t.group_email ?? '—'}
                    </td>
                    <td className="px-4 py-3 text-right tabular-nums">
                      {t.tasks_completed_30d}
                    </td>
                    <td className="px-4 py-3 text-xs text-muted-foreground">
                      {t.last_used_at ? formatDate(t.last_used_at) : '—'}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  )
}
