package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/testerscommunity/orchestrator/internal/adb"
	"github.com/testerscommunity/orchestrator/internal/container"
	"github.com/testerscommunity/orchestrator/internal/emulator"
	"github.com/testerscommunity/orchestrator/internal/health"
)

type Manager struct {
	pool         *emulator.Pool
	container    *container.Manager
	healthMon    *health.Monitor
	adb          *adb.Client
	logger       *zap.Logger
	mu           sync.Mutex
	running      bool
	stopCh       chan struct{}
	autoStart    bool
	checkInterval time.Duration
}

type Config struct {
	Pool            *emulator.Pool
	Container       *container.Manager
	HealthMonitor   *health.Monitor
	ADBClient       *adb.Client
	Logger          *zap.Logger
	AutoStart       bool
	CheckInterval   time.Duration
}

func NewManager(cfg Config) *Manager {
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 60 * time.Second
	}
	return &Manager{
		pool:          cfg.Pool,
		container:     cfg.Container,
		healthMon:     cfg.HealthMonitor,
		adb:           cfg.ADBClient,
		logger:        cfg.Logger,
		stopCh:        make(chan struct{}),
		autoStart:     cfg.AutoStart,
		checkInterval: cfg.CheckInterval,
	}
}

func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return fmt.Errorf("manager already running")
	}
	m.running = true
	m.mu.Unlock()

	m.logger.Info("lifecycle manager started",
		zap.Int("emulators", m.pool.Count()),
		zap.Bool("auto_start", m.autoStart),
		zap.Duration("check_interval", m.checkInterval))

	if m.autoStart {
		if err := m.StartAllEmulators(ctx); err != nil {
			m.logger.Error("auto-start failed", zap.Error(err))
		}
	}

	go m.runHealthLoop(ctx)

	<-ctx.Done()
	m.logger.Info("lifecycle manager stopping...")
	close(m.stopCh)
	return nil
}

func (m *Manager) StartAllEmulators(ctx context.Context) error {
	m.logger.Info("starting all emulators", zap.Int("count", m.pool.Count()))

	if err := m.container.UpAll(ctx); err != nil {
		return fmt.Errorf("docker compose up: %w", err)
	}

	m.refreshAllStatuses(ctx)
	m.waitForAllReady(ctx, 8*time.Minute)
	return nil
}

func (m *Manager) StartEmulator(ctx context.Context, serial string) error {
	e, err := m.pool.Get(serial)
	if err != nil {
		return err
	}

	if e.Status == emulator.StatusReady || e.Status == emulator.StatusBooting {
		return fmt.Errorf("emulator %s already %s", serial, e.Status)
	}

	service := container.ServiceFor(e.Index, m.container.ServicePrefix)
	m.pool.SetStatus(serial, emulator.StatusBooting, "")

	if err := m.container.Up(ctx, service); err != nil {
		m.pool.SetStatus(serial, emulator.StatusError, err.Error())
		return fmt.Errorf("docker up: %w", err)
	}

	m.refreshEmulatorStatus(ctx, serial)

	go func() {
		bootCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if _, err := m.healthMon.WaitForBoot(bootCtx, serial); err != nil {
			m.pool.SetStatus(serial, emulator.StatusError, err.Error())
			m.logger.Error("boot timeout", zap.String("serial", serial), zap.Error(err))
			return
		}
		m.pool.SetStatus(serial, emulator.StatusReady, "")
		m.logger.Info("emulator ready", zap.String("serial", serial))
	}()

	return nil
}

func (m *Manager) StopEmulator(ctx context.Context, serial string) error {
	e, err := m.pool.Get(serial)
	if err != nil {
		return err
	}

	service := container.ServiceFor(e.Index, m.container.ServicePrefix)
	if err := m.container.Stop(ctx, service); err != nil {
		return fmt.Errorf("docker stop: %w", err)
	}
	return m.pool.SetStatus(serial, emulator.StatusStopped, "")
}

func (m *Manager) RestartEmulator(ctx context.Context, serial string) error {
	e, err := m.pool.Get(serial)
	if err != nil {
		return err
	}

	service := container.ServiceFor(e.Index, m.container.ServicePrefix)
	m.pool.SetStatus(serial, emulator.StatusBooting, "")

	if err := m.container.Restart(ctx, service); err != nil {
		m.pool.SetStatus(serial, emulator.StatusError, err.Error())
		return fmt.Errorf("docker restart: %w", err)
	}

	go func() {
		bootCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if _, err := m.healthMon.WaitForBoot(bootCtx, serial); err != nil {
			m.pool.SetStatus(serial, emulator.StatusError, err.Error())
			m.logger.Error("boot timeout after restart", zap.String("serial", serial), zap.Error(err))
			return
		}
		m.pool.SetStatus(serial, emulator.StatusReady, "")
		m.logger.Info("emulator ready after restart", zap.String("serial", serial))
	}()

	return nil
}

