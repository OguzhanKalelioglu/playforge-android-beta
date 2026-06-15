import createMiddleware from 'next-intl/middleware'
import { SUPPORTED_LOCALES, DEFAULT_LOCALE } from './lib/brand'

export default createMiddleware({
  locales: [...SUPPORTED_LOCALES],
  defaultLocale: DEFAULT_LOCALE,
  localePrefix: 'always',
})

export const config = {
  matcher: ['/((?!api|_next|_vercel|.*\\..*).*)'],
}
