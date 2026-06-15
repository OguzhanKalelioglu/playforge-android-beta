package task

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/testerscommunity/orchestrator/internal/appium"
)

// Task, orchestrator içinde çalıştırılabilen bir UI automation görevi
// Tüm görevler (opt_in, download, engage, review, login) bunu implement eder
type Task interface {
	// Type, görev tipini döndürür (logging/metrics için)
	Type() string

	// Run, görevi çalıştırır. Env içindeki session, emulator, report callback'leri kullanır.
	// Watchdog ctx zaten uygulanmış olarak gelir (10dk default).
	Run(ctx context.Context, env Env) error
}

// Env, task'a sağlanan çalışma ortamı
// Her Run() çağrısında yeni bir Env oluşturulur
type Env struct {
	// Session, Appium WebDriver session
	// Quit() defer ile çağrılmalı (runner otomatik yapar)
	Session *appium.Session

	// AssignmentID, hangi tester için çalışıldığı (DB update için)
	AssignmentID string

	// TestID, hangi test için çalışıldığı
	TestID string

	// PackageName, hedef uygulama (com.example.app)
	PackageName string

	// Logger, structured log
	Logger *zap.Logger

	// Report, step event callback'i (activity_sink'e iletir)
	Report func(step Step)

	// ScreenshotDir, screenshot dosyalarının kaydedileceği dizin
	ScreenshotDir string

	// Anti, anti-detection helper (gaussian delays, jitter, vs.)
	Anti AntiDetect
}

// AntiDetect, gesture humanization için helper interface
// İmplementasyon: internal/taskrunner.AntiDetect
type AntiDetect interface {
	// Delay, aksiyonlar arası bekleme (gaussian, ms)
	Delay(ctx context.Context) time.Duration

	// GestureDelay, gesture'lar arası mikro bekleme (ms)
	GestureDelay() time.Duration

	// JitterXY, koordinatı ±N piksel oynatır (anti-pattern)
	JitterXY(x, y int) (int, int)

	// SwipePath, x1,y1 → x2,y2 arasında Bezier eğrisi + jitter
	// n: ara nokta sayısı
	SwipePath(x1, y1, x2, y2, n int) []Point

	// AppLaunchPause, app launch sonrası rastgele bekleme
	AppLaunchPause() time.Duration

	// EngagementDuration, 2-5dk arası rastgele süre
	EngagementDuration(min, maxSec int) time.Duration
}

// Point, (x,y) koordinat
type Point struct {
	X int
	Y int
}

// Step, task içindeki bir adımın sonucu
// Report callback'i ile activity_sink'e gönderilir
type Step struct {
	Name        string        // "opt_in.click_become_tester"
	Status      StepStatus    // in_progress, ok, error
	Error       error         // hata varsa
	Screenshot  string        // dosya path'i (relative)
	DurationMs  int64         // adım süresi
	Metadata    map[string]interface{} // ekstra bilgi
	StartedAt   time.Time
	CompletedAt time.Time
}

type StepStatus string

const (
	StepInProgress StepStatus = "in_progress"
	StepOK         StepStatus = "ok"
	StepError      StepStatus = "error"
	StepSkipped    StepStatus = "skipped"
)

// ReportStep, Env üzerinden step raporlar
// Task içinden: env.Report(Step{Name: "opt_in.click", Status: StepOK, ...})
func (e *Env) ReportStep(name string, fn func() error) error {
	startedAt := time.Now()
	step := Step{
		Name:      name,
		Status:    StepInProgress,
		StartedAt: startedAt,
		Metadata:  map[string]interface{}{},
	}
	e.Report(step)

	err := fn()
	step.CompletedAt = time.Now()
	step.DurationMs = step.CompletedAt.UnixMilli() - startedAt.UnixMilli()

	if err != nil {
		step.Status = StepError
		step.Error = err
	} else {
		step.Status = StepOK
	}
	e.Report(step)
	return err
}

// ReportStepWithScreenshot, screenshot alıp step raporlar
func (e *Env) ReportStepWithScreenshot(name string, fn func() (screenshotPath string, err error)) error {
	startedAt := time.Now()
	step := Step{
		Name:      name,
		Status:    StepInProgress,
		StartedAt: startedAt,
		Metadata:  map[string]interface{}{},
	}
	e.Report(step)

	screenshotPath, err := fn()
	step.CompletedAt = time.Now()
	step.DurationMs = step.CompletedAt.UnixMilli() - startedAt.UnixMilli()
	step.Screenshot = screenshotPath

	if err != nil {
		step.Status = StepError
		step.Error = err
	} else {
		step.Status = StepOK
	}
	e.Report(step)
	return err
}

// Log, structured log atar
func (e *Env) Log(msg string, fields ...zap.Field) {
	e.Logger.Info(msg, append([]zap.Field{
		zap.String("task_test_id", e.TestID),
		zap.String("task_assignment_id", e.AssignmentID),
	}, fields...)...)
}
