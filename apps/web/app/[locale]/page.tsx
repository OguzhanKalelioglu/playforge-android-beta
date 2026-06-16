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
import { getTranslations } from 'next-intl/server'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { SiteHeader } from '@/components/site-header'
import { SiteFooter } from '@/components/site-footer'
import { formatCurrency } from '@/lib/format'
import { LOCALE_CONFIG, type Locale } from '@/lib/brand'
import { PLANS, getCurrencyForLocale, getIntlLocale } from '@/lib/pricing'

const featureKeys = ['managed', 'automation', 'live', 'reviews', 'warming', 'notifications'] as const
const featureIcons: Record<(typeof featureKeys)[number], typeof Smartphone> = {
  managed: Smartphone,
  automation: Clock,
  live: BarChart3,
  reviews: Star,
  warming: Shield,
  notifications: Mail,
}

const stepKeys = ['step1', 'step2', 'step3'] as const

export default async function HomePage({ params }: { params: Promise<{ locale: Locale }> }) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'marketing' })
  const currency = getCurrencyForLocale(locale)
  const intl = getIntlLocale(locale)

  return (
    <div className="flex min-h-screen flex-col">
      <SiteHeader variant="marketing" locale={locale} />

      <main className="flex-1">
        {/* HERO */}
        <section className="border-b">
          <div className="container py-20 md:py-28">
            <div className="mx-auto max-w-3xl text-center">
              <Badge variant="muted" className="mb-6">
                {t('badge')}
              </Badge>
              <h1 className="text-4xl font-semibold tracking-tightish md:text-6xl">
                {t('heroTitle1')}
                <br />
                <span className="text-primary">{t('heroTitle2')}</span>
              </h1>
              <p className="mt-6 text-lg text-muted-foreground md:text-xl">
                {t('heroSubtitle')}
              </p>
              <div className="mt-10 flex flex-col items-center justify-center gap-3 sm:flex-row">
                <Button size="lg" asChild>
                  <Link href={`/${locale}/register`}>
                    {t('ctaStart')} <ArrowRight className="ml-2 h-4 w-4" />
                  </Link>
                </Button>
                <Button size="lg" variant="ghost" asChild>
                  <Link href="#pricing">{t('ctaPricing')}</Link>
                </Button>
              </div>
              <dl className="mt-16 grid grid-cols-3 overflow-hidden rounded-xl border bg-card text-left shadow-sm">
                <div className="border-r p-5">
                  <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                    {t('stat1Label')}
                  </dt>
                  <dd className="mt-2 text-4xl font-semibold tabular-nums leading-none">
                    25
                  </dd>
                </div>
                <div className="border-r p-5">
                  <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                    {t('stat2Label')}
                  </dt>
                  <dd className="mt-2 text-4xl font-semibold tabular-nums leading-none">
                    14
                  </dd>
                </div>
                <div className="p-5">
                  <dt className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                    {t('stat3Label')}
                  </dt>
                  <dd className="mt-2 text-4xl font-semibold tabular-nums leading-none">
                    ~18h
                  </dd>
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
                {t('featuresTitle')}
              </h2>
              <p className="mt-3 text-muted-foreground">{t('featuresSubtitle')}</p>
            </div>
            <div className="mt-12 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {featureKeys.map((key) => {
                const Icon = featureIcons[key]
                return (
                  <Card key={key}>
                    <CardHeader>
                      <Icon className="h-5 w-5 text-primary" strokeWidth={1.75} />
                      <CardTitle className="mt-4 text-base">
                        {t(`features.${key}.title`)}
                      </CardTitle>
                    </CardHeader>
                    <CardContent>
                      <p className="text-sm text-muted-foreground">
                        {t(`features.${key}.description`)}
                      </p>
                    </CardContent>
                  </Card>
                )
              })}
            </div>
          </div>
        </section>

        {/* HOW */}
        <section id="how" className="border-b bg-muted/30">
          <div className="container py-20">
            <div className="mx-auto max-w-2xl text-center">
              <h2 className="text-3xl font-semibold tracking-tightish md:text-4xl">
                {t('howTitle')}
              </h2>
              <p className="mt-3 text-muted-foreground">{t('howSubtitle')}</p>
            </div>
            <ol className="mt-12 grid gap-8 md:grid-cols-3">
              {stepKeys.map((key, i) => (
                <li key={key} className="relative rounded-lg border bg-card p-6">
                  <div className="font-mono text-xs text-muted-foreground">
                    {t('step')} 0{i + 1}
                  </div>
                  <h3 className="mt-3 text-lg font-semibold">{t(`how.${key}Title`)}</h3>
                  <p className="mt-2 text-sm text-muted-foreground">{t(`how.${key}Body`)}</p>
                </li>
              ))}
            </ol>
          </div>
        </section>

        {/* PRICING */}
        <section id="pricing" className="border-b">
          <div className="container py-20">
            <div className="mx-auto max-w-2xl text-center">
              <h2 className="text-3xl font-semibold tracking-tightish md:text-4xl">
                {t('pricingTitle')}
              </h2>
              <p className="mt-3 text-muted-foreground">{t('pricingSubtitle')}</p>
            </div>
            <div className="mt-12 grid gap-6 md:grid-cols-3">
              {PLANS.map((plan) => {
                const highlighted = plan.highlighted
                return (
                  <Card
                    key={plan.slug}
                    className={highlighted ? 'border-primary shadow-sm ring-1 ring-primary/20' : ''}
                  >
                    <CardHeader>
                      <div className="flex items-center justify-between">
                        <CardTitle className="text-lg">{t(`pricing.${plan.slug}.name`)}</CardTitle>
                        {highlighted && <Badge>{t('recommended')}</Badge>}
                      </div>
                      <div className="mt-4 flex items-baseline gap-1">
                        <span className="text-4xl font-semibold tabular-nums">
                          {formatCurrency(plan.priceUSD, currency, intl)}
                        </span>
                        <span className="text-sm text-muted-foreground">{t('perTest')}</span>
                      </div>
                      <p className="mt-2 text-xs font-medium text-muted-foreground">
                        {plan.testerCount} testers · {plan.durationDays} days
                      </p>
                      <CardDescription className="mt-2">
                        {t(`pricing.${plan.slug}.description`)}
                      </CardDescription>
                    </CardHeader>
                    <CardContent>
                      <ul className="space-y-2.5 text-sm">
                        {[1, 2, 3, 4, 5, 6].map((n) => {
                          const feat = t(`pricing.${plan.slug}.feature${n}`)
                          return (
                            <li key={n} className="flex items-start gap-2">
                              <CheckCircle2
                                className="mt-0.5 h-4 w-4 shrink-0 text-success"
                                strokeWidth={1.75}
                              />
                              <span>{feat}</span>
                            </li>
                          )
                        })}
                      </ul>
                      <Button
                        className="mt-6 w-full"
                        size="lg"
                        variant={highlighted ? 'default' : 'outline'}
                        asChild
                      >
                        <Link href={`/${locale}/register?plan=${plan.slug}`}>{t('choosePlan')}</Link>
                      </Button>
                    </CardContent>
                  </Card>
                )
              })}
            </div>
          </div>
        </section>

        {/* FAQ */}
        <section id="faq" className="border-b bg-muted/30">
          <div className="container py-20">
            <div className="mx-auto max-w-2xl text-center">
              <h2 className="text-3xl font-semibold tracking-tightish md:text-4xl">
                {t('faqTitle')}
              </h2>
            </div>
            <div className="mx-auto mt-12 grid max-w-3xl gap-3">
              {[1, 2, 3, 4, 5, 6].map((n) => (
                <details
                  key={n}
                  className="group rounded-lg border bg-card p-4 [&_summary::-webkit-details-marker]:hidden"
                >
                  <summary className="flex cursor-pointer items-center justify-between gap-4 font-medium">
                    {t(`faqs.q${n}`)}
                    <span
                      aria-hidden
                      className="ml-auto text-muted-foreground transition-transform group-open:rotate-180"
                    >
                      ▾
                    </span>
                  </summary>
                  <p className="mt-3 text-sm text-muted-foreground">{t(`faqs.a${n}`)}</p>
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
                <h3 className="text-sm font-semibold">{t('legalTitle')}</h3>
                <p className="mt-1 text-sm text-muted-foreground">{t('legalBody')}</p>
                <p className="mt-2 text-xs">
                  <Link href={`/${locale}/legal/terms`} className="underline">
                    Terms
                  </Link>{' '}
                  ·{' '}
                  <Link href={`/${locale}/legal/refund`} className="underline">
                    Refund
                  </Link>
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* CTA */}
        <section className="container py-20 text-center">
          <h2 className="text-3xl font-semibold tracking-tightish md:text-4xl">
            {t('ctaFinalTitle')}
          </h2>
          <p className="mx-auto mt-3 max-w-lg text-muted-foreground">
            {t('ctaFinalSubtitle')}
          </p>
          <Button size="lg" className="mt-8" asChild>
            <Link href={`/${locale}/register`}>
              {t('ctaFinal')} <ArrowRight className="ml-2 h-4 w-4" />
            </Link>
          </Button>
        </section>
      </main>

      <SiteFooter locale={locale} />
    </div>
  )
}
