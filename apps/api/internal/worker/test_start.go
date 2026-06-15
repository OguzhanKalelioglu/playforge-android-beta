package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/model"
)

// TestStartHandler, opt-in + download + initial engage task'ını yürütür
type TestStartHandler struct {
	Runner *RunnerClient
	Logger *zap.Logger
}

func NewTestStartHandler(runner *RunnerClient, logger *zap.Logger) *TestStartHandler {
	return &TestStartHandler{Runner: runner, Logger: logger}
}

// ProcessTask, Asynq'nun çağırdığı ana method
func (h *TestStartHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	p, err := model.DecodeTestStart(t.Payload())
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	h.Logger.Info("test_start task started",
		zap.String("test_id", p.TestID),
		zap.String("assignment_id", p.AssignmentID),
		zap.String("package", p.PackageName))

	if err := h.Runner.StartTest(ctx, p); err != nil {
		h.Logger.Error("test_start runner call failed",
			zap.String("test_id", p.TestID),
			zap.Error(err))
		return err // Asynq retry'ı tetikler
	}

	h.Logger.Info("test_start task dispatched",
		zap.String("test_id", p.TestID))
	return nil
}
