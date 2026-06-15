package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/config"
	"github.com/testerscommunity/api/internal/lib"
	"github.com/testerscommunity/api/internal/worker"
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

	// Runner client (orchestrator'a HTTP çağrı)
	runnerURL := os.Getenv("RUNNER_URL")
	if runnerURL == "" {
		runnerURL = "http://127.0.0.1:9000"
	}
	runnerToken := os.Getenv("RUNNER_API_TOKEN")
	if runnerToken == "" {
		runnerToken = cfg.OrchestratorAPIToken
	}
	runner := worker.NewRunnerClient(runnerURL, runnerToken, logger)

	mux := asynq.NewServeMux()
	mux.Use(worker.RecoveryMiddleware(logger))
	mux.Use(worker.LoggingMiddleware(logger))
	mux.Use(worker.TimeoutMiddleware(15 * time.Minute))

	// Task handler'ları register
	mux.HandleFunc(string(cfg.TaskTypeTestStart), worker.NewTestStartHandler(runner, logger).ProcessTask)
	mux.HandleFunc(string(cfg.TaskTypeDailyEngagement), worker.NewDailyEngagementHandler(runner, logger).ProcessTask)
	mux.HandleFunc(string(cfg.TaskTypeWriteReview), worker.NewWriteReviewHandler(runner, logger).ProcessTask)
	mux.HandleFunc(string(cfg.TaskTypeLoginGoogle), worker.NewLoginGoogleHandler(runner, logger).ProcessTask)

	srv := asynq.NewServer(redisOpt, asynq.Config{
		Concurrency: 10,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
		Logger:       &lib.AsynqLogger{Logger: logger},
		BaseContext: func() context.Context { return context.Background() },
	})

	go func() {
		logger.Info("worker starting",
			zap.Int("concurrency", 10),
			zap.String("runner_url", runnerURL))
		if err := srv.Run(mux); err != nil {
			logger.Fatal("worker failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("worker shutting down...")
	srv.Shutdown()
}
