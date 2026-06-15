package task

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// ReviewTask, Play Store'da yorum yazar
// 14. gün, 5 yıldız, karışık puan dağılımı
type ReviewTask struct {
	Stars   int
	Comment string
}

func (t *ReviewTask) Type() string { return "review" }

func (t *ReviewTask) Run(ctx context.Context, env Env) error {
	if t.Stars < 1 || t.Stars > 5 {
		t.Stars = 5 // default
	}
	if t.Comment == "" {
		t.Comment = "Güzel uygulama, tavsiye ederim." // fallback
	}

	env.Log("review started", zap.Int("stars", t.Stars))

	// 1. Play Store'da uygulama sayfasını aç
	if err := env.ReportStep("open_app_page", func() error {
		return env.Session.AppActivate(ctx, "com.android.vending")
	}); err != nil {
		return fmt.Errorf("open play store: %w", err)
	}

	// 2. Yıldız seç
	if err := env.ReportStep("select_stars", func() error {
		return tapStars(ctx, env, t.Stars)
	}); err != nil {
		return fmt.Errorf("select stars: %w", err)
	}

	env.Anti.Delay(ctx)

	// 3. Yorum yaz
	if err := env.ReportStep("write_comment", func() error {
		elem, err := env.Session.WaitForElement(ctx,
			"appium:-android uiautomator",
			`new UiSelector().resourceId("com.android.vending:id/review_text")`,
			10*time.Second)
		if err != nil {
			// Alternatif selector
			elem, err = env.Session.WaitForElement(ctx,
				"appium:-android uiautomator",
				`new UiSelector().className("android.widget.EditText")`,
				10*time.Second)
			if err != nil {
				return err
			}
		}
		if err := elem.Clear(ctx); err != nil {
			return err
		}
		return elem.SendKeys(ctx, t.Comment)
	}); err != nil {
		return fmt.Errorf("write comment: %w", err)
	}

	env.Anti.Delay(ctx)

	// 4. Post butonuna tıkla
	if err := env.ReportStep("post_review", func() error {
		elem, err := env.Session.WaitForElement(ctx,
			"appium:-android uiautomator",
			`new UiSelector().textMatches("(?i)(post|submit|publish|gönder|yayınla)")`,
			10*time.Second)
		if err != nil {
			return err
		}
		return elem.Tap(ctx)
	}); err != nil {
		return fmt.Errorf("post review: %w", err)
	}

	env.Log("review completed", zap.Int("stars", t.Stars))
	return nil
}

func tapStars(ctx context.Context, env Env, stars int) error {
	// 5 yıldızın x koordinatları (Pixel 5, ~1080px width)
	// Yıldızlar genelde review dialog'unda ortalı
	// Her yıldız ~150px aralıklı
	// Real implementasyon: ekran çözünürlüğüne göre dinamik hesapla
	starWidth := 100
	startX := 200 + (1080-200-5*starWidth)/2
	y := 600 // placeholder

	x := startX + (stars-1)*starWidth + rand.Intn(20) - 10
	y = y + rand.Intn(20) - 10

	x, y = env.Anti.JitterXY(x, y)
	return env.Session.Tap(ctx, x, y)
}
