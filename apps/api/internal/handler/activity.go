package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/model"
)

// ActivityHandler, orchestrator'dan gelen step event'lerini DB'ye yazar
// POST /v1/activity (X-Activity-Token header ile auth)
type ActivityHandler struct {
	DB     *pgxpool.Pool
	Token  string
	Logger *zap.Logger
}

func NewActivityHandler(db *pgxpool.Pool, token string, logger *zap.Logger) *ActivityHandler {
	return &ActivityHandler{DB: db, Token: token, Logger: logger}
}

func (h *ActivityHandler) Register(r *gin.Engine) {
	r.POST("/v1/activity", h.authMiddleware(), h.Ingest)
}

func (h *ActivityHandler) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Activity-Token")
		if h.Token != "" && token != h.Token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Next()
	}
}

type batchRequest struct {
	Events []model.ActivityEvent `json:"events"`
}

// Ingest, batch event'leri activity_logs tablosuna yazar
func (h *ActivityHandler) Ingest(c *gin.Context) {
	var req batchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: " + err.Error()})
		return
	}
	if len(req.Events) == 0 {
		c.JSON(http.StatusOK, gin.H{"inserted": 0})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	inserted, err := h.insertBatch(ctx, req.Events)
	if err != nil {
		h.Logger.Error("activity insert failed", zap.Error(err), zap.Int("count", len(req.Events)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
		return
	}

	h.Logger.Debug("activity events inserted", zap.Int("count", inserted))
	c.JSON(http.StatusOK, gin.H{
		"inserted": inserted,
		"received": len(req.Events),
	})
}

func (h *ActivityHandler) insertBatch(ctx context.Context, events []model.ActivityEvent) (int, error) {
	tx, err := h.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	const q = `
		INSERT INTO activity_logs (
			test_assignment_id, action, performed_at, success,
			error_message, screenshot_path, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	inserted := 0
	for _, ev := range events {
		assignmentID := h.resolveAssignmentID(ctx, tx, ev.AssignmentID)

		success := ev.Status == "ok" || ev.Status == "completed" || ev.Status == string(model.EventStepComplete)
		metadata := ev.Metadata
		if len(metadata) == 0 {
			metadata = []byte(`{}`)
		}

		occurredAt := ev.OccurredAt
		if occurredAt.IsZero() {
			occurredAt = time.Now().UTC()
		}

		_, err = tx.Exec(ctx, q,
			assignmentID,
			ev.EventType,
			occurredAt,
			success,
			nilIfEmpty(ev.ErrorMessage),
			nilIfEmpty(ev.ScreenshotPath),
			metadata,
		)
		if err != nil {
			return inserted, err
		}
		inserted++
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return inserted, nil
}

// resolveAssignmentID, gelen assignment_id string'ini UUID'ye çevirir
// Geçersizse veya bulunamazsa nil (NULL) döner — log yine de kaydedilir
func (h *ActivityHandler) resolveAssignmentID(ctx context.Context, tx pgx.Tx, id string) any {
	if id == "" {
		return nil
	}

	parsed, err := uuid.Parse(id)
	if err != nil {
		h.Logger.Debug("assignment_id not a valid UUID", zap.String("id", id))
		return nil
	}

	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM test_assignments WHERE id = $1)", parsed).Scan(&exists)
	if err != nil {
		h.Logger.Warn("assignment existence check failed", zap.String("id", id), zap.Error(err))
		return nil
	}
	if !exists {
		return nil
	}
	return parsed
}

func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}
