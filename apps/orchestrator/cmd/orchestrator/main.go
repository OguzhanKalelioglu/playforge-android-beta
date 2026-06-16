package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/testerscommunity/orchestrator/internal/adb"
	"github.com/testerscommunity/orchestrator/internal/api"
	"github.com/testerscommunity/orchestrator/internal/config"
	"github.com/testerscommunity/orchestrator/internal/container"
	"github.com/testerscommunity/orchestrator/internal/emulator"
	"github.com/testerscommunity/orchestrator/internal/health"
	"github.com/testerscommunity/orchestrator/internal/lib"
	"github.com/testerscommunity/orchestrator/internal/lifecycle"
	"github.com/testerscommunity/orchestrator/internal/taskrunner"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config load error: %v\n", err)
		os.Exit(1)
	}

	logger, err := lib.NewLogger(cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "logger error: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("orchestrator booting",
		zap.String("env", cfg.AppEnv),
		zap.Int("emulators", cfg.EmulatorCount),
		zap.Bool("auto_start", cfg.AutoStartEmul),
		zap.String("appium_url", cfg.AppiumURL),
		zap.String("activity_api_url", cfg.ActivityAPIURL))

	adbClient := adb.NewClient(cfg.AdbHost, cfg.AdbPort)
	if err := adbClient.StartADBServer(); err != nil {
		logger.Warn("adb start-server failed (continuing)", zap.Error(err))
	}

	pool := emulator.NewPool(cfg.EmulatorCount)
	containerMgr := container.NewManager(cfg.ComposePath, cfg.ComposeProject, cfg.ServicePrefix)
	healthMon := health.NewMonitor(cfg.AdbHost, cfg.AdbPort)

	manager := lifecycle.NewManager(lifecycle.Config{
		Pool:          pool,
		Container:     containerMgr,
		HealthMonitor: healthMon,
		ADBClient:     adbClient,
		Logger:        logger,
		AutoStart:     cfg.AutoStartEmul,
		CheckInterval: time.Duration(cfg.CheckInterval) * time.Second,
	})

	// Activity sink (orchestrator → API)
	activitySink := taskrunner.NewActivitySink(
		cfg.ActivityAPIURL,
		cfg.ActivityAPIToken,
		cfg.ActivityFallbackDir,
		logger,
	)

	// Task runner
	runner := taskrunner.NewRunner(taskrunner.Config{
		Pool:          pool,
		AppiumURL:     cfg.AppiumURL,
		ActivitySink:  activitySink,
		Logger:        logger,
		Watchdog:      time.Duration(cfg.TaskWatchdogMin) * time.Minute,
		ScreenshotDir: cfg.ScreenshotDir,
	})

	managerCtx, managerCancel := context.WithCancel(context.Background())
	defer managerCancel()

	go func() {
		if err := manager.Start(managerCtx); err != nil {
			logger.Error("lifecycle manager error", zap.Error(err))
		}
	}()

	// Activity sink start
	sinkCtx, sinkCancel := context.WithCancel(context.Background())
	defer sinkCancel()
	activitySink.Start(sinkCtx)

	// HTTP server
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), lib.RequestLogger(logger))

	taskHandler := api.NewTaskHandler(runner, logger)
	srv := api.NewServer(pool, manager, taskHandler, logger, cfg.APIToken)
	srv.Register(r)
	taskHandler.Register(r, srv.AuthMiddleware())

	httpSrv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		logger.Info("orchestrator API starting", zap.String("port", cfg.Port))
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutdown signal received, draining...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown error", zap.Error(err))
	}

	sinkCancel()
	activitySink.Close()
	managerCancel()
	time.Sleep(2 * time.Second)

	logger.Info("orchestrator exited")
}
