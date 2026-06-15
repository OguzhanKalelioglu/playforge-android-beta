package profile

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/google/uuid"
)

// DeviceProfile, bir Google hesabı için cihaz parmak izi
// Anti-detect için her hesaba farklı profil atanır
type DeviceProfile struct {
	ID              string
	AndroidID       string // Settings.Secure.ANDROID_ID
	IMEI            string
	MacAddress      string
	Model           string
	Manufacturer    string
	AndroidVersion  string
	BuildNumber     string
	ScreenRes       string
	Density         string
	UserAgent       string
	Locale          string // tr-TR, en-US
	Timezone        string // Europe/Istanbul
	Telephony       string // operator ismi
	GLVendor        string
	GLRenderer      string
	CPU_ABI         string
	NetworkType     string // wifi, mobile
	SIMOperator     string
	Keyboard        string
	FontsHash       string
	BatteryLevel    int
	SignalStrength  int
	AppsInstalled   []string // fake installed list
}

// 10 farklı cihaz profili seed listesi (gerçek model isimleri)
var baseModels = []struct {
	Model        string
	Manufacturer string
	Android      string
	Build        string
	Screen       string
	Density      string
	CPU          string
}{
	{"SM-A525F", "samsung", "13", "A525FXXS5DWL1", "1080x2400", "420", "arm64-v8a"},
	{"Mi 11 Lite", "Xiaomi", "12", "RKQ1.200826.002", "1080x2400", "402", "arm64-v8a"},
	{"CPH2451", "OPPO", "13", "CPH2451_13.0.1", "1080x2412", "409", "arm64-v8a"},
	{"M2102J20SG", "Xiaomi", "12", "QKQ1.200826.002", "1080x2400", "395", "arm64-v8a"},
	{"NE2213", "OnePlus", "13", "NE2213_13.0.0", "1080x2412", "411", "arm64-v8a"},
	{"motorola edge 30", "motorola", "12", "S3RXS32.50-23-7", "1080x2400", "405", "arm64-v8a"},
	{"Pixel 6a", "Google", "14", "TQ3A.230901.001", "1080x2400", "429", "arm64-v8a"},
	{"RMX3085", "realme", "12", "RMX3085_11_C.12", "1080x2400", "409", "arm64-v8a"},
	{"V2207", "vivo", "13", "V2207_13.0.0", "1080x2400", "409", "arm64-v8a"},
	{"TECNO LE8", "TECNO", "12", "TECNO-LK8_12.0.0", "720x1600", "320", "arm64-v8a"},
}

var (
	locales    = []string{"tr-TR", "en-US", "de-DE", "fr-FR"}
	timezones  = []string{"Europe/Istanbul", "Europe/Berlin", "America/New_York", "Europe/London"}
	operators  = []string{"Turkcell", "Vodafone TR", "Türk Telekom", "AT&T", "T-Mobile", "Vodafone DE"}
	userAgents = []string{
		"Mozilla/5.0 (Linux; Android 13; SM-A525F) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 12; Mi 11 Lite) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 13; CPH2451) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 14; Pixel 6a) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 13; motorola edge 30) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Mobile Safari/537.36",
	}
	appsCommon = []string{
		"com.whatsapp", "com.instagram.android", "com.twitter.android",
		"com.spotify.music", "com.netflix.mediaclient", "com.google.android.youtube",
		"com.zhiliaoapp.musically", // TikTok
		"org.telegram.messenger",
		"com.facebook.katana",
		"com.discord",
	}
)

// Manager, 25 hesap için 25 farklı profil üretir
type Manager struct {
	profiles map[string]*DeviceProfile // tester_id (email) → profile
	rng      *rand.Rand
}

func NewManager() *Manager {
	return &Manager{
		profiles: make(map[string]*DeviceProfile),
		rng:      rand.New(rand.NewSource(timeNow().UnixNano())),
	}
}

// GenerateForTester, belirli bir tester için benzersiz profil üretir
func (m *Manager) GenerateForTester(testerEmail string) *DeviceProfile {
	if existing, ok := m.profiles[testerEmail]; ok {
		return existing
	}
	idx := m.rng.Intn(len(baseModels))
	tmpl := baseModels[idx]

	locale := locales[m.rng.Intn(len(locales))]
	timezone := timezones[m.rng.Intn(len(timezones))]
	operator := operators[m.rng.Intn(len(operators))]
	ua := userAgents[m.rng.Intn(len(userAgents))]

	p := &DeviceProfile{
		ID:             uuid.NewString(),
		AndroidID:      randomHex(16),
		IMEI:           randomIMEI(),
		MacAddress:     randomMAC(),
		Model:          tmpl.Model,
		Manufacturer:   tmpl.Manufacturer,
		AndroidVersion: tmpl.Android,
		BuildNumber:    tmpl.Build,
		ScreenRes:      tmpl.Screen,
		Density:        tmpl.Density,
		UserAgent:      ua,
		Locale:         locale,
		Timezone:       timezone,
		Telephony:      operator,
		GLVendor:       randomGLVendor(),
		GLRenderer:     randomGLRenderer(),
		CPU_ABI:        tmpl.CPU,
		NetworkType:    pickNetwork(),
		SIMOperator:    operator,
		Keyboard:       pickKeyboard(),
		FontsHash:      randomHex(8),
		BatteryLevel:   30 + m.rng.Intn(70),
		SignalStrength: -90 + m.rng.Intn(50),
		AppsInstalled:  append([]string{}, appsCommon...),
	}
	m.profiles[testerEmail] = p
	return p
}

