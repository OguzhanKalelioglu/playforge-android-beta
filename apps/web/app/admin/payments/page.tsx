import { cookies } from 'next/headers'
import { Card } from '@/components/ui/card'
import { StatusBadge } from '@/components/ui/status-badge'
import { Money } from '@/components/ui/money'
import { formatDate } from '@/lib/format'

const API_BASE = process.env.INTERNAL_API_URL ?? 'http://127.0.0.1:8080'

interface AdminPayment {
  id: string
  user_email: string
  test_id: string | null
  amount: number
  currency: string
  status: 'pending' | 'completed' | 'refunded' | 'failed' | 'cancelled'
  iyzico_payment_id: string | null
  paid_at: string | null
  created_at: string
}

async function getPayments(token: string): Promise<AdminPayment[]> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/admin/payments`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: 'no-store',
    })
    if (!res.ok) return []
    return res.json()
  } catch {
    return []
  }
}

export const metadata = { title: 'Admin — Ödemeler' }

export default async function AdminPaymentsPage() {
  const token = (await cookies()).get('access_token')?.value ?? ''
  const payments = await getPayments(token)
  const totalRevenue = payments
    .filter((p) => p.status === 'completed')
    .reduce((s, p) => s + p.amount, 0)
  const refunded = payments
    .filter((p) => p.status === 'refunded')
    .reduce((s, p) => s + p.amount, 0)

  return (
    <div className="container py-8">
      <header className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold tracking-tightish">Ödemeler</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Iyzico üzerinden geçen tüm ödemeler.
          </p>
        </div>
        <div className="flex gap-3 text-right">
          <div>
            <div className="text-xs text-muted-foreground">Toplam gelir</div>
            <div className="text-lg font-semibold tabular-nums">
              <Money amount={totalRevenue} />
            </div>
          </div>
          <div>
            <div className="text-xs text-muted-foreground">İade</div>
            <div className="text-lg font-semibold tabular-nums text-destructive">
              <Money amount={refunded} />
            </div>
          </div>
        </div>
      </header>

      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="border-b bg-muted/40 text-left text-xs uppercase tracking-wider text-muted-foreground">
              <tr>
                <th scope="col" className="px-4 py-3 font-medium">ID</th>
                <th scope="col" className="px-4 py-3 font-medium">Müşteri</th>
                <th scope="col" className="px-4 py-3 font-medium text-right">Tutar</th>
                <th scope="col" className="px-4 py-3 font-medium">Durum</th>
                <th scope="col" className="px-4 py-3 font-medium">Iyzico</th>
                <th scope="col" className="px-4 py-3 font-medium">Tarih</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {payments.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-12 text-center text-sm text-muted-foreground">
                    Henüz ödeme yok.
                  </td>
                </tr>
              ) : (
                payments.map((p) => (
                  <tr key={p.id} className="hover:bg-muted/30">
                    <td className="px-4 py-3 font-mono text-xs">{p.id.slice(0, 8)}</td>
                    <td className="px-4 py-3 text-xs">{p.user_email}</td>
                    <td className="px-4 py-3 text-right tabular-nums">
                      <Money amount={p.amount} />
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge status={p.status} />
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-muted-foreground">
                      {p.iyzico_payment_id ?? '—'}
                    </td>
                    <td className="px-4 py-3 text-xs text-muted-foreground">
                      {formatDate(p.created_at)}
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