func (m *Manager) WipeEmulator(ctx context.Context, serial string) error {
	if _, err := m.pool.Get(serial); err != nil {
		return err
	}

	m.pool.SetStatus(serial, emulator.StatusWiping, "")

	if err := m.adb.WipeData(ctx, serial); err != nil {
		m.pool.SetStatus(serial, emulator.StatusError, err.Error())
		return fmt.Errorf("wipe data: %w", err)
	}

	time.Sleep(2 * time.Second)
	booted, _ := m.adb.IsBooted(ctx, serial)
	if booted {
		m.pool.SetStatus(serial, emulator.StatusReady, "")
	} else {
		m.pool.SetStatus(serial, emulator.StatusBooting, "")
		go func() {
			bootCtx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			if _, err := m.healthMon.WaitForBoot(bootCtx, serial); err != nil {
				m.pool.SetStatus(serial, emulator.StatusError, err.Error())
				return
			}
			m.pool.SetStatus(serial, emulator.StatusReady, "")
		}()
	}

	return nil
}

func (m *Manager) ResetForTest(ctx context.Context, serial string) error {
	m.logger.Info("resetting emulator for test", zap.String("serial", serial))

	if err := m.WipeEmulator(ctx, serial); err != nil {
		return fmt.Errorf("wipe: %w", err)
	}

	return nil
}

func (m *Manager) StopAllEmulators(ctx context.Context) error {
	m.logger.Warn("stopping ALL emulators (orchestrator down)")
	return m.container.Down(ctx)
}

func (m *Manager) runHealthLoop(ctx context.Context) {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.checkAllHealth(ctx)
		}
	}
}

func (m *Manager) checkAllHealth(ctx context.Context) {
	for _, e := range m.pool.List() {
		if e.Status == emulator.StatusStopped {
			continue
		}

		healthCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		res, err := m.healthMon.Check(healthCtx, e.Serial)
		cancel()

		if err != nil {
			if e.Status != emulator.StatusError {
				m.logger.Warn("emulator health check failed",
					zap.String("serial", e.Serial),
					zap.Error(err))
				m.pool.SetStatus(e.Serial, emulator.StatusError, err.Error())
			}
			continue
		}

		if res.BootCompleted && e.Status == emulator.StatusBooting {
			m.pool.SetStatus(e.Serial, emulator.StatusReady, "")
			m.logger.Info("emulator transitioned to ready",
				zap.String("serial", e.Serial))
		} else if !res.BootCompleted && e.Status == emulator.StatusReady {
			m.pool.SetStatus(e.Serial, emulator.StatusBooting, "")
			m.logger.Warn("emulator went back to booting (reboot detected)",
				zap.String("serial", e.Serial))
		}
	}
}

func (m *Manager) refreshAllStatuses(ctx context.Context) {
	containers, err := m.container.PS(ctx)
	if err != nil {
		m.logger.Warn("failed to list containers", zap.Error(err))
		return
	}

	containerByService := map[string]string{}
	for _, c := range containers {
		if c.Service != "" {
			containerByService[c.Service] = c.ID
		}
	}

	for _, e := range m.pool.List() {
		service := container.ServiceFor(e.Index, m.container.ServicePrefix)
		containerID, exists := containerByService[service]
		if exists && containerID != "" {
			if err := m.pool.SetContainerID(e.Serial, containerID); err == nil {
				_ = e
			}
		}
	}
}

func (m *Manager) refreshEmulatorStatus(ctx context.Context, serial string) {
	e, err := m.pool.Get(serial)
	if err != nil {
		return
	}
	containerID, err := m.container.ContainerID(ctx, container.ServiceFor(e.Index, m.container.ServicePrefix))
	if err == nil && containerID != "" {
		_ = m.pool.SetContainerID(serial, containerID)
	}
}

func (m *Manager) waitForAllReady(ctx context.Context, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		counts := m.pool.Counts()
		ready := counts[emulator.StatusReady]
		booting := counts[emulator.StatusBooting]
		errored := counts[emulator.StatusError]

		if booting == 0 && errored == 0 {
			m.logger.Info("all emulators ready", zap.Int("ready", ready))
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}

	counts := m.pool.Counts()
	m.logger.Warn("wait for all ready timed out",
		zap.Any("counts", counts))
}
