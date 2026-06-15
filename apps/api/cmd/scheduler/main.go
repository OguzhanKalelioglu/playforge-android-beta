package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/config"
	"github.com/testerscommunity/api/internal/lib"
	"github.com/testerscommunity/api/internal/scheduler"
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

	redisOpt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		logger.Fatal("redis url parse failed", zap.Error(err))
	}

	location, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		location = time.UTC
	}

	sched := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{
		Location: location,
		Logger:   &lib.AsynqLogger{Logger: logger},
	})

	registrar := scheduler.NewRegistrar(sched, logger)

	// Periyodik job örneği: stale assignment cleanup
	// (her gün 04:00 — eski pending assignment'ları temizle)
	if _, err := sched.Register("0 4 * * *",
		asynq.NewTask("system:cleanup_stale", nil),
	); err != nil {
		logger.Warn("daily cleanup registration failed", zap.Error(err))
	}

	// Periyodik job: warming hesapları active'e çevir (3 gün geçmiş)
	if _, err := sched.Register("0 5 * * *",
		asynq.NewTask("system:check_warming", nil),
	); err != nil {
		logger.Warn("warming check registration failed", zap.Error(err))
	}

	_ = registrar // Plan kayıtları scheduler.Register14DayPlan() üzerinden tetiklenir

	go func() {
		logger.Info("scheduler starting", zap.String("timezone", location.String()))
		if err := sched.Run(); err != nil {
			logger.Fatal("scheduler failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("scheduler shutting down...")
	sched.Shutdown()
}
