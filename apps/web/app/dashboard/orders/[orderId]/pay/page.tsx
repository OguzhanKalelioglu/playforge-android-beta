import { notFound } from 'next/navigation'
import { cookies } from 'next/headers'
import Link from 'next/link'
import { ChevronLeft, Lock, ShieldCheck, CreditCard } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { StatusBadge } from '@/components/ui/status-badge'
import { Money } from '@/components/ui/money'
import { PackageName } from '@/components/ui/package-name'
import { api, ApiError } from '@/lib/api'
import { formatDate } from '@/lib/format'

export const metadata = { title: 'Ödeme' }

export default async function PayPage({
  params,
}: {
  params: Promise<{ orderId: string }>
}) {
  const { orderId } = await params
  const token = (await cookies()).get('access_token')?.value ?? ''

  let order
  try {
    order = await api.order(orderId, token)
  } catch (e) {
    if (e instanceof ApiError && e.status === 404) notFound()
    throw e
  }

  if (order.status === 'paid') {
    return (
      <div className="container max-w-2xl py-12">
        <div className="rounded-md border border-success/30 bg-success/5 p-6 text-center">
          <h1 className="text-xl font-semibold">Ödeme tamamlandı</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Sipariş onaylandı. Test 24 saat içinde başlayacak.
          </p>
          <Button asChild className="mt-6">
            <Link href={`/dashboard`}>Testlerime Dön</Link>
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="container max-w-4xl py-8">
      <Link
        href="/dashboard"
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ChevronLeft className="h-4 w-4" /> Testlerim
      </Link>

      <h1 className="mt-3 text-2xl font-semibold tracking-tightish">Ödeme</h1>
      <p className="mt-1 text-sm text-muted-foreground">
        Sipariş numarası: <span className="font-mono">{order.id.slice(0, 8)}</span>
      </p>

      <div className="mt-8 grid gap-6 lg:grid-cols-[1fr_360px]">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Lock className="h-4 w-4" /> Güvenli ödeme
            </CardTitle>
          </CardHeader>
          <CardContent>
            {order.payment_url ? (
              <div className="space-y-4">
                <p className="text-sm text-muted-foreground">
                  Iyzico güvenli ödeme sayfasına yönlendiriliyorsun. Kart bilgilerin bu sitede
                  saklanmaz.
                </p>
                <Button asChild size="lg" className="w-full">
                  <a href={order.payment_url} target="_self">
                    <CreditCard className="mr-2 h-4 w-4" /> Ödemeyi Tamamla
                  </a>
                </Button>
                <p className="text-xs text-muted-foreground">
                  Bu sayfa 30 dakika sonra geçerliliğini yitirir.
                </p>
              </div>
            ) : (
              <div className="rounded-md border border-warning/30 bg-warning/5 p-4 text-sm">
                Ödeme formu henüz hazır değil. Birkaç saniye sonra yenileyin.
              </div>
            )}

            <div className="mt-6 flex items-center gap-2 text-xs text-muted-foreground">
              <ShieldCheck className="h-3.5 w-3.5" />
              <span>256-bit SSL · 3D Secure · Iyzico altyapısı</span>
            </div>
          </CardContent>
        </Card>

        <aside>
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Sipariş özeti</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <Row label="Paket" value={order.plan_name} />
              <Row label="Durum" value={<StatusBadge status={order.status} />} />
              <Row label="Toplam" value={<Money amount={order.total} />} />
              <Row label="Para birimi" value={order.currency} />
              <Row label="Oluşturuldu" value={formatDate(order.created_at)} />
              <Row label="Geçerlilik" value={formatDate(order.expires_at)} />
            </CardContent>
          </Card>
        </aside>
      </div>
    </div>
  )
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-muted-foreground">{label}</span>
      <span className="font-medium">{value}</span>
    </div>
  )
}
