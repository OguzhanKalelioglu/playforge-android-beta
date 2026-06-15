import type { Metadata } from 'next'

export const metadata: Metadata = { title: 'İade Politikası' }

export default function RefundPage() {
  return (
    <div className="container max-w-3xl py-12">
      <h1 className="text-3xl font-bold">İade Politikası</h1>
      <p className="mt-2 text-sm text-muted-foreground">Son güncelleme: 15 Haziran 2026</p>

      <div className="prose prose-slate dark:prose-invert mt-8 space-y-6 text-sm">
        <section>
          <h2 className="text-xl font-semibold">1. İade Koşulları</h2>
          <div className="space-y-3">
            <div className="rounded-lg border p-4">
              <h3 className="font-semibold text-green-700 dark:text-green-400">%100 İade</h3>
              <p className="mt-1">Test henüz başlamadıysa (hesap ataması yapılmadıysa) tam iade yapılır.</p>
            </div>
            <div className="rounded-lg border p-4">
              <h3 className="font-semibold text-yellow-700 dark:text-yellow-400">%50 İade</h3>
              <p className="mt-1">
                Test başladıysa ve 7 günden az süre geçtiyse, ödenen tutarın yarısı iade edilir.
              </p>
            </div>
            <div className="rounded-lg border p-4">
              <h3 className="font-semibold text-red-700 dark:text-red-400">İade Yok</h3>
              <p className="mt-1">
                Test başladıktan 7 gün sonra iade yapılmaz. 14 günlük test süreci başlamışsa,
                hizmet kullanılmış sayılır.
              </p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-xl font-semibold">2. İade Talebi Nasıl Yapılır?</h2>
          <ol className="list-decimal pl-6 space-y-1">
            <li>
              <a href="mailto:refund@testerscomm.net" className="text-primary underline">refund@testerscomm.net</a>{' '}
              adresine e-posta gönderin
            </li>
            <li>Konu: &quot;İade Talebi - [Test ID]&quot;</li>
            <li>İçerik: Test ID, sipariş tarihi, iade gerekçesi</li>
            <li>Talep 3 iş günü içinde değerlendirilir</li>
            <li>Onaylanan iade, 5-10 iş günü içinde karta geri yansır</li>
          </ol>
        </section>

        <section>
          <h2 className="text-xl font-semibold">3. AB Müşterileri İçin 14 Günlük Cayma Hakkı</h2>
          <p>
            AB/AEA ülkelerinden alışveriş yapan müşteriler, hizmetin başlamadığı durumlarda 14 günlük
            cayma hakkına sahiptir. Test başladıktan sonra bu hak kullanılamaz (hizmet ifasına başlanmıştır).
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold">4. İade Edilemeyecek Durumlar</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>Test başarıyla tamamlandıktan sonra (hizmet ifa edildi)</li>
            <li>Yasaklı içerik tespit edilmesi (kullanım şartları ihlali)</li>
            <li>Sahte bilgi ile kayıt yapılması</li>
            <li>Hizmetin kötüye kullanılması</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold">5. İletişim</h2>
          <p>
            İade soruları: <a href="mailto:refund@testerscomm.net" className="text-primary underline">refund@testerscomm.net</a>
          </p>
        </section>
      </div>
    </div>
  )
}
