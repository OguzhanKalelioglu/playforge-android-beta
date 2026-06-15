import { cookies } from 'next/headers'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { StatusBadge } from '@/components/ui/status-badge'
import { Money } from '@/components/ui/money'
import { PackageName } from '@/components/ui/package-name'
import { formatDate } from '@/lib/format'

const API_BASE = process.env.INTERNAL_API_URL ?? 'http://127.0.0.1:8080'

interface AdminOrder {
  id: string
  user_email: string
  package_name: string
  plan_name: string
  status: string
  total: number
  currency: string
  created_at: string
  paid_at: string | null
}

async function getOrders(token: string): Promise<AdminOrder[]> {
  try {
    const res = await fetch(`${API_BASE}/api/v1/admin/orders`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: 'no-store',
    })
    if (!res.ok) return []
    return res.json()
  } catch {
    return []
  }
}

export const metadata = { title: 'Admin — Siparişler' }

export default async function AdminOrdersPage() {
  const token = (await cookies()).get('access_token')?.value ?? ''
  const orders = await getOrders(token)

  return (
    <div className="container py-8">
      <header className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold tracking-tightish">Siparişler</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Tüm siparişler, ödeme durumları ve müşteri bilgileri.
          </p>
        </div>
        <span className="text-xs text-muted-foreground">{orders.length} kayıt</span>
      </header>

      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead className="border-b bg-muted/40 text-left text-xs uppercase tracking-wider text-muted-foreground">
              <tr>
                <th scope="col" className="px-4 py-3 font-medium">Sipariş</th>
                <th scope="col" className="px-4 py-3 font-medium">Müşteri</th>
                <th scope="col" className="px-4 py-3 font-medium">Paket</th>
                <th scope="col" className="px-4 py-3 font-medium">Plan</th>
                <th scope="col" className="px-4 py-3 font-medium text-right">Tutar</th>
                <th scope="col" className="px-4 py-3 font-medium">Durum</th>
                <th scope="col" className="px-4 py-3 font-medium">Tarih</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {orders.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center text-sm text-muted-foreground">
                    Henüz sipariş yok.
                  </td>
                </tr>
              ) : (
                orders.map((o) => (
                  <tr key={o.id} className="hover:bg-muted/30">
                    <td className="px-4 py-3">
                      <span className="font-mono text-xs">{o.id.slice(0, 8)}</span>
                    </td>
                    <td className="px-4 py-3">{o.user_email}</td>
                    <td className="px-4 py-3">
                      <PackageName pkg={o.package_name} max={22} />
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">{o.plan_name}</td>
                    <td className="px-4 py-3 text-right">
                      <Money amount={o.total} />
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge status={o.status} />
                    </td>
                    <td className="px-4 py-3 text-xs text-muted-foreground">
                      {formatDate(o.created_at)}
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
