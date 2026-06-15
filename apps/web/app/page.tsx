import Link from 'next/link'
import { CheckCircle2, Shield, Clock, BarChart3, Users, Star, ArrowRight, AlertTriangle } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

const features = [
  {
    icon: Users,
    title: '25 Gerçek Hesap',
    description: '25 farklı Google hesabıyla eş zamanlı test. Her hesap kendine özel cihaz profili kullanır.',
  },
  {
    icon: Clock,
    title: '14 Gün Kapsamlı',
    description: "Google'ın 20 aktif kullanıcı şartını rahatça geçer. Günlük engagement otomatik olarak yapılır.",
  },
  {
    icon: BarChart3,
    title: 'Detaylı Rapor',
    description: 'Her hesabın aktivitesini, screenshotları ve gerçek zamanlı ilerlemeyi dashboarddan izleyin.',
  },
  {
    icon: Star,
    title: 'Bonus Reviews',
    description: '10 hesaptan organik görünümlü, karışık puanlı kullanıcı yorumu.',
  },
]

const howItWorks = [
  {
    step: '01',
    title: 'Paket Adınızı Girin',
    description: 'Google Play Store\'daki paket adınızı ve test linkinizi forma girin. 1 dakika sürer.',
  },
  {
    step: '02',
    title: 'Ödemeyi Yapın',
    description: '₺2.999 — kredi kartı veya banka kartı ile. Iyzico güvencesiyle 3D Secure ödeme.',
  },
  {
    step: '03',
    title: 'Sonuçları İzleyin',
    description: '24 saat içinde test başlar. Dashboard\'dan canlı ilerlemeyi takip edin. 14 gün sonunda detaylı rapor.',
  },
]

const faqs = [
  {
    q: 'Bu hizmet yasal mı?',
    a: 'Google Play Console Dağıtım Sözleşmesi otomatik indirme/yorum artırımını kısıtlar. Bu hizmet bu kuralların sınırında hareket eder. Müşteri sözleşmemizde risk paylaşımı detaylı açıklanmıştır.',
  },
  {
    q: 'Test ne kadar sürede başlar?',
    a: 'Ödeme onayından sonra 24 saat içinde 25 hesap için opt-in ve indirme işlemi başlar. Toplam 14 günlük test süreci başlatılır.',
  },
  {
    q: 'Para iadesi var mı?',
    a: 'Test başlamadan önce tam iade. Başladıktan sonra kullanılmayan günler için kısmi iade (en fazla %50). İade politikamızı inceleyin.',
  },
  {
    q: 'Hangi ülkeler destekleniyor?',
    a: 'Şu anda Türkiye, ABD ve Avrupa. Paket adınız bu ülkelerde yayınlanıyorsa test yapılabilir.',
  },
  {
    q: 'Reviews gerçek mi?',
    a: 'Reviews gerçek hesaplardan yazılır, ancak Google\'ın ToS\'una göre "organik olmayan" sayılabilir. Bu risk müşteriye aittir.',
  },
  {
    q: 'Hesaplar banlenir mi?',
    a: 'Yumuşak warming ve organik davranış simülasyonu uyguluyoruz, ancak Google\'ın algoritma değişiklikleri risk taşır. Bu risk platformumuzun kontrolü dışındadır.',
  },
]

