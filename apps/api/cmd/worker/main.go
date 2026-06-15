package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

	mux := asynq.NewServeMux()
	mux.Use(lib.LoggingMiddleware(logger))

	// Worker handlers buraya eklenecek (Hafta 5-6)
	// mux.HandleFunc(tasks.TypeTestStart, handlers.HandleTestStart)
	// mux.HandleFunc(tasks.TypeDailyEngagement, handlers.HandleDailyEngagement)
	// mux.HandleFunc(tasks.TypeWriteReview, handlers.HandleWriteReview)

	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
		Logger: &lib.AsynqLogger{Logger: logger},
	})

	go func() {
		logger.Info("worker starting", zap.Int("concurrency", 10))
		if err := srv.Run(mux); err != nil {
			logger.Fatal("worker failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("worker shutting down...")
	srv.Shutdown()
	_ = context.Background()
}
