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

	redisOpt, err := asynq.ParseRedisURL(cfg.RedisURL)
	if err != nil {
		logger.Fatal("redis url parse failed", zap.Error(err))
	}

	location, err := time.LoadLocation("Europe/Istanbul")
	if err != nil {
		location = time.UTC
	}

	scheduler := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{
		Location: location,
		Logger:   &lib.AsynqLogger{Logger: logger},
	})

	// Periodic jobs buraya eklenecek (Hafta 5-6)
	// _, _ = scheduler.Register("@every 1h", asynq.NewTask(tasks.TypeCleanupStaleSessions, nil))
	// _, _ = scheduler.Register("@every 24h", asynq.NewTask(tasks.TypeCheckWarmingAccounts, nil))

	go func() {
		logger.Info("scheduler starting", zap.String("timezone", location.String()))
		if err := scheduler.Run(); err != nil {
			logger.Fatal("scheduler failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("scheduler shutting down...")
	scheduler.Shutdown()
}
