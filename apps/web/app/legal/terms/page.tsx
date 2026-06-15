import type { Metadata } from 'next'

export const metadata: Metadata = { title: 'Kullanım Şartları' }

export default function TermsPage() {
  return (
    <div className="container max-w-3xl py-12">
      <h1 className="text-3xl font-bold">Kullanım Şartları</h1>
      <p className="mt-2 text-sm text-muted-foreground">Son güncelleme: 15 Haziran 2026</p>

      <div className="prose prose-slate dark:prose-invert mt-8 space-y-6 text-sm">
        <section>
          <h2 className="text-xl font-semibold">1. Hizmet Tanımı</h2>
          <p>
            TestersCommunity, Android uygulamalarının Google Play Store kapalı beta test programı kapsamında
            çok sayıda Google hesabı kullanılarak test edilmesini sağlayan bir hizmettir. Müşteri (uygulama sahibi),
            paket adı ve test link bilgilerini platforma girer; hizmet sağlayıcı (işletme) bu testleri 25 farklı
            Google hesabı üzerinden 14 gün boyunca yürütür.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold">2. Yasal Uyarı ve Risk Paylaşımı</h2>
          <p>
            Google Play Console Dağıtım Sözleşmesi ve Google Hizmet Şartları, otomatik indirme, yapay yorum
            artırımı ve çoklu hesap kullanımı gibi faaliyetleri kısıtlayabilir. Bu hizmet bu kuralların sınırında
            hareket etmektedir.
          </p>
          <p>
            <strong>Müşteri,</strong> aşağıdaki riskleri kabul eder:
          </p>
          <ul className="list-disc pl-6 space-y-1">
            <li>Test hesaplarının Google tarafından kapatılması</li>
            <li>Uygulamasının Google Play Store&apos;dan kaldırılması veya yayınlanmasının reddedilmesi</li>
            <li>Müşterinin Google geliştirici hesabının askıya alınması</li>
            <li>Google&apos;ın algoritma değişiklikleri nedeniyle test sonuçlarının olumsuz etkilenmesi</li>
            <li>Yasal yaptırımlar (nadir de olsa)</li>
          </ul>
          <p>
            <strong>Hizmet sağlayıcı,</strong> teknik olarak mümkün olan en iyi sonuçları elde etmek için
            organik görünümlü davranış simülasyonu, yumuşak hesap warming ve rastgeleleştirme tekniklerini
            uygular. Ancak Google&apos;ın tespit sistemlerini atlatma garantisi vermez.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold">3. Yasaklı Kullanım</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>Yetişkin içerikli uygulamalar</li>
            <li>Kumar, bahis uygulamaları</li>
            <li>Zararlı yazılım (malware) içeren uygulamalar</li>
            <li>Yasa dışı faaliyetlere yönlendiren uygulamalar</li>
            <li>Spam veya dolandırıcılık amaçlı uygulamalar</li>
            <li>Başkalarının fikri mülkiyet haklarını ihlal eden uygulamalar</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold">4. Sorumluluk Sınırı</h2>
          <p>
            Hizmet sağlayıcı, hizmetin kesintiye uğraması, Google tarafından engellenmesi veya beklenen sonuçların
            alınamaması durumunda müşterinin uğradığı dolaylı zararlardan (itibar kaybı, gelir kaybı vb.)
            sorumlu değildir. Müşterinin azami talep hakkı, ödediği ücretle sınırlıdır.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold">5. Veri Saklama</h2>
          <p>
            Müşteri verileri Türkiye&apos;deki sunucularda saklanır. KVKK ve GDPR kapsamında müşteri, kişisel
            verilerinin silinmesini talep edebilir. Test verileri (screenshot, aktivite logları) hizmet
            tamamlandıktan 30 gün sonra otomatik silinir.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold">6. Uyuşmazlık Çözümü</h2>
          <p>
            Bu sözleşmeden doğan uyuşmazlıklarda Türkiye Cumhuriyeti mahkemeleri ve icra daireleri yetkilidir.
            İstanbul mahkemeleri münhasır yargı yetkisine sahiptir.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold">7. İletişim</h2>
          <p>
            Sorularınız için: <a href="mailto:legal@testerscomm.net" className="text-primary underline">legal@testerscomm.net</a>
          </p>
        </section>
      </div>
    </div>
  )
}