// ApplyToDevice, ADB shell komutları ile profili cihaza uygular
// (Yeni emulator cold-boot'tan sonra çağrılır)
func (p *DeviceProfile) ApplyToDevice(adbExec func(string) (string, error)) error {
	commands := []string{
		// ANDROID_ID
		fmt.Sprintf("settings put secure android_id %s", p.AndroidID),
		// Build props (gerçek cihaz görünümü)
		fmt.Sprintf("setprop ro.product.model %s", p.Model),
		fmt.Sprintf("setprop ro.product.manufacturer %s", p.Manufacturer),
		fmt.Sprintf("setprop ro.build.version.release %s", p.AndroidVersion),
		fmt.Sprintf("setprop ro.build.display.id %s", p.BuildNumber),
		fmt.Sprintf("setprop ro.build.version.sdk %d", sdkForVersion(p.AndroidVersion)),
		fmt.Sprintf("setprop ro.product.cpu.abi %s", p.CPU_ABI),
		// Locale
		fmt.Sprintf("setprop persist.sys.locale %s", strings.ReplaceAll(p.Locale, "-", "_")),
		fmt.Sprintf("setprop persist.sys.timezone %s", p.Timezone),
		// Telephony
		fmt.Sprintf("setprop gsm.operator.alpha %s", p.Telephony),
		fmt.Sprintf("setprop gsm.sim.operator.alpha %s", p.SIMOperator),
		// Screen
		fmt.Sprintf("wm size %s", p.ScreenRes),
		fmt.Sprintf("wm density %s", p.Density),
		// GL
		fmt.Sprintf("setprop ro.hardware.egl %s", p.GLVendor),
		fmt.Sprintf("setprop ro.hardware.gralloc %s", p.GLRenderer),
	}
	for _, cmd := range commands {
		if _, err := adbExec(cmd); err != nil {
			return fmt.Errorf("apply %q: %w", cmd, err)
		}
	}
	return nil
}

func sdkForVersion(v string) int {
	switch v {
	case "14":
		return 34
	case "13":
		return 33
	case "12":
		return 31
	case "11":
		return 30
	}
	return 33
}

func pickNetwork() string {
	opts := []string{"wifi", "4g", "5g", "4g", "wifi"}
	return opts[rand.Intn(len(opts))]
}

func pickKeyboard() string {
	opts := []string{
		"com.google.android.inputmethod.latin/com.android.inputmethod.latin.LatinIME",
		"com.samsung.android.honeyboard/com.samsung.android.honeyboard.service.HoneyBoardService",
		"com.touchtype.swiftkey/com.touchtype.KeyboardService",
	}
	return opts[rand.Intn(len(opts))]
}

func randomHex(n int) string {
	const hexchars = "0123456789abcdef"
	b := make([]byte, n)
	for i := range b {
		b[i] = hexchars[rand.Intn(len(hexchars))]
	}
	return string(b)
}

func randomIMEI() string {
	// 15 haneli Luhn geçerli
	b := make([]byte, 14)
	for i := range b {
		b[i] = byte('0' + rand.Intn(10))
	}
	// Check digit (basitleştirilmiş)
	sum := 0
	for i := 0; i < 14; i++ {
		d := int(b[14-1-i] - '0')
		if i%2 == 0 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	check := (10 - (sum % 10)) % 10
	return string(b) + string(byte('0'+check))
}

func randomMAC() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = byte(rand.Intn(256))
	}
	// Yerel olarak yönetilen, unicast
	b[0] = (b[0] & 0xFE) | 0x02
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", b[0], b[1], b[2], b[3], b[4], b[5])
}

func randomGLVendor() string {
	opts := []string{"Qualcomm", "ARM", "Imagination Technologies", "NVIDIA"}
	return opts[rand.Intn(len(opts))]
}

func randomGLRenderer() string {
	opts := []string{
		"Adreno (TM) 640",
		"Mali-G77 MC9",
		"PowerVR Rogue GX6250",
		"GeForce GTX 1050",
	}
	return opts[rand.Intn(len(opts))]
}
