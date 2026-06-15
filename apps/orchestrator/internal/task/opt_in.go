package task

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// OptInTask, Play Store üzerinden kapalı beta test'e katılım
// "Become a tester" butonuna tıklar
type OptInTask struct{}

func (t *OptInTask) Type() string { return "opt_in" }

func (t *OptInTask) Run(ctx context.Context, env Env) error {
	env.Log("opt_in started")

	// 1. Test link'i Chrome'da aç
	if err := env.ReportStep("open_test_link", func() error {
		return openTestLink(ctx, env)
	}); err != nil {
		return fmt.Errorf("open test link: %w", err)
	}

	// 2. "Become a tester" butonunu bul ve tıkla
	if err := env.ReportStep("click_become_tester", func() error {
		elem, err := env.Session.WaitForElement(ctx,
			"appium:-android uiautomator",
			`new UiSelector().textContains("Become a tester")`,
			30*time.Second)
		if err != nil {
			return fmt.Errorf("find become a tester button: %w", err)
		}
		return elem.Tap(ctx)
	}); err != nil {
		return fmt.Errorf("click become a tester: %w", err)
	}

	// 3. Onay sayfası "Got it" / "Confirm" tıkla (varsa)
	if err := env.ReportStep("confirm_opt_in", func() error {
		elem, err := env.Session.WaitForElement(ctx,
			"appium:-android uiautomator",
			`new UiSelector().textMatches("(?i)(got it|confirm|join)")`,
			10*time.Second)
		if err != nil {
			// Opsiyonel adım, hata durumunda skip
			env.Log("confirm button not found, skipping", zap.Error(err))
			return nil
		}
		return elem.Tap(ctx)
	}); err != nil {
		return fmt.Errorf("confirm opt in: %w", err)
	}

	env.Log("opt_in completed")
	return nil
}

func openTestLink(ctx context.Context, env Env) error {
	// Test linki Env'de yoksa helper'a dönüşmez
	// (test_start task'ında payload üzerinden gelir)
	// Burada basit: Chrome'da URL aç
	// Real implementasyon: env.Metadata["test_link"] kullanır
	return env.Session.ActivityStart(ctx, "com.android.chrome", "com.google.android.apps.chrome.Main")
}
