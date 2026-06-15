package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/model"
)

// DailyEngagementHandler, günlük 2-5dk engagement task'ı
type DailyEngagementHandler struct {
	Runner *RunnerClient
	Logger *zap.Logger
}

func NewDailyEngagementHandler(runner *RunnerClient, logger *zap.Logger) *DailyEngagementHandler {
	return &DailyEngagementHandler{Runner: runner, Logger: logger}
}

func (h *DailyEngagementHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	p, err := model.DecodeDailyEngagement(t.Payload())
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	h.Logger.Info("daily_engagement task started",
		zap.String("test_id", p.TestID),
		zap.String("assignment_id", p.AssignmentID),
		zap.Int("day", p.Day),
		zap.String("package", p.PackageName))

	if err := h.Runner.StartEngagement(ctx, p); err != nil {
		h.Logger.Error("engagement runner call failed",
			zap.String("test_id", p.TestID),
			zap.Int("day", p.Day),
			zap.Error(err))
		return err
	}

	h.Logger.Info("daily_engagement task dispatched",
		zap.String("test_id", p.TestID),
		zap.Int("day", p.Day))
	return nil
}
