import { NextIntlClientProvider } from 'next-intl'
import { getMessages, getTranslations } from 'next-intl/server'
import { notFound, redirect } from 'next/navigation'
import { setRequestLocale } from 'next-intl/server'
import { SUPPORTED_LOCALES, type Locale, BRAND } from '@/lib/brand'

import '@/app/globals.css'

export async function generateStaticParams() {
  return SUPPORTED_LOCALES.map((locale) => ({ locale }))
}

export async function generateMetadata({ params }: { params: Promise<{ locale: Locale }> }) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'common' })
  return {
    title: {
      default: `${BRAND.name} — ${t('tagline')}`,
      template: `%s | ${BRAND.name}`,
    },
    description: BRAND.shortDescription,
    authors: [{ name: BRAND.name }],
    creator: BRAND.name,
    openGraph: {
      siteName: BRAND.name,
      title: `${BRAND.name} — ${t('tagline')}`,
      description: BRAND.shortDescription,
      url: BRAND.web,
      locale: locale,
      type: 'website',
    },
  }
}

export default async function LocaleLayout({
  children,
  params,
}: {
  children: React.ReactNode
  params: Promise<{ locale: string }>
}) {
  const { locale } = await params
  if (!SUPPORTED_LOCALES.includes(locale as Locale)) {
    notFound()
  }
  setRequestLocale(locale)
  const messages = await getMessages({ locale })

  return (
    <html lang={locale} suppressHydrationWarning>
      <body className="min-h-screen bg-background text-foreground antialiased">
        <NextIntlClientProvider locale={locale} messages={messages}>
          {children}
        </NextIntlClientProvider>
      </body>
    </html>
  )
}
