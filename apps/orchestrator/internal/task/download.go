package task

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// DownloadTask, Play Store'dan uygulamayı indirir + kurar
type DownloadTask struct{}

func (t *DownloadTask) Type() string { return "download" }

func (t *DownloadTask) Run(ctx context.Context, env Env) error {
	env.Log("download started", zap.String("package", env.PackageName))

	// 1. Play Store'u aç
	if err := env.ReportStep("open_play_store", func() error {
		return env.Session.AppActivate(ctx, "com.android.vending")
	}); err != nil {
		return fmt.Errorf("open play store: %w", err)
	}

	// 2. Search'e package adını yaz
	if err := env.ReportStep("search_package", func() error {
		return searchPackage(ctx, env, env.PackageName)
	}); err != nil {
		return fmt.Errorf("search package: %w", err)
	}

	// 3. Install butonuna tıkla
	if err := env.ReportStep("click_install", func() error {
		elem, err := env.Session.WaitForElement(ctx,
			"appium:-android uiautomator",
			`new UiSelector().textMatches("(?i)(install|get|update)")`,
			15*time.Second)
		if err != nil {
			return err
		}
		return elem.Tap(ctx)
	}); err != nil {
		return fmt.Errorf("click install: %w", err)
	}

	// 4. İndirme + kurulum tamamlanmasını bekle
	if err := env.ReportStep("wait_install_complete", func() error {
		// Open butonu görünene kadar bekle
		_, err := env.Session.WaitForElement(ctx,
			"appium:-android uiautomator",
			`new UiSelector().textMatches("(?i)open")`,
			5*time.Minute)
		return err
	}); err != nil {
		return fmt.Errorf("wait install: %w", err)
	}

	env.Log("download completed")
	return nil
}

func searchPackage(ctx context.Context, env Env, pkg string) error {
	// Search ikonuna tıkla (üst toolbar)
	// Real implementasyon: search box locator
	// Şimdilik placeholder: package adını intent ile gönder
	return env.Session.ActivityStart(ctx, "com.android.vending", "com.google.android.finsky.activities.MainActivity")
}
