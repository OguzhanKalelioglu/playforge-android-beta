import type { Metadata } from 'next'

export const metadata: Metadata = { title: 'Gizlilik Politikası' }

export default function PrivacyPage() {
  return (
    <div className="container max-w-3xl py-12">
      <h1 className="text-3xl font-bold">Gizlilik Politikası</h1>
      <p className="mt-2 text-sm text-muted-foreground">Son güncelleme: 15 Haziran 2026</p>

      <div className="prose prose-slate dark:prose-invert mt-8 space-y-6 text-sm">
        <section>
          <h2 className="text-xl font-semibold">1. Toplanan Veriler</h2>
          <p>Platforma kayıt olduğunuzda aşağıdaki veriler toplanır:</p>
          <ul className="list-disc pl-6 space-y-1">
            <li>Ad, soyad</li>
            <li>E-posta adresi</li>
            <li>Şifre (bcrypt ile hash&apos;lenmiş, düz metin saklanmaz)</li>
            <li>IP adresi (güvenlik ve log amaçlı)</li>
            <li>Tarayıcı bilgisi (user agent)</li>
          </ul>
          <p>Test hizmeti kapsamında:</p>
          <ul className="list-disc pl-6 space-y-1">
            <li>Uygulama paket adı</li>
            <li>Test linki</li>
            <li>Ödeme bilgileri (Iyzico altyapısında, bizde saklanmaz)</li>
            <li>Test çıktıları (aktivite logları, screenshotlar)</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold">2. Verilerin Kullanım Amacı</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>Hizmetin sunulması ve test sürecinin yönetimi</li>
            <li>Ödeme işlemlerinin gerçekleştirilmesi</li>
            <li>Size özel test raporlarının oluşturulması</li>
            <li>Yasal yükümlülüklerin yerine getirilmesi (fatura, vergi)</li>
            <li>Sahtecilik (fraud) tespiti ve önlenmesi</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold">3. Üçüncü Taraf Hizmetler</h2>
          <p>Verileriniz aşağıdaki üçüncü taraf hizmetlerle paylaşılır:</p>
          <ul className="list-disc pl-6 space-y-1">
            <li><strong>Iyzico:</strong> Ödeme işlemleri (KVKK uyumlu, Türkiye)</li>
            <li><strong>Hetzner:</strong> Sunucu barındırma (Almanya, GDPR uyumlu)</li>
            <li><strong>Backblaze B2:</strong> Yedekleme (ABD)</li>
            <li><strong>Google Cloud:</strong> E-posta ve log hizmetleri (opsiyonel)</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold">4. KVKK Kapsamındaki Haklarınız</h2>
          <p>6698 sayılı KVKK uyarınca aşağıdaki haklara sahipsiniz:</p>
          <ul className="list-disc pl-6 space-y-1">
            <li>Kişisel verilerinizin işlenip işlenmediğini öğrenme</li>
            <li>İşlenmişse buna ilişkin bilgi talep etme</li>
            <li>İşlenme amacını ve amacına uygun kullanılıp kullanılmadığını öğrenme</li>
            <li>Yurt içinde/dışında aktarıldığı üçüncü kişileri öğrenme</li>
            <li>Eksik/yanlış işlenen verilerin düzeltilmesini isteme</li>
            <li>Şartlar oluştuğunda silinmesini/yok edilmesini isteme</li>
            <li>Otomatik sistemlerle aleyhine sonuç doğan analizlere itiraz etme</li>
          </ul>
          <p>
            Bu haklarınızı kullanmak için{' '}
            <a href="mailto:kvkk@testerscomm.net" className="text-primary underline">kvkk@testerscomm.net</a>{' '}
            adresine yazılı talep gönderin. 30 gün içinde yanıtlanır.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold">5. Çerezler</h2>
          <p>
            Platform, oturum yönetimi için zorunlu çerezler (httpOnly JWT) kullanır. Pazarlama veya analitik
            çerezi kullanılmaz.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold">6. Veri Saklama Süreleri</h2>
          <ul className="list-disc pl-6 space-y-1">
            <li>Hesap bilgileri: hesap silinene kadar</li>
            <li>Test verileri: test tamamlandıktan 30 gün sonra</li>
            <li>Ödeme kayıtları: 5 yıl (vergi mevzuatı)</li>
            <li>Yedeklemeler: 30 gün</li>
          </ul>
        </section>

        <section>
          <h2 className="text-xl font-semibold">7. İletişim</h2>
          <p>
            Veri sorumlusu iletişim: <a href="mailto:kvkk@testerscomm.net" className="text-primary underline">kvkk@testerscomm.net</a>
          </p>
        </section>
      </div>
    </div>
  )
}
