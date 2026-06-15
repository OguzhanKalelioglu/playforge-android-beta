import Link from 'next/link'
import {
  CheckCircle2,
  Smartphone,
  Clock,
  BarChart3,
  Star,
  ArrowRight,
  AlertTriangle,
  Shield,
  Mail,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { SiteHeader } from '@/components/site-header'
import { SiteFooter } from '@/components/site-footer'
import { formatCurrency } from '@/lib/format'

const features = [
  {
    icon: Smartphone,
    title: '25 yönetilen hesap',
    description:
      'Her biri kendi cihaz profiliyle (model, Android sürümü, dil, saat dilimi) farklı bir Google hesabı. Aynı anda çalışırlar.',
  },
  {
    icon: Clock,
    title: '14 günlük otomasyon',
    description:
      "Google'ın 14-günlük kapalı test eşiğini rahatça geçer. Günlük engagement, opt-in ve indirme dahil.",
  },
  {
    icon: BarChart3,
    title: 'Canlı ilerleme',
    description:
      'Her hesabın aktivitesini, screenshotları ve engagement dakikalarını dashboarddan anlık izleyin.',
  },
  {
    icon: Star,
    title: 'Yorumlar dahil',
    description:
      '14. günde 10 organik görünümlü yorum, karışık puanla. Tonu ve dili seçebilirsiniz.',
  },
  {
    icon: Shield,
    title: 'Yumuşak warming',
    description:
      'Yeni hesaplar 3 günlük warming döneminden geçer; organik davranış simülasyonu uygulanır.',
  },
  {
    icon: Mail,
    title: 'Bildirimler',
    description:
      'Ödeme onayı, test başlangıcı, günlük özet ve tamamlanma raporu e-posta ile gelir.',
  },
]

const how = [
  {
    step: '01',
    title: 'Paket adınızı girin',
    body: 'Google Play Console\'daki paket adınızı ve kapalı test linkinizi forma yazın. 1 dakika.',
  },
  {
    step: '02',
    title: 'Ödemeyi yapın',
    body: '₺4.999 — kredi kartı, banka kartı. Iyzico altyapısı, 3D Secure, anlık onay.',
  },
  {
    step: '03',
    title: 'Sonuçları izleyin',
    body: '24 saat içinde 25 hesap opt-in yapar. Günlük engagement başlar. 14. gün rapor hazır.',
  },
]

const tiers = [
  {
    slug: 'basic',
    name: 'Basic',
    price: 4999,
    description: 'Başlangıç için yeterli.',
    features: [
      '25 Google Play tester hesabı',
      '14 günlük engagement',
      'Opt-in + indirme otomatik',
      'Günlük aktivite logları',
      '10 adet review (karışık puan)',
      'E-posta ile günlük özet',
      'E-posta destek',
    ],
    highlighted: false,
  },
  {
    slug: 'pro',
    name: 'Pro',
    price: 7999,
    description: 'Yaygın seçim.',
    features: [
      'Basic\'in tüm özellikleri',
      'Haftalık detaylı ilerleme raporu',
      'Screenshot timeline',
      'Custom yorum tonu ve dili',
      'Öncelikli destek (24 saat yanıt)',
      'WhatsApp bildirim kanalı',
    ],
    highlighted: true,
  },
  {
    slug: 'enterprise',
    name: 'Enterprise',
    price: 12999,
    description: 'Yüksek bütçeli ve sürekli yayınlar için.',
    features: [
      'Pro\'nun tüm özellikleri',
      'Premium ban koruma (rotasyonlu hesap)',
      'Dedicated account manager',
      'API erişimi',
      'SLA — 4 saat yanıt',
      'Aylık 1 ücretsiz yeniden test',
    ],
    highlighted: false,
  },
]

const faqs = [
  {
    q: 'Bu hizmet yasal mı?',
    a: 'Google Play Console Dağıtım Sözleşmesi, otomatik indirme ve yorum artırımını kısıtlar. Bu hizmet bu kuralların sınırında hareket eder. Müşteri sözleşmemiz risk paylaşımını detaylı açıklar — kullanmadan önce okuyun.',
  },
  {
    q: 'Test ne kadar sürede başlar?',
    a: 'Ödeme onayından sonra 24 saat içinde opt-in ve indirme başlar. 14 günlük döngü başlatılır.',
  },
  {
    q: 'Para iadesi var mı?',
    a: 'Test başlamadan önce tam iade. Başladıktan sonra kullanılmayan günler için kısmi iade (en fazla %50). İade politikamıza bakın.',
  },
  {
    q: 'Hangi ülkeler destekleniyor?',
    a: 'Şu anda Türkiye, ABD ve Avrupa. Paket adınız bu ülkelerde yayınlanıyorsa test yapılabilir.',
  },
  {
    q: 'Yorumlar gerçek mi?',
    a: 'Yorumlar gerçek hesaplardan yazılır, ancak Google ToS\'una göre "organik olmayan" sayılabilir. Bu risk müşteriye aittir.',
  },
  {
    q: 'Hesaplar banlenir mi?',
    a: 'Yumuşak warming ve organik davranış simülasyonu uyguluyoruz, ancak Google\'ın algoritma değişiklikleri risk taşır. Enterprise pakette rotasyonlu hesap kullanıyoruz.',
  },
]

export default function HomePage() {
  return (
    <div className="flex min-h-screen flex-col">
      <SiteHeader variant="marketing" />

      <main className="flex-1">
        {/* HERO */}
        <section className="border-b">
          <div className="container py-20 md:py-28">
            <div className="mx-auto max-w-3xl text-center">
              <Badge variant="muted" className="mb-6">
                Google Play 14-gün şartı için
              </Badge>
              <h1 className="text-4xl font-semibold tracking-tightish md:text-6xl">
                25 hesap, 14 gün,
                <br />
                <span className="text-primary">otomatik test.</span>
              </h1>
              <p className="mt-6 text-lg text-muted-foreground md:text-xl">
                Uygulamanız 25 yönetilen Google hesabı tarafından indirilir, günlük olarak
                kullanılır ve 14. günde yorumlanır. Siz sadece sonuçları izlersiniz.
              </p>
              <div className="mt-10 flex flex-col items-center justify-center gap-3 sm:flex-row">
                <Button size="lg" asChild>
                  <Link href="/register">
                    Test Başlat <ArrowRight className="ml-2 h-4 w-4" />
                  </Link>
                </Button>
                <Button size="lg" variant="ghost" asChild>
                  <Link href="#pricing">Fiyatları Gör</Link>
                </Button>
              </div>
              <dl className="mt-16 grid grid-cols-3 gap-4 border-y py-6 text-left">
                <div>
                  <dt className="text-xs text-muted-foreground">Eşzamanlı hesap</dt>
                  <dd className="mt-1 text-2xl font-semibold tabular-nums">25</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground">Test süresi</dt>
                  <dd className="mt-1 text-2xl font-semibold tabular-nums">14 gün</dd>
                </div>
                <div>
                  <dt className="text-xs text-muted-foreground">Ortalama başlangıç</dt>
                  <dd className="mt-1 text-2xl font-semibold tabular-nums">~18 sa</dd>
                </div>
              </dl>
            </div>
          </div>
        </section>

        {/* FEATURES */}
        <section id="features" className="border-b">
          <div className="container py-20">
            <div className="mx-auto max-w-2xl text-center">
              <h2 className="text-3xl font-semibold tracking-tightish md:text-4xl">
                Ne yapıyoruz
              </h2>
              <p className="mt-3 text-muted-foreground">
                Uçtan uca otomatik. Siz sadece paket adını girip ödemeyi yaparsınız.
              </p>
            </div>
            <div className="mt-12 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {features.map((f) => (
                <Card key={f.title}>
                  <CardHeader>
                    <f.icon className="h-5 w-5 text-primary" strokeWidth={1.75} />
                    <CardTitle className="mt-4 text-base">{f.title}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm text-muted-foreground">{f.description}</p>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        </section>

        {/* HOW IT WORKS */}
        <section id="how" className="border-b bg-muted/30">
          <div className="container py-20">
            <div className="mx-auto max-w-2xl text-center">
              <h2 className="text-3xl font-semibold tracking-tightish md:text-4xl">
                Nasıl çalışır
              </h2>
              <p className="mt-3 text-muted-foreground">3 adım. Ödeme onayından 24 saat sonra başlar.</p>
            </div>
            <ol className="mt-12 grid gap-8 md:grid-cols-3">
              {how.map((s) => (
                <li key={s.step} className="relative rounded-lg border bg-card p-6">
                  <div className="font-mono text-xs text-muted-foreground">ADIM {s.step}</div>
                  <h3 className="mt-3 text-lg font-semibold">{s.title}</h3>
                  <p className="mt-2 text-sm text-muted-foreground">{s.body}</p>
                </li>
              ))}
            </ol>
          </div>
        </section>

        {/* PRICING */}
        <section id="pricing" className="border-b">
          <div className="container py-20">
            <div className="mx-auto max-w-2xl text-center">
              <h2 className="text-3xl font-semibold tracking-tightish md:text-4xl">Fiyatlandırma</h2>
              <p className="mt-3 text-muted-foreground">
                Paket başına tek ödeme. 14 günlük tam test.
              </p>
            </div>
            <div className="mt-12 grid gap-6 md:grid-cols-3">
              {tiers.map((t) => (
                <Card
                  key={t.slug}
                  className={
                    t.highlighted
                      ? 'border-primary shadow-sm ring-1 ring-primary/20'
                      : ''
                  }
                >
                  <CardHeader>
                    <div className="flex items-center justify-between">
                      <CardTitle className="text-lg">{t.name}</CardTitle>
                      {t.highlighted && <Badge>Önerilen</Badge>}
                    </div>
                    <div className="mt-4 flex items-baseline gap-1">
                      <span className="text-4xl font-semibold tabular-nums">
                        {formatCurrency(t.price)}
                      </span>
                      <span className="text-sm text-muted-foreground">/ test</span>
                    </div>
                    <CardDescription className="mt-2">{t.description}</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <ul className="space-y-2.5 text-sm">
                      {t.features.map((f) => (
                        <li key={f} className="flex items-start gap-2">
                          <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-success" strokeWidth={1.75} />
                          <span>{f}</span>
                        </li>
                      ))}
                    </ul>
                    <Button
                      className="mt-6 w-full"
                      size="lg"
                      variant={t.highlighted ? 'default' : 'outline'}
                      asChild
                    >
                      <Link href={`/register?plan=${t.slug}`}>Bu Paketi Seç</Link>
                    </Button>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        </section>

        {/* FAQ */}
        <section id="faq" className="border-b bg-muted/30">
          <div className="container py-20">
            <div className="mx-auto max-w-2xl text-center">
              <h2 className="text-3xl font-semibold tracking-tightish md:text-4xl">
                Sık sorulan sorular
              </h2>
            </div>
            <div className="mx-auto mt-12 grid max-w-3xl gap-3">
              {faqs.map((faq) => (
                <details
                  key={faq.q}
                  className="group rounded-lg border bg-card p-4 [&_summary::-webkit-details-marker]:hidden"
                >
                  <summary className="flex cursor-pointer items-center justify-between gap-4 font-medium">
                    {faq.q}
                    <span
                      aria-hidden
                      className="ml-auto text-muted-foreground transition-transform group-open:rotate-180"
                    >
                      ▾
                    </span>
                  </summary>
                  <p className="mt-3 text-sm text-muted-foreground">{faq.a}</p>
                </details>
              ))}
            </div>
          </div>
        </section>

        {/* LEGAL WARNING */}
        <section className="border-b bg-warning/5">
          <div className="container py-8">
            <div className="mx-auto flex max-w-3xl items-start gap-3">
              <AlertTriangle className="mt-0.5 h-5 w-5 shrink-0 text-warning" strokeWidth={1.75} />
              <div>
                <h3 className="text-sm font-semibold">Yasal uyarı</h3>
                <p className="mt-1 text-sm text-muted-foreground">
                  Bu hizmet Google Play Store Hizmet Şartları&apos;nın bazı maddelerini ihlal
                  edebilir. Müşteri sözleşmemizde tüm risk paylaşımı detaylı açıklanmıştır.
                  Hizmeti kullanmadan önce{' '}
                  <Link href="/legal/terms" className="underline">
                    kullanım şartlarını
                  </Link>{' '}
                  ve{' '}
                  <Link href="/legal/refund" className="underline">
                    iade politikasını
                  </Link>{' '}
                  okuyun.
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* CTA */}
        <section className="container py-20 text-center">
          <h2 className="text-3xl font-semibold tracking-tightish md:text-4xl">
            Başlamak için hazır mısın?
          </h2>
          <p className="mx-auto mt-3 max-w-lg text-muted-foreground">
            Dakikalar içinde hesap oluştur, paket adını gir, ödemeyi yap. Geri kalanını sistem
            halleder.
          </p>
          <Button size="lg" className="mt-8" asChild>
            <Link href="/register">
              Hesap Oluştur <ArrowRight className="ml-2 h-4 w-4" />
            </Link>
          </Button>
        </section>
      </main>

      <SiteFooter />
    </div>
  )
}
