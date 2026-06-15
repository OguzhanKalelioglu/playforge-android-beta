package lib

import (
	"context"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type AsynqLogger struct {
	Logger *zap.Logger
}

func (a *AsynqLogger) Debug(args ...interface{}) { a.Logger.Sugar().Debug(args...) }
func (a *AsynqLogger) Info(args ...interface{})  { a.Logger.Sugar().Info(args...) }
func (a *AsynqLogger) Warn(args ...interface{})  { a.Logger.Sugar().Warn(args...) }
func (a *AsynqLogger) Error(args ...interface{}) { a.Logger.Sugar().Error(args...) }
func (a *AsynqLogger) Fatal(args ...interface{}) { a.Logger.Sugar().Fatal(args...) }

func LoggingMiddleware(logger *zap.Logger) asynq.MiddlewareFunc {
	return func(h asynq.Handler) asynq.Handler {
		return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
			logger.Info("task started",
				zap.String("type", t.Type()),
				zap.Int("payload_size", len(t.Payload())))
			err := h.ProcessTask(ctx, t)
			if err != nil {
				logger.Error("task failed",
					zap.String("type", t.Type()),
					zap.Error(err))
				return err
			}
			logger.Info("task completed", zap.String("type", t.Type()))
			return nil
		})
	}
}