export default function HomePage() {
  return (
    <div className="flex min-h-screen flex-col">
      <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur">
        <div className="container flex h-16 items-center justify-between">
          <Link href="/" className="flex items-center gap-2 text-lg font-semibold">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
              T
            </div>
            TestersCommunity
          </Link>
          <nav className="flex items-center gap-4">
            <Link href="/login" className="text-sm text-muted-foreground hover:text-foreground">
              Giriş Yap
            </Link>
            <Button asChild>
              <Link href="/register">Başla</Link>
            </Button>
          </nav>
        </div>
      </header>

      <main className="flex-1">
        <section className="container py-20 md:py-32">
          <div className="mx-auto max-w-3xl text-center">
            <Badge variant="secondary" className="mb-6">
              Google Play 14-gün şartı için
            </Badge>
            <h1 className="text-4xl font-bold tracking-tight md:text-6xl">
              Uygulamanızı 25 gerçek kullanıcıyla test edin
            </h1>
            <p className="mt-6 text-lg text-muted-foreground md:text-xl">
              14 günlük kapsamlı kapalı beta testi. Günlük engagement, detaylı rapor ve bonus reviews — hepsi otomatik.
            </p>
            <div className="mt-10 flex flex-col items-center justify-center gap-4 sm:flex-row">
              <Button size="lg" asChild>
                <Link href="/register">
                  Hemen Başla <ArrowRight className="ml-2 h-4 w-4" />
                </Link>
              </Button>
              <Button size="lg" variant="outline" asChild>
                <Link href="#pricing">Fiyatı Gör</Link>
              </Button>
            </div>
            <div className="mt-12 flex items-center justify-center gap-8 text-sm text-muted-foreground">
              <div className="flex items-center gap-2">
                <CheckCircle2 className="h-4 w-4 text-green-600" /> 24 saatte başlar
              </div>
              <div className="flex items-center gap-2">
                <CheckCircle2 className="h-4 w-4 text-green-600" /> 14 günlük test
              </div>
              <div className="flex items-center gap-2">
                <CheckCircle2 className="h-4 w-4 text-green-600" /> Canlı rapor
              </div>
            </div>
          </div>
        </section>

        <section className="border-y bg-muted/40">
          <div className="container py-20">
            <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
              {features.map((feature) => (
                <Card key={feature.title}>
                  <CardHeader>
                    <feature.icon className="h-10 w-10 text-primary" />
                    <CardTitle className="mt-4">{feature.title}</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <CardDescription>{feature.description}</CardDescription>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        </section>

        <section className="container py-20">
          <div className="mx-auto max-w-2xl text-center">
            <h2 className="text-3xl font-bold tracking-tight md:text-4xl">Nasıl Çalışır?</h2>
            <p className="mt-4 text-lg text-muted-foreground">3 basit adım, 14 günlük test.</p>
          </div>
          <div className="mt-12 grid gap-8 md:grid-cols-3">
            {howItWorks.map((item) => (
              <div key={item.step} className="relative">
                <div className="text-6xl font-bold text-primary/20">{item.step}</div>
                <h3 className="mt-2 text-xl font-semibold">{item.title}</h3>
                <p className="mt-2 text-muted-foreground">{item.description}</p>
              </div>
            ))}
          </div>
        </section>

        <section id="pricing" className="border-y bg-muted/40">
          <div className="container py-20">
            <div className="mx-auto max-w-md">
              <Card>
                <CardHeader className="text-center">
                  <Badge className="mx-auto">Tek Paket</Badge>
                  <CardTitle className="mt-4 text-3xl">₺2.999</CardTitle>
                  <CardDescription>14 günlük tam test, tek ödeme</CardDescription>
                </CardHeader>
                <CardContent>
                  <ul className="space-y-3 text-sm">
                    {[
                      '25 farklı Google hesabı',
                      '14 günlük engagement',
                      'Opt-in + indirme otomatik',
                      'Günlük aktivite logları',
                      'Screenshot raporları',
                      '10 adet review (karışık puan)',
                      'Detaylı PDF rapor',
                      'Canlı dashboard',
                      'Iyzico güvenli ödeme',
                    ].map((item) => (
                      <li key={item} className="flex items-start gap-2">
                        <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-green-600" />
                        <span>{item}</span>
                      </li>
                    ))}
                  </ul>
                  <Button className="mt-6 w-full" size="lg" asChild>
                    <Link href="/register">Satın Al</Link>
                  </Button>
                </CardContent>
              </Card>
            </div>
          </div>
        </section>

        <section className="container py-20">
          <div className="mx-auto max-w-2xl text-center">
            <h2 className="text-3xl font-bold tracking-tight md:text-4xl">Sık Sorulan Sorular</h2>
          </div>
          <div className="mx-auto mt-12 max-w-3xl space-y-4">
            {faqs.map((faq) => (
              <Card key={faq.q}>
                <CardHeader>
                  <CardTitle className="text-lg">{faq.q}</CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-sm text-muted-foreground">{faq.a}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </section>

        <section className="border-t bg-yellow-50 dark:bg-yellow-950/20">
          <div className="container py-8">
            <div className="mx-auto max-w-3xl">
              <div className="flex items-start gap-3">
                <AlertTriangle className="mt-1 h-5 w-5 shrink-0 text-yellow-600" />
                <div>
                  <h3 className="font-semibold">Yasal Uyarı</h3>
                  <p className="mt-1 text-sm text-muted-foreground">
                    Bu hizmet Google Play Store Hizmet Şartları&apos;nın bazı maddelerini ihlal edebilir.
                    Müşteri sözleşmemizde tüm risk paylaşımı detaylı olarak açıklanmıştır. Hizmeti kullanmadan
                    önce <Link href="/legal/terms" className="underline">kullanım şartlarını</Link> ve{' '}
                    <Link href="/legal/refund" className="underline">iade politikasını</Link> okuyun.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="container py-20">
          <div className="mx-auto max-w-2xl text-center">
            <h2 className="text-3xl font-bold tracking-tight md:text-4xl">Başlamak için hazır mısın?</h2>
            <p className="mt-4 text-lg text-muted-foreground">
              Dakikalar içinde hesap oluştur, paket adını gir, ödemeyi yap. Geri kalanı biz hallediyoruz.
            </p>
            <div className="mt-8">
              <Button size="lg" asChild>
                <Link href="/register">
                  Ücretsiz Hesap Oluştur <ArrowRight className="ml-2 h-4 w-4" />
                </Link>
              </Button>
            </div>
          </div>
        </section>
      </main>

      <footer className="border-t py-8">
        <div className="container flex flex-col items-center justify-between gap-4 md:flex-row">
          <div className="text-sm text-muted-foreground">
            © 2026 TestersCommunity. Tüm hakları saklıdır.
          </div>
          <nav className="flex gap-6 text-sm text-muted-foreground">
            <Link href="/legal/terms" className="hover:text-foreground">Kullanım Şartları</Link>
            <Link href="/legal/privacy" className="hover:text-foreground">Gizlilik</Link>
            <Link href="/legal/refund" className="hover:text-foreground">İade Politikası</Link>
            <Link href="mailto:hello@testerscomm.net" className="hover:text-foreground">İletişim</Link>
          </nav>
        </div>
      </footer>
    </div>
  )
}
