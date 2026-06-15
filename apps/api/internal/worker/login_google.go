package worker

import (
	"context"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/model"
)

// LoginGoogleHandler, yeni emulator'de Google hesabı ekleme
// En zor task: 5x retry (warming sonrası gerekebilir)
type LoginGoogleHandler struct {
	Runner *RunnerClient
	Logger *zap.Logger
}

func NewLoginGoogleHandler(runner *RunnerClient, logger *zap.Logger) *LoginGoogleHandler {
	return &LoginGoogleHandler{Runner: runner, Logger: logger}
}

func (h *LoginGoogleHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	p, err := model.DecodeLoginGoogle(t.Payload())
	if err != nil {
		return fmt.Errorf("decode payload: %w", err)
	}

	h.Logger.Info("login_google task started",
		zap.String("test_id", p.TestID),
		zap.String("assignment_id", p.AssignmentID),
		zap.String("email", maskEmail(p.Email)))

	if err := h.Runner.StartLoginGoogle(ctx, p); err != nil {
		h.Logger.Error("login_google runner call failed",
			zap.String("test_id", p.TestID),
			zap.Error(err))
		return err
	}

	h.Logger.Info("login_google task dispatched",
		zap.String("test_id", p.TestID))
	return nil
}

func maskEmail(email string) string {
	for i, c := range email {
		if c == '@' {
			return email[:i/2] + "***" + email[i:]
		}
	}
	return email
}
