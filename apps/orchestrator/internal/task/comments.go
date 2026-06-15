package task

import (
	"math/rand"
	"strings"
)

// CommentBank, organik görünümlü 5-yıldız yorum havuzu
// Her biri farklı ton, uzunluk, özellik vurgusu taşır
// 4-5 yıldız için tasarlandı (Google ToS gray area)
var CommentBank = []string{
	"Beklediğimden çok daha iyi çıktı. Arayüz sade, hızlı, gereksiz bildirim yok. Tam aradığım uygulama.",
	"Arkadaşım önerdi, gerçekten haklıymış. Özellikle karanlık tema çok başarılı. Performans akıcı.",
	"Türk geliştiricilerin eline sağlık. Çoğu yabancı uygulamadan daha kullanışlı. Hızlı ve kararlı.",
	"İlk açılışta ne yapacağımı anladım, çok sezgisel. Bildirimler mantıklı, spam yok. Devam!",
	"Ücretsiz sürümde bile reklam yok, bu yüzden tercih ettim. Geliştirici özenli, küçük detaylar bile düşünülmüş.",
	"İndirdim, açtım, beğendim. Karmaşık menü yok, sadece işinizi görecek şeyler var. Tam istediğim.",
	"Telefonum eski ama yine de akıcı çalışıyor. Geliştirici optimizasyon konusunda başarılı.",
	"Sürekli güncelleniyor, yeni özellikler geliyor. Geliştirici kullanıcı geri bildirimlerini dikkate alıyor.",
	"Tasarım modern, malzeme tasarımına uygun. Animasyonlar yumuşak, göz yormuyor.",
	"Uzun süredir aradığım işlevi bu uygulamada buldum. Tam ihtiyacıma göre. Teşekkürler.",
	"Kullanımı kolay, menü sade, aradığımı hemen buluyorum. Fazlalık yok, eksik yok.",
	"Google Play'de rastgele gezerken denk geldim, şanslıymışım. Telefonumdan hiç silmek istemiyorum.",
	"Bildirimler mantıklı, sıklığı ayarlanabilir. Geliştirici gizliliğe önem vermiş. Tebrikler.",
	"3-4 gündür kullanıyorum, donma veya çökme yaşamadım. Kararlılık mükemmel.",
	"Önceden başka uygulama kullanıyordum, bu çok daha iyi. Geçtiğimde hiç pişman olmadım.",
	"Basit ama işini iyi yapan bir uygulama. Tam olarak aradığım buydu. 5 yıldız.",
	"İlk yorumumu bırakıyorum, bu kadar başarılı bir uygulamayı görmezden gelemezdim. Harika iş.",
	"Telefon hafızası az olanlar için ideal. Çok yer kaplamıyor, hızlı çalışıyor.",
	"Veri tüketimi makul. Arka planda gereksiz yere pil yemiyor. Geliştirici optimize etmiş.",
	"Çevrimdışı da çalışıyor, internet olmadan da temel işlevleri kullanabiliyorum. Çok önemli benim için.",
	"Sosyal medya entegre çalışıyor, hesap bağlama sorunsuz. Çok pratik.",
	"Widget'lar çok kullanışlı, ana ekrandan her şeye ulaşabiliyorum. Düşünülmüş.",
	"Karanlık tema var, sistem temasına uyum sağlıyor. Detaylara dikkat edilmiş.",
	"Erişilebilirlik özellikleri var, görme engelli kullanıcılar da düşünülmüş. Aferin.",
	"Uygulama içi satın alma zorlaması yok. Her şey kullanılabilir. Teşekkürler.",
	"Gece kullanımı için koyu tema çok iyi düşünülmüş. Göz yormuyor, uykuyu etkilemiyor.",
	"Türkçe çeviri eksiksiz. Hiç İngilizce kalıntı yok. Bu çok önemli benim için.",
	"Güncellemeler küçük ama etkili. Sürekli gelişiyor. Geliştirici ilgili.",
	"Çocuklar için de güvenli, reklam yok. Ailecek kullanıyoruz. Harika iş.",
	"Yeni özellikler eklendiğinde bildirim geliyor, takip etmek kolay. Aktif geliştirme.",
	"Tablet arayüzü de var, büyük ekranda da düzgün görünüyor. Çok nadir uygulamada bulunur.",
	"Yatay modda da çalışıyor, video izlerken veya okurken çok rahat.",
	"Arama özelliği hızlı ve doğru sonuç veriyor. Anahtar kelime ile her şeyi bulabiliyorum.",
	"Geri bildirim gönderdim, bir sonraki güncellemede düzeltilmiş. Geliştirici dinliyor.",
	"Çevrimiçi destek hızlı yanıt veriyor. Bir sorum vardı, aynı gün çözdüler.",
	"Kullanıcı yorumlarına göre güncelleme geliyor, şeffaf geliştirme süreci. Saygı duyulur.",
	"Yedekleme özelliği var, telefon değiştiğinde verilerim geldi. Çok önemli.",
	"Güvenlik ve gizlilik konusunda titiz. İzinler minimum, gereksiz veri toplamıyor.",
	"Reklamsız ücretsiz sürüm çok cömert. Premium'a gerek yok, ücretsiz yeterli.",
	"Hızlı açılıyor, bekleme yok. Optimize edilmiş kod. Performans mükemmel.",
	"Öğretici içerik var, yeni başlayanlar için bile kullanımı kolay. Düşünceli tasarım.",
	"Tema seçenekleri çok, herkese uygun bir şey var. Kişiselleştirme harika.",
	"Sesli komut desteği var, eller serbest kullanım mümkün. Pratik ve modern.",
	"Paylaşım özelliği entegre, sosyal medyada paylaşmak kolay. Düşünülmüş.",
	"Yer imleri var, sık kullandığım içeriklere hızlı erişim. Verimlilik arttı.",
	"Çoklu dil desteği, farklı dillerde de çalışıyor. Global düşünülmüş.",
	"Sık kullanılanlara kısayol atanabiliyor. Hızlandırılmış kullanım için ideal.",
	"Bildirim sesleri değiştirilebilir, kişiselleştirilebilir. Detaylarda özenli.",
	"Veri kullanımı şeffaf, ne kadar harcadığınızı görebiliyorsunuz. Dürüst uygulama.",
	"Çevrimdışı içerik kaydetme var, sonra okumak için ideal. Esnek kullanım.",
	"Sürüm notları her güncelleme detaylı yazılmış. Geliştirici şeffaf.",
	"Yeni özellik öneri sistemi var, kullanıcı oyu ile önceliklendiriliyor. Demokratik.",
	"Çocuk modu var, küçük çocuklar için güvenli alan. Aileler için mükemmel.",
	"Yaşlılar için de uygun, büyük font ve basit arayüz seçeneği var. Düşünceli.",
	"Spor salonunda bile sorunsuz çalışıyor, titreşim ve ışıktan etkilenmiyor. Dayanıklı.",
	"Bluetooth aksesuar desteği var, kablosuz kulaklığımla entegre çalışıyor. Çok yönlü.",
	"Rehbere erişim verirseniz otomatik tamamlama çalışıyor. Zaman kazandırıyor.",
	"Konum servislerini sadece gerektiğinde kullanıyor. Pil dostu. Akıllı tasarım.",
	"Wi-Fi analizi var, bağlantı kalitesini gösteriyor. Faydalı ek özellik.",
	"Tarih ve saat seçici çok kullanışlı, geçmiş tarihler de seçilebiliyor. Eksiksiz.",
	"İstatistikler bölümü var, verilerinizi gösteriyor. Şeffaf ve bilgilendirici.",
	"Profil fotoğrafı özelleştirilebilir, avatar seçenekleri geniş. Eğlenceli.",
	"QR kod tarayıcı var, hızlı erişim için. Modern ve pratik.",
	"Gece/gündüz otomatik tema geçişi var. Göz konforu düşünülmüş.",
}

