import Link from 'next/link'
import { CheckCircle2, ArrowRight } from 'lucide-react'
import { Button } from '@/components/ui/button'

export const metadata = { title: 'Ödeme Onaylandı' }

export default function PaymentSuccessPage() {
  return (
    <div className="container max-w-xl py-16 text-center">
      <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-success/10 text-success">
        <CheckCircle2 className="h-6 w-6" />
      </div>
      <h1 className="mt-6 text-2xl font-semibold tracking-tightish">Ödeme alındı</h1>
      <p className="mt-2 text-muted-foreground">
        Siparişin başarıyla oluşturuldu. Test 24 saat içinde başlayacak. E-posta ile bildirim
        alacaksın.
      </p>
      <div className="mt-8 flex justify-center gap-3">
        <Button asChild>
          <Link href="/dashboard">
            Testlerime Git <ArrowRight className="ml-2 h-4 w-4" />
          </Link>
        </Button>
        <Button asChild variant="outline">
          <Link href="/dashboard/new">Yeni Test Başlat</Link>
        </Button>
      </div>
    </div>
  )
}
