import Link from 'next/link'
import { useTranslations } from 'next-intl'
import { BRAND, type Locale } from '@/lib/brand'

interface SiteFooterProps {
  locale: Locale
}

export function SiteFooter({ locale }: SiteFooterProps) {
  const t = useTranslations('footer')

  return (
    <footer className="border-t bg-card">
      <div className="container py-12">
        <div className="grid gap-8 md:grid-cols-4">
          <div className="md:col-span-2">
            <div className="flex items-center gap-2">
              <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground text-sm font-semibold">
                P
              </div>
              <span className="text-sm font-semibold tracking-tightish">{BRAND.name}</span>
            </div>
            <p className="mt-3 max-w-sm text-sm text-muted-foreground">{t('tagline')}</p>
          </div>

          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              {t('product')}
            </h3>
            <ul className="mt-3 space-y-2 text-sm">
              <li>
                <Link href={`/${locale}#features`} className="hover:text-foreground">
                  {t('productFeatures')}
                </Link>
              </li>
              <li>
                <Link href={`/${locale}#how`} className="hover:text-foreground">
                  {t('productHow')}
                </Link>
              </li>
              <li>
                <Link href={`/${locale}#pricing`} className="hover:text-foreground">
                  {t('productPricing')}
                </Link>
              </li>
            </ul>
          </div>

          <div>
            <h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              {t('legal')}
            </h3>
            <ul className="mt-3 space-y-2 text-sm">
              <li>
                <Link href={`/${locale}/legal/terms`} className="hover:text-foreground">
                  {t('legalTerms')}
                </Link>
              </li>
              <li>
                <Link href={`/${locale}/legal/privacy`} className="hover:text-foreground">
                  {t('legalPrivacy')}
                </Link>
              </li>
              <li>
                <Link href={`/${locale}/legal/refund`} className="hover:text-foreground">
                  {t('legalRefund')}
                </Link>
              </li>
            </ul>
            <h3 className="mt-6 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              {t('contact')}
            </h3>
            <ul className="mt-3 space-y-2 text-sm">
              <li>
                <Link href={`mailto:${BRAND.email.support}`} className="hover:text-foreground">
                  {t('emailUs')}
                </Link>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-10 flex flex-col items-start justify-between gap-2 border-t pt-6 text-xs text-muted-foreground sm:flex-row">
          <span>{t('copyright', { year: new Date().getFullYear() })}</span>
          <span>{BRAND.domain}</span>
        </div>
      </div>
    </footer>
  )
}
