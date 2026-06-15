# Google Hesap Açma Prosedürü (Manuel Checklist)

> **Yasal Uyarı:** Google ToS'a göre otomatik hesap açma yasaktır. Bu prosedür **tamamen manuel** uygulanmalıdır. Günde 1-2 hesap, 12-15 günde 25'e ulaşılır.

## Genel Kurallar

- ✅ Her hesap için **farklı kişilik** oluştur (isim, doğum tarihi, profil fotoğrafı)
- ✅ Aynı cihaz/IP'den günde **max 1-2 hesap** aç
- ✅ **2FA AÇMA** (test sürecini zorlaştırır)
- ✅ Hesap açıldıktan sonra **24-72 saat warming** dönemi
- ✅ Warming döneminde hafif aktivite (email oku, YouTube, Search)

## Her Hesap İçin Adımlar

### Hazırlık

1. Bu checklist'i doldur (şifreli dosyada sakla):
```
# Tester #___
- Email:          _________________@gmail.com
- Password:       _________________ (32+ karakter)
- Name:           _________________
- Birthdate:      ____-__-__
- Gender:         ___________
- Recovery email: _________________@proton.me
- Phone:          +90___________
- Profile photo:  thispersondoesnotexist.com'dan indir, kaydet
- Açılış tarihi:  ____-__-__
- Warming bitti:  ____-__-__
- Durum:          [ ] warming  [ ] active  [ ] cooling  [ ] disabled
```

2. Mac Mini'de **Safari** veya **Firefox** aç (Chrome değil)
3. **VPN KAPALI** olmalı (kendi IP'n kullanılacak)
4. Tarayıcı geçmişini temizle (veya Incognito/Private mode)

### Hesap Açma (accounts.google.com/signup)

1. https://accounts.google.com/signup adresine git
2. Ad, soyad gir (yukarıdaki kişiliğe göre)
3. Cinsiyet seç (karışık: yarısı kadın, yarısı erkek)
4. Doğum tarihi gir (18-30 yaş arası, farklı günler)
5. **Kendi e-posta adresi** kısmına yeni Gmail adresi yaz
6. Şifre oluştur (32+ karakter, unique)
7. Telefon numarası:
   - **İlk 5 hesap:** Kendi numaranı kullan, son 2 haneyi değiştir
   - **Sonraki 20 hesap:** SMS-Activate.org'dan numara kirala (~$0.10-0.50)
8. SMS kodu gelince gir
9. "İfadeleri görmesine izin ver" → varsayılan bırak
10. **"Daha az kişiselleştirilmiş reklamlar"** seçeneğini işaretle
11. **KVKK/GDPR onayı** ver

### Hesap Açıldıktan Sonra

1. **Profil fotoğrafı** yükle (thispersondoesnotexist.com'dan generate et, downloads klasörüne kaydet)
2. **Recovery email** ekle (ProtonMail hesabın)
3. **2FA AÇMA** (test otomasyonu bozar)
4. **Play Store**'a giriş yap
5. **Ödeme yöntemi ekleme** (gerek yok)
6. 24 saat bekle (warming dönemi)

### Warming Dönemi (3 gün)

Her gün 5-10 dakika:

- [ ] Gmail'de 2-3 email oku (spam'e düşenler dahil)
- [ ] YouTube'da 1-2 video izle (herhangi bir konu)
- [ ] Google Search'te 2-3 arama yap
- [ ] Google Maps'te bir yere bak
- [ ] Google News'te bir haber oku

Bu aktiviteler "bot değilim" sinyali verir.

### 3. Gün Sonunda

1. Play Store'a giriş yapabiliyor musun kontrol et
2. Bir uygulama ara (örn: WhatsApp) → arama sonucu geliyor mu
3. Herhangi bir uygulamayı **test amaçlı indir**, 5 saniye aç, sil
4. Bu "gerçek kullanıcı" kanıtı oluşturur

### PostgreSQL'e Kayıt

```sql
-- Encryption key: DB_ENCRYPTION_KEY environment variable
INSERT INTO testers (
    email,
    password_encrypted,
    recovery_email,
    phone,
    status,
    notes
) VALUES (
    'tester001.realname@gmail.com',
    pgp_sym_encrypt('şifre-buraya', current_setting('app.encryption_key')),
    'protonmail001@proton.me',
    '+90xxxxxxxxx',
    'active',
    '2026-06-XX açıldı, 2026-06-XX warming bitti'
);
```

## Google Groups Oluşturma (Her Test İçin)

1. https://groups.google.com adresine git
2. "Grup oluştur"
3. Grup adı: `test-<test_id>-<customer-slug>@googlegroups.com`
4. Ayarlar:
   - **Üyelik:** "Yalnızca davet edilen kullanıcılar"
   - **Yeni üyeleri onayla:** "Otomatik onayla"
   - **Üyelerin birbirini görmesi:** Hayır
5. **25 tester email'ini** yapıştır (virgülle ayrılmış)
6. **Müşterinin geliştirici email'ini** de ekle
7. Kaydet

**Not:** Google bazen 25 üyeyi tek seferde kabul etmeyebilir. Günde 5-10 hesap ekleyerek ilerle.

## Toplu Yapılacaklar

Haftalık kontrol:

- [ ] Yeni hesap açıldı mı? (günde 1-2)
- [ ] Warming dönemi tamamlanan hesap var mı?
- [ ] Ban yemiş hesap var mı? (login dene)
- [ ] Yeni Google Groups oluşturuldu mu?

## Ban Tespiti

Bir hesap banlendiğinde şu belirtiler olur:

- **Soft ban:** Play Store'a giriş yapamaz, ama Gmail çalışır
- **Hard ban:** Tüm Google servislerinden atılır, "suspended" mesajı
- **Phone verification:** Girişte sürekli SMS kodu ister

Tespit edildiğinde:

```sql
UPDATE testers
SET status = 'disabled',
    notes = COALESCE(notes, '') || E'\n[2026-XX-XX] Disabled: Google suspended'
WHERE id = '<tester_id>';
```

Yeni hesap açarak 25'i tamamla.
