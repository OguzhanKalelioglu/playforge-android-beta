import { cookies } from 'next/headers'
import { Card } from '@/components/ui/card'
import { StatusBadge } from '@/components/ui/status-badge'
import { PackageName } from '@/components/ui/package-name'
import { ProgressBar } from '@/components/ui/progress'
import { formatDate } from '@/lib/format'

const API_BASE = process.env.INTERNAL_API_URL ?? 'http://127.0.0.1:8080'

interface AdminTest {
  id: string
  user_email: string
  package_name: string
  status: string
  progress: { total: number; installed: number; reviewed: number }
  started_at: string | null
  ends_at: string | null
  created_at: string
}

async function getTests(token: string): Promise<AdminTest[]> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/admin/tests`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: 'no-store',
    })
    if (!res.ok) return []
    return res.json()
  } catch {
    return []
  }
}

export const metadata = { title: 'Admin — Testler' }

export default async function AdminTestsPage() {
  const token = (await cookies()).get('access_token')?.value ?? ''
  const tests = await getTests(token)

  return (
    <div className="container py-8">
      <header className="mb-6">
        <h1 className="text-2xl font-semibold tracking-tightish">Testler</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Tüm aktif ve geçmiş testler, müşteri bazında.
        </p>
      </header>

      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="border-b bg-muted/40 text-left text-xs uppercase tracking-wider text-muted-foreground">
              <tr>
                <th scope="col" className="px-4 py-3 font-medium">Müşteri</th>
                <th scope="col" className="px-4 py-3 font-medium">Paket</th>
                <th scope="col" className="px-4 py-3 font-medium">Durum</th>
                <th scope="col" className="px-4 py-3 font-medium w-64">İlerleme</th>
                <th scope="col" className="px-4 py-3 font-medium">Tarih</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {tests.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center text-sm text-muted-foreground">
                    Henüz test yok.
                  </td>
                </tr>
              ) : (
                tests.map((t) => (
                  <tr key={t.id} className="hover:bg-muted/30">
                    <td className="px-4 py-3 text-xs">{t.user_email}</td>
                    <td className="px-4 py-3">
                      <PackageName pkg={t.package_name} />
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge status={t.status} />
                    </td>
                    <td className="px-4 py-3">
                      <ProgressBar
                        value={t.progress.installed}
                        total={t.progress.total}
                        showLabel
                        size="sm"
                      />
                    </td>
                    <td className="px-4 py-3 text-xs text-muted-foreground">
                      {t.started_at ? formatDate(t.started_at) : formatDate(t.created_at)}
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
