import { getRequestConfig } from 'next-intl/server'
import { notFound } from 'next/navigation'
import { SUPPORTED_LOCALES, type Locale, DEFAULT_LOCALE } from './lib/brand'

export default getRequestConfig(async ({ requestLocale }) => {
  let locale = await requestLocale
  if (!locale || !SUPPORTED_LOCALES.includes(locale as Locale)) {
    locale = DEFAULT_LOCALE
  }
  return {
    locale,
    messages: (await import(`./messages/${locale}.json`)).default,
  }
})
