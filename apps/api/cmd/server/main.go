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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/config"
	"github.com/testerscommunity/api/internal/db"
	"github.com/testerscommunity/api/internal/handler"
	"github.com/testerscommunity/api/internal/lib"
	"github.com/testerscommunity/api/internal/middleware"
	"github.com/testerscommunity/api/internal/repository"
	"github.com/testerscommunity/api/internal/service"
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

	ctx := context.Background()

	pg, err := db.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("postgres init failed", zap.Error(err))
	}
	defer pg.Close()

	rdb, err := lib.NewRedis(ctx, cfg.RedisURL)
	if err != nil {
		logger.Fatal("redis init failed", zap.Error(err))
	}
	defer rdb.Close()

	// Asynq client
	asynqRedis, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		logger.Fatal("asynq redis parse failed", zap.Error(err))
	}
	asynqClient := asynq.NewClient(asynqRedis)
	defer asynqClient.Close()

	// JWT
	jwtMgr := lib.NewJWTManager(cfg.JWTSecret)
	middleware.SetDefaultJWT(jwtMgr)

	// Repositories
	userRepo := repository.NewUserRepository(pg.Pool)
	sessionRepo := repository.NewSessionRepository(pg.Pool)
	planRepo := repository.NewPlanRepository(pg.Pool)
	orderRepo := repository.NewOrderRepository(pg.Pool)
	paymentRepo := repository.NewPaymentRepository(pg.Pool)
	testRepo := repository.NewTestRepository(pg.Pool)

	// Services
	authSvc := service.NewAuthService(userRepo, sessionRepo, jwtMgr)
	stripe := service.NewStripeClient(logger)
	orderSvc := service.NewOrderService(orderRepo, planRepo, testRepo, paymentRepo, stripe, asynqClient, logger)

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), lib.RequestLogger(logger))

	corsConfig := cors.Config{
		AllowOrigins:     []string{"https://" + cfg.Domain, "https://www." + cfg.Domain},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	if cfg.AppEnv == "development" {
		corsConfig.AllowOrigins = []string{"http://localhost:3000"}
	}
	r.Use(cors.New(corsConfig))

	// Health
	healthHandler := handler.NewHealthHandler(pg.Pool, rdb)
	r.GET("/health", healthHandler.Health)
	r.GET("/liveness", healthHandler.Liveness)

	// Auth
	authH := handler.NewAuthHandler(authSvc, logger)
	authH.Register(r)

	// Plans, Orders
	orderH := handler.NewOrderHandler(orderSvc, stripe, asynqClient, logger)
	orderH.Register(r)

	// Tests
	testH := handler.NewTestHandler(pg.Pool, logger)
	testH.Register(r)

	// Admin
	adminH := handler.NewAdminHandler(pg.Pool, logger)
	adminH.Register(r)

	// Activity (orchestrator'dan)
	activityH := handler.NewActivityHandler(pg.Pool, cfg.OrchestratorAPIToken, logger)
	activityH.Register(r)

	// Legacy ping
	r.GET("/api/v1/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		logger.Info("api server starting", zap.String("port", cfg.Port), zap.String("env", cfg.AppEnv))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("forced shutdown", zap.Error(err))
	}
	logger.Info("server exited")
}
