import type { Metadata } from 'next'
import { getTranslations } from 'next-intl/server'
import { BRAND, type Locale } from '@/lib/brand'

export async function generateMetadata({ params }: { params: Promise<{ locale: Locale }> }) {
  const { locale } = await params
  const t = await getTranslations({ locale, namespace: 'legal' })
  return { title: t('termsTitle') }
}

export default async function TermsPage({ params }: { params: Promise<{ locale: Locale }> }) {
  const { locale } = await params
  return (
    <div className="container max-w-3xl py-12">
      <h1 className="text-3xl font-bold">
        {locale === 'tr' ? 'Kullanım Şartları' : 'Terms of Service'}
      </h1>
      <p className="mt-2 text-sm text-muted-foreground">
        {locale === 'tr' ? 'Son güncelleme: 15 Haziran 2026' : 'Last updated: June 15, 2026'}
      </p>

      <div className="prose prose-slate dark:prose-invert mt-8 space-y-6 text-sm">
        {locale === 'tr' ? (
          <>
            <section>
              <h2 className="text-xl font-semibold">1. Hizmet Tanımı</h2>
              <p>
                {BRAND.name}, Android uygulamalarının Google Play Store kapalı beta test programı kapsamında
                çok sayıda Google hesabı kullanılarak test edilmesini sağlayan bir hizmettir.
              </p>
            </section>
            <section>
              <h2 className="text-xl font-semibold">2. Yasal Uyarı ve Risk Paylaşımı</h2>
              <p>
                Google Play Console Dağıtım Sözleşmesi ve Google Hizmet Şartları, otomatik indirme, yapay
                yorum artırımı ve çoklu hesap kullanımı gibi faaliyetleri kısıtlayabilir. Bu hizmet bu
                kuralların sınırında hareket etmektedir.
              </p>
              <p><strong>Müşteri</strong> aşağıdaki riskleri kabul eder:</p>
              <ul className="list-disc pl-6 space-y-1">
                <li>Test hesaplarının Google tarafından kapatılması</li>
                <li>Uygulamasının Google Play Store&apos;dan kaldırılması</li>
                <li>Müşterinin Google geliştirici hesabının askıya alınması</li>
                <li>Google&apos;ın algoritma değişiklikleri nedeniyle test sonuçlarının olumsuz etkilenmesi</li>
              </ul>
            </section>
            <section>
              <h2 className="text-xl font-semibold">3. Yasaklı Kullanım</h2>
              <ul className="list-disc pl-6 space-y-1">
                <li>Yetişkin içerikli uygulamalar</li>
                <li>Kumar, bahis uygulamaları</li>
                <li>Zararlı yazılım (malware) içeren uygulamalar</li>
                <li>Yasa dışı faaliyetlere yönlendiren uygulamalar</li>
                <li>Spam veya dolandırıcılık amaçlı uygulamalar</li>
              </ul>
            </section>
            <section>
              <h2 className="text-xl font-semibold">4. Sorumluluk Sınırı</h2>
              <p>
                Hizmet sağlayıcı, hizmetin kesintiye uğraması, Google tarafından engellenmesi veya beklenen
                sonuçların alınamaması durumunda dolaylı zararlardan sorumlu değildir. Azami talep hakkı
                ödenen ücretle sınırlıdır.
              </p>
            </section>
            <section>
              <h2 className="text-xl font-semibold">5. Veri Saklama</h2>
              <p>
                Müşteri verileri Türkiye&apos;deki sunucularda saklanır. KVKK ve GDPR kapsamında müşteri,
                kişisel verilerinin silinmesini talep edebilir. Test verileri hizmet tamamlandıktan 30 gün
                sonra otomatik silinir.
              </p>
            </section>
            <section>
              <h2 className="text-xl font-semibold">6. İletişim</h2>
              <p>
                Sorularınız için:{' '}
                <a href={`mailto:${BRAND.email.legal}`} className="text-primary underline">
                  {BRAND.email.legal}
                </a>
              </p>
            </section>
          </>
        ) : (
          <>
            <section>
              <h2 className="text-xl font-semibold">1. Service Description</h2>
              <p>
                {BRAND.name} is a service that enables Android apps to be tested through Google Play
                Store&apos;s closed beta testing program using multiple Google accounts. The customer (app
                owner) enters the package name and test link; the service provider runs these tests through
                25 different Google accounts over 14 days.
              </p>
            </section>
            <section>
              <h2 className="text-xl font-semibold">2. Legal Notice and Risk Sharing</h2>
              <p>
                The Google Play Console Distribution Agreement and Google Terms of Service may restrict
                automated downloads, artificial review inflation, and multi-account use. This service
                operates at the edge of these rules.
              </p>
              <p><strong>The customer</strong> accepts the following risks:</p>
              <ul className="list-disc pl-6 space-y-1">
                <li>Google closing test accounts</li>
                <li>App being removed or rejected from the Google Play Store</li>
                <li>Customer&apos;s Google developer account being suspended</li>
                <li>Test results being adversely affected by Google&apos;s algorithm changes</li>
              </ul>
            </section>
            <section>
              <h2 className="text-xl font-semibold">3. Prohibited Use</h2>
              <ul className="list-disc pl-6 space-y-1">
                <li>Adult content applications</li>
                <li>Gambling or betting applications</li>
                <li>Malware-containing applications</li>
                <li>Apps leading to illegal activities</li>
                <li>Spam or fraud applications</li>
                <li>Apps violating others&apos; intellectual property</li>
              </ul>
            </section>
            <section>
              <h2 className="text-xl font-semibold">4. Limitation of Liability</h2>
              <p>
                The service provider is not liable for indirect damages (reputation loss, lost revenue, etc.)
                if the service is interrupted, blocked by Google, or expected results are not achieved. The
                customer&apos;s maximum claim is limited to the fee paid.
              </p>
            </section>
            <section>
              <h2 className="text-xl font-semibold">5. Data Retention</h2>
              <p>
                Customer data is stored on servers in the EU/Turkey. Under GDPR/KVKK, the customer may
                request deletion of personal data. Test data (screenshots, activity logs) is automatically
                deleted 30 days after service completion.
              </p>
            </section>
            <section>
              <h2 className="text-xl font-semibold">6. Contact</h2>
              <p>
                Questions:{' '}
                <a href={`mailto:${BRAND.email.legal}`} className="text-primary underline">
                  {BRAND.email.legal}
                </a>
              </p>
            </section>
          </>
        )}
      </div>
    </div>
  )
}