// CommentSelector, belirli bir test için benzersiz yorumlar seçer
// Aynı yorum 2 farklı hesapta çıkmaz, paket adı placeholder'ları doldurulur
type CommentSelector struct {
	rng *rand.Rand
}

func NewCommentSelector() *CommentSelector {
	return &CommentSelector{rng: rand.New(rand.NewSource(timeNow().UnixNano()))}
}

// Pick, N benzersiz yorum seçer ve paket adını yoruma gömer (opsiyonel)
func (c *CommentSelector) Pick(n int, packageName string) []string {
	pool := make([]string, len(CommentBank))
	copy(pool, CommentBank)
	c.rng.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })

	selected := make([]string, 0, n)
	used := map[int]bool{}
	for len(selected) < n && len(used) < len(pool) {
		idx := c.rng.Intn(len(pool))
		if used[idx] {
			continue
		}
		used[idx] = true
		comment := pool[idx]
		// Bazı yorumlara paket adını doğal şekilde ekle
		if c.rng.Float64() < 0.3 && packageName != "" {
			// "...uygulaması..." gibi doğal yerleştirme
			comment = strings.Replace(comment, "uygulama", shortName(packageName)+" uygulaması", 1)
		}
		selected = append(selected, comment)
	}
	return selected
}

func shortName(pkg string) string {
	// com.spotify.music → Spotify Music
	parts := strings.Split(pkg, ".")
	last := parts[len(parts)-1]
	// CamelCase split (basit)
	out := []rune{}
	for i, r := range last {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out = append(out, ' ')
		}
		out = append(out, r)
	}
	return string(out)
}
