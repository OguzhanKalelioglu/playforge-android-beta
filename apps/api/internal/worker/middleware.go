package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// LoggingMiddleware, Asynq task başlangıç/bitiş log'larını yazar
func LoggingMiddleware(logger *zap.Logger) asynq.MiddlewareFunc {
	return func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			start := time.Now()
			logger.Info("task processing started",
				zap.String("type", t.Type()),
				zap.Int("payload_size", len(t.Payload())))

			err := h.ProcessTask(ctx, t)
			duration := time.Since(start)

			if err != nil {
				logger.Error("task processing failed",
					zap.String("type", t.Type()),
					zap.Duration("duration", duration),
					zap.Error(err))
				return err
			}

			logger.Info("task processing completed",
				zap.String("type", t.Type()),
				zap.Duration("duration", duration))
			return nil
		})
	}
}

// RecoveryMiddleware, panic olursa yakalar ve error olarak döndürür
func RecoveryMiddleware(logger *zap.Logger) asynq.MiddlewareFunc {
	return func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) (err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("task panic recovered",
						zap.String("type", t.Type()),
						zap.Any("panic", r))
					err = fmt.Errorf("task panicked: %v", r)
				}
			}()
			return h.ProcessTask(ctx, t)
		})
	}
}

// TimeoutMiddleware, task'a timeout uygular
// Asynq'nun ProcessIn ile birlikte çalışır
func TimeoutMiddleware(timeout time.Duration) asynq.MiddlewareFunc {
	return func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			return h.ProcessTask(ctx, t)
		})
	}
}
