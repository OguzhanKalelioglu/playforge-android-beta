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
	"github.com/testerscommunity/orchestrator/internal/emulator"
	"github.com/testerscommunity/orchestrator/internal/lib"
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

	adbClient := adb.NewClient(cfg.AdbHost, cfg.AdbPort)
	if err := adbClient.StartADBServer(); err != nil {
		logger.Warn("adb start-server failed (continuing)", zap.Error(err))
	}

	pool := emulator.NewPool(cfg.EmulatorCount)
	logger.Info("emulator pool initialized", zap.Int("count", cfg.EmulatorCount))

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), lib.RequestLogger(logger))

	srv := api.NewServer(pool, logger, cfg.APIToken)
	srv.Register(r)

	httpSrv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		logger.Info("orchestrator starting",
			zap.String("port", cfg.Port),
			zap.Int("emulators", cfg.EmulatorCount),
			zap.String("adb", fmt.Sprintf("%s:%d", cfg.AdbHost, cfg.AdbPort)))
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down orchestrator...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("forced shutdown", zap.Error(err))
	}
	logger.Info("orchestrator exited")
}
