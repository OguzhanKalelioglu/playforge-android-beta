package taskrunner

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"

	"github.com/testerscommunity/orchestrator/internal/appium"
	"github.com/testerscommunity/orchestrator/internal/emulator"
	"github.com/testerscommunity/orchestrator/internal/task"
)

// Runner, task'ları çalıştıran ana orkestrasyon katmanı
// - Emulator pool'dan hazır alır (CAS: ready → busy)
// - Appium session açar
// - Watchdog ile süre sınırı koyar (default 10dk)
// - Step'leri ActivitySink'e raporlar
// - Hata durumunda retry (exp backoff)
// - Single-flight per testID (aynı anda 1 task)
type Runner struct {
	pool      *emulator.Pool
	appiumURL string
	anti      *AntiDetect
	sink      *ActivitySink
	logger    *zap.Logger

	// Single-flight: aynı testID için paralel Run engellenir
	sf singleflight.Group

	// Watchdog timeout (per task)
	watchdog time.Duration

	// Screenshot dizini (relative, Mini PC'de)
	screenshotDir string

	// Retry policy
	maxRetries int
}

type Config struct {
	Pool         *emulator.Pool
	AppiumURL    string
	ActivitySink *ActivitySink
	Logger       *zap.Logger
	Watchdog     time.Duration
	ScreenshotDir string
}

func NewRunner(cfg Config) *Runner {
	if cfg.Watchdog == 0 {
		cfg.Watchdog = 10 * time.Minute
	}
	if cfg.ScreenshotDir == "" {
		cfg.ScreenshotDir = "./screenshots"
	}
	return &Runner{
		pool:         cfg.Pool,
		appiumURL:    cfg.AppiumURL,
		anti:         NewAntiDetect(),
		sink:         cfg.ActivitySink,
		logger:       cfg.Logger,
		watchdog:     cfg.Watchdog,
		screenshotDir: cfg.ScreenshotDir,
		maxRetries:   3,
	}
}

// Submit, bir task'ı çalıştırma kuyruğuna ekler
// Aynı testID için zaten çalışan task varsa, mevcut sonucu bekler (single-flight)
func (r *Runner) Submit(ctx context.Context, t task.Task, params SubmitParams) (<-chan RunResult, error) {
	key := fmt.Sprintf("%s:%s", params.TestID, t.Type())
	ch := r.sf.DoChan(key, func() (interface{}, error) {
		return r.runWithRetry(ctx, t, params)
	})

	resultCh := make(chan RunResult, 1)
	go func() {
		res := <-ch
		if res.Shared {
			// Single-flight suppression: başka bir Submit zaten çalıştırıyor
			// Caller'ın bunu handle etmesi için özel error
			resultCh <- RunResult{Err: ErrAlreadyRunning}
		} else {
			if errVal, ok := res.Val.(error); ok {
				resultCh <- RunResult{Err: errVal}
			} else if rr, ok := res.Val.(RunResult); ok {
				resultCh <- rr
			} else {
				resultCh <- RunResult{Err: fmt.Errorf("unexpected result type: %T", res.Val)}
			}
		}
		close(resultCh)
	}()
	return resultCh, nil
}

var ErrAlreadyRunning = errors.New("task already running for this testID")

// SubmitParams, task çalıştırma parametreleri
type SubmitParams struct {
	TestID       string
	AssignmentID string
	PackageName  string
}

// RunResult, task sonucu
type RunResult struct {
	Err        error
	Duration   time.Duration
	StepCount  int
	EmulatorSerial string
}

// runWithRetry, exp backoff ile retry yapar
func (r *Runner) runWithRetry(ctx context.Context, t task.Task, params SubmitParams) (interface{}, error) {
	var lastErr error
	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			// Exp backoff: 30s, 60s, 120s
			backoff := time.Duration(30<<attempt) * time.Second
			r.logger.Warn("retrying task",
				zap.String("task_type", t.Type()),
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff),
				zap.Error(lastErr))
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		result, err := r.runOnce(ctx, t, params)
		if err == nil {
			return result, nil
		}
		lastErr = err

		// Watchdog timeout → retry
		// Appium session lost → retry
		// Other errors → retry (ama bazıları için break)
		if !isRetryable(err) {
			r.logger.Error("task failed (non-retryable)",
				zap.String("task_type", t.Type()),
				zap.Error(err))
			return RunResult{Err: err}, err
		}
	}
	return RunResult{Err: lastErr}, lastErr
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if errors.Is(err, appium.ErrSessionLost) {
		return true
	}
	if errors.Is(err, appium.ErrServerUnreachable) {
		return true
	}
	// Default: retry
	return true
}

