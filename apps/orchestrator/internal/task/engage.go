package task

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"

	"github.com/testerscommunity/orchestrator/internal/appium"
)

// EngageTask, günlük engagement (rastgele swipe/tap/bekle pattern'i)
// 2-5 dakika arası sürer, organik görünüm için AntiDetect kullanır
type EngageTask struct {
	Day int // 1-14 arası, engagement yoğunluğunu belirler
}

func (t *EngageTask) Type() string { return "engage" }

func (t *EngageTask) Run(ctx context.Context, env Env) error {
	// Süre hesaplama: gün 7 ve 10 ağır (5dk), diğerleri 2-3dk
	minSec, maxSec := 120, 180
	if t.Day == 7 || t.Day == 10 {
		minSec, maxSec = 240, 300
	}
	duration := env.Anti.EngagementDuration(minSec, maxSec)
	deadline := time.Now().Add(duration)

	env.Log("engage started", zap.Int("day", t.Day), zap.Duration("duration", duration))

	// 1. Uygulamayı aç
	if err := env.ReportStep("launch_app", func() error {
		return env.Session.AppActivate(ctx, env.PackageName)
	}); err != nil {
		return fmt.Errorf("launch app: %w", err)
	}

	// Anti-detection delay
	env.Anti.Delay(ctx)

	stepCount := 0
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		action := rand.Intn(5) // 0-4
		var actionErr error
		switch action {
		case 0:
			actionErr = env.Session.Scroll(ctx, appium.DirUp)
		case 1:
			actionErr = env.Session.Scroll(ctx, appium.DirDown)
		case 2:
			x, y := env.Anti.JitterXY(540+rand.Intn(100)-50, 1200+rand.Intn(200)-100)
			actionErr = env.Session.Tap(ctx, x, y)
		case 3:
			actionErr = env.Session.Back(ctx)
		case 4:
			time.Sleep(env.Anti.GestureDelay())
		}

		if actionErr != nil {
			env.Log("engage action error (continuing)", zap.Error(actionErr))
		}

		stepCount++
		env.Anti.Delay(ctx) // Aksiyonlar arası bekleme (gaussian)

		// Her 30 saniyede bir screenshot al
		if stepCount%15 == 0 {
			if _, err := env.Session.ScreenshotToFile(ctx, env.ScreenshotDir, fmt.Sprintf("engage_d%d_s%d", t.Day, stepCount)); err != nil {
				env.Log("screenshot failed", zap.Error(err))
			}
		}
	}

	env.Log("engage completed", zap.Int("step_count", stepCount))
	return nil
}
