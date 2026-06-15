'use client'

import { useEffect, useState, Suspense } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Loader2, ShieldCheck, ChevronLeft } from 'lucide-react'
import Link from 'next/link'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Money } from '@/components/ui/money'
import { Badge } from '@/components/ui/badge'
import { api, type PlanTier } from '@/lib/api'

const schema = z.object({
  plan_slug: z.string().min(1, 'Paket seçin.'),
  package_name: z
    .string()
    .min(3, 'Geçerli bir paket adı girin.')
    .regex(/^[a-z][a-z0-9_]*(\.[a-z0-9_]+)+$/i, 'Format: com.example.app'),
  test_link: z.string().url('Geçerli bir URL girin.'),
})
type FormValues = z.infer<typeof schema>

function NewOrderForm() {
  const router = useRouter()
  const search = useSearchParams()
  const presetPlan = search.get('plan')
  const [plans, setPlans] = useState<PlanTier[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { plan_slug: presetPlan ?? '' },
  })

  useEffect(() => {
    api
      .plans()
      .then((p) => {
        setPlans(p)
        if (!presetPlan && p.length > 0) setValue('plan_slug', p[1]?.slug ?? p[0].slug)
        if (presetPlan) setValue('plan_slug', presetPlan)
      })
      .catch(() => setError('Planlar yüklenemedi.'))
      .finally(() => setLoading(false))
  }, [presetPlan, setValue])

  const selected = plans.find((p) => p.slug === watch('plan_slug'))

  const onSubmit = async (values: FormValues) => {
    setSubmitting(true)
    setError(null)
    try {
      const order = await api.createOrder(values, '')
      router.push(`/dashboard/orders/${order.id}/pay`)
    } catch (e) {
      setError('Sipariş oluşturulamadı. Tekrar deneyin.')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return (
      <div className="container flex min-h-[60vh] items-center justify-center">
        <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="container max-w-5xl py-8">
      <Link
        href="/dashboard"
        className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ChevronLeft className="h-4 w-4" /> Testlerim
      </Link>

      <h1 className="mt-3 text-2xl font-semibold tracking-tightish">Yeni test başlat</h1>
      <p className="mt-1 text-sm text-muted-foreground">
        Paket seç, paket adını ve test linkini gir, ödemeye geç.
      </p>

      <form onSubmit={handleSubmit(onSubmit)} className="mt-8 grid gap-8 lg:grid-cols-[1fr_320px]">
        <div className="space-y-8">
          {/* Plan selection */}
          <section>
            <Label className="text-base">1. Paket seç</Label>
            <p className="mt-1 text-sm text-muted-foreground">İhtiyacına uygun olanı seç.</p>
            <div className="mt-4 grid gap-3 sm:grid-cols-3">
              {plans.map((p) => {
                const isSel = watch('plan_slug') === p.slug
                return (
                  <label
                    key={p.slug}
                    className={`relative cursor-pointer rounded-lg border p-4 transition-colors ${
                      isSel ? 'border-primary bg-primary/5 ring-1 ring-primary/20' : 'bg-card hover:border-muted-foreground/30'
                    }`}
                  >
                    <input type="radio" value={p.slug} className="sr-only" {...register('plan_slug')} />
                    <div className="flex items-center justify-between">
                      <span className="text-sm font-semibold">{p.name}</span>
                      {p.slug === 'pro' && <Badge>Önerilen</Badge>}
                    </div>
                    <div className="mt-2 text-2xl font-semibold tabular-nums">
                      {p.price_try.toLocaleString('tr-TR')} ₺
                    </div>
                    <p className="mt-1 text-xs text-muted-foreground">
                      {p.tester_count} hesap · {p.duration_days} gün
                    </p>
                  </label>
                )
              })}
            </div>
          </section>

          {/* Package details */}
          <section className="space-y-4">
            <div>
              <Label className="text-base">2. Paket bilgileri</Label>
              <p className="mt-1 text-sm text-muted-foreground">
                Google Play Console&apos;da görünen tam paket adı ve kapalı test davet linki.
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="package_name">Paket adı</Label>
              <Input
                id="package_name"
                placeholder="com.example.myapp"
                className="font-mono"
                aria-invalid={!!errors.package_name}
                {...register('package_name')}
              />
              {errors.package_name && (
                <p className="text-xs text-destructive">{errors.package_name.message}</p>
              )}
              <p className="text-xs text-muted-foreground">Örn: com.spotify.music</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="test_link">Kapalı test linki</Label>
              <Input
                id="test_link"
                type="url"
                placeholder="https://play.google.com/apps/internaltest/..."
                aria-invalid={!!errors.test_link}
                {...register('test_link')}
              />
              {errors.test_link && (
                <p className="text-xs text-destructive">{errors.test_link.message}</p>
              )}
              <p className="text-xs text-muted-foreground">
                Play Console → Testing → Closed testing → Testers sekmesinden kopyala.
              </p>
            </div>
          </section>

          {error && (
            <div role="alert" className="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive">
              {error}
            </div>
          )}

          <Button type="submit" size="lg" disabled={submitting}>
            {submitting ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" /> Sipariş oluşturuluyor
              </>
            ) : (
              'Ödemeye Geç'
            )}
          </Button>
        </div>

        {/* Order summary */}
        <aside className="lg:sticky lg:top-20 lg:self-start">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Sipariş özeti</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              {selected ? (
                <>
                  <div className="flex justify-between">
                    <span className="text-muted-foreground">{selected.name} paket</span>
                    <Money amount={selected.price_try} />
                  </div>
                  <div className="flex justify-between text-xs text-muted-foreground">
                    <span>KDV dahil</span>
                    <span>14 gün, 25 hesap</span>
                  </div>
                  <div className="border-t pt-3">
                    <div className="flex justify-between font-semibold">
                      <span>Toplam</span>
                      <Money amount={selected.price_try} />
                    </div>
                  </div>
                </>
              ) : (
                <p className="text-muted-foreground">Paket seçilmemiş.</p>
              )}
              <div className="flex items-start gap-2 rounded-md bg-muted/50 p-3 text-xs text-muted-foreground">
                <ShieldCheck className="mt-0.5 h-3.5 w-3.5 shrink-0" />
                <span>256-bit SSL. 3D Secure. Iyzico güvencesiyle.</span>
              </div>
            </CardContent>
          </Card>
        </aside>
      </form>
    </div>
  )
}

export default function NewOrderPage() {
  return (
    <Suspense fallback={null}>
      <NewOrderForm />
    </Suspense>
  )
}