// runOnce, tek bir task çalıştırma denemesi
func (r *Runner) runOnce(ctx context.Context, t task.Task, params SubmitParams) (RunResult, error) {
	// 1. Emulator al (blocking, ready olana kadar bekler)
	handle, err := r.pool.AcquireForTestBlocking(ctx, params.TestID, params.AssignmentID, 5*time.Second)
	if err != nil {
		return RunResult{}, fmt.Errorf("acquire emulator: %w", err)
	}
	defer func() {
		if rErr := r.pool.Release(handle.Serial); rErr != nil {
			r.logger.Warn("emulator release failed", zap.Error(rErr))
		}
	}()

	r.logger.Info("task started",
		zap.String("task_type", t.Type()),
		zap.String("serial", handle.Serial),
		zap.String("test_id", params.TestID),
		zap.String("assignment_id", params.AssignmentID))

	// 2. Watchdog context
	watchCtx, cancel := context.WithTimeout(ctx, r.watchdog)
	defer cancel()

	// 3. Appium session aç
	caps := appium.AndroidEmulatorCaps(handle.Serial, params.PackageName, "")
	session, err := appium.New(r.appiumURL).CreateSession(watchCtx, caps)
	if err != nil {
		return RunResult{Err: err, EmulatorSerial: handle.Serial}, fmt.Errorf("create appium session: %w", err)
	}
	defer func() {
		// Idempotent quit
		if qErr := session.Quit(watchCtx); qErr != nil {
			r.logger.Warn("session quit failed", zap.Error(qErr))
		}
	}()

	// 4. Env hazırla
	screenshotDir := filepath.Join(r.screenshotDir, params.TestID)
	env := task.Env{
		Session:       session,
		AssignmentID:  params.AssignmentID,
		TestID:        params.TestID,
		PackageName:   params.PackageName,
		Logger:        r.logger.With(zap.String("task", t.Type())),
		ScreenshotDir: screenshotDir,
		Anti:          r.anti,
		Report: func(step task.Step) {
			r.sink.Emit(activityEvent{
				EventType:      mapStepToEventType(step.Status),
				TestID:         params.TestID,
				AssignmentID:   params.AssignmentID,
				StepName:       step.Name,
				Status:         string(step.Status),
				ErrorMessage:   errString(step.Error),
				ScreenshotPath: step.Screenshot,
				Metadata:       step.Metadata,
				OccurredAt:     time.Now().UTC(),
			})
		},
	}

	// 5. Task çalıştır
	start := time.Now()
	runErr := t.Run(watchCtx, env)
	duration := time.Since(start)

	if runErr != nil {
		return RunResult{
			Err:            runErr,
			Duration:       duration,
			EmulatorSerial: handle.Serial,
		}, runErr
	}

	r.logger.Info("task completed",
		zap.String("task_type", t.Type()),
		zap.String("serial", handle.Serial),
		zap.Duration("duration", duration))

	return RunResult{
		Duration:       duration,
		EmulatorSerial: handle.Serial,
	}, nil
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func mapStepToEventType(status task.StepStatus) string {
	switch status {
	case task.StepInProgress:
		return "step_start"
	case task.StepOK:
		return "step_complete"
	case task.StepError:
		return "step_error"
	case task.StepSkipped:
		return "step_complete"
	default:
		return "log"
	}
}

// AcquireForTestBlocking, pool'dan blocking acquire yapar
// ready state'te emulator yoksa timeout'a kadar bekler
func (r *Runner) AcquireForTestBlocking(ctx context.Context, testID, assignmentID string) (*emulator.Emulator, error) {
	return r.pool.AcquireForTestBlocking(ctx, testID, assignmentID, 5*time.Second)
}

// SetWatchdog, watchdog timeout'unu değiştirir (test amaçlı)
func (r *Runner) SetWatchdog(d time.Duration) {
	r.watchdog = d
}

// WaitForAllIdle, tüm task'lar bitene kadar bekler (test amaçlı)
func (r *Runner) WaitForAllIdle(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		busy := r.pool.Counts()[emulator.StatusBusy]
		if busy == 0 {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for all idle")
}

var _ = sync.Mutex{} // keep sync import for future
