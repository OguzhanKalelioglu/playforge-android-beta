package appium

// Capabilities, W3C WebDriver capabilities
// Appium 2.x + UiAutomator2 driver için
type Capabilities struct {
	// W3C standard
	PlatformName      string // "Android"
	BrowserName       string // usually empty for native
	AcceptInsecureCerts bool

	// Appium
	DeviceName         string // "emulator-5554" (pool serial)
	AutomationName     string // "UiAutomator2"
	App                string // "com.example.app" veya path
	AppPackage         string
	AppActivity        string
	AppWaitActivity    string
	NoReset            bool
	FullReset          bool
	AutoGrantPermissions bool
	NewCommandTimeout  int    // seconds
	Uiautomator2ServerLaunchTimeout int
	Uiautomator2ServerInstallTimeout int
	SkipServerInstallation bool

	// Vendor-specific (appium prefix)
	UDID         string
	SystemPort   int
	ChromeDriverPort int
	Orientation  string
	Locale       string
	TimeZone     string
}

// AndroidEmulatorCaps, varsayılan Android emulator ayarları
func AndroidEmulatorCaps(deviceName, appPackage, appActivity string) Capabilities {
	return Capabilities{
		PlatformName:          "Android",
		AutomationName:        "UiAutomator2",
		DeviceName:            deviceName,
		AppPackage:            appPackage,
		AppActivity:           appActivity,
		NoReset:               true,
		AutoGrantPermissions:  true,
		NewCommandTimeout:     600, // 10 dk, watchdog'dan büyük
		Uiautomator2ServerLaunchTimeout: 60000,
		Uiautomator2ServerInstallTimeout: 60000,
		SkipServerInstallation: true,
	}
}

// FreshCaps, opt-in/download için (FullReset)
func FreshCaps(deviceName, appPackage string) Capabilities {
	c := AndroidEmulatorCaps(deviceName, appPackage, "")
	c.NoReset = false
	c.FullReset = true
	return c
}
