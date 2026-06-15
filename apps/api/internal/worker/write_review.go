package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/model"
)

// WriteReviewHandler, gün 14 review yazma task'ı
type WriteReviewHandler struct {
	Runner *RunnerClient
	Logger *zap.Logger
}

func NewWriteReviewHandler(runner *RunnerClient, logger *zap.Logger) *WriteReviewHandler {
	return &WriteReviewHandler{Runner: runner, Logger: logger}
}

func (h *WriteReviewHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	p, err := model.DecodeWriteReview(t.Payload())
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	h.Logger.Info("write_review task started",
		zap.String("test_id", p.TestID),
		zap.String("assignment_id", p.AssignmentID),
		zap.Int("stars", p.Stars),
		zap.String("language", p.Language))

	if err := h.Runner.StartReview(ctx, p); err != nil {
		h.Logger.Error("review runner call failed",
			zap.String("test_id", p.TestID),
			zap.Error(err))
		return err
	}

	h.Logger.Info("write_review task dispatched",
		zap.String("test_id", p.TestID))
	return nil
}
