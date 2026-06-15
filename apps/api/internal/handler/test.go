package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/middleware"
	"github.com/testerscommunity/api/internal/repository"
)

type TestHandler struct {
	db     *pgxpool.Pool
	tests  *repository.TestRepository
	logger *zap.Logger
}

func NewTestHandler(db *pgxpool.Pool, logger *zap.Logger) *TestHandler {
	return &TestHandler{db: db, tests: repository.NewTestRepository(db), logger: logger}
}

func (h *TestHandler) Register(r *gin.Engine) {
	auth := r.Group("/api/v1", middleware.AuthRequiredJWT())
	auth.GET("/tests", h.List)
	auth.GET("/tests/:id", h.Detail)
	auth.GET("/tests/:id/activity", h.Activity)
	auth.GET("/tests/:id/reviews", h.Reviews)
}

func (h *TestHandler) List(c *gin.Context) {
	uid, _ := uuid.Parse(c.GetString(middleware.CtxUserID))
	role := c.GetString(middleware.CtxRole)

	var tests []*repository.Test
	var err error
	if role == "admin" {
		tests, err = h.tests.ListAll(c.Request.Context(), "", 100)
	} else {
		tests, err = h.tests.ListByUser(c.Request.Context(), uid)
	}
	if err != nil {
		h.logger.Error("list tests failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load failed"})
		return
	}

	out := make([]gin.H, 0, len(tests))
	for _, t := range tests {
		progress, _ := h.tests.Progress(c.Request.Context(), t.ID)
		out = append(out, gin.H{
			"id":            t.ID,
			"package_name":  t.PackageName,
			"test_link":     t.TestLink,
			"status":        t.Status,
			"started_at":    t.StartedAt,
			"ends_at":       t.EndsAt,
			"created_at":    t.CreatedAt,
			"progress":      progress,
		})
	}
	c.JSON(http.StatusOK, out)
}

func (h *TestHandler) Detail(c *gin.Context) {
	uid, _ := uuid.Parse(c.GetString(middleware.CtxUserID))
	role := c.GetString(middleware.CtxRole)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	t, err := h.tests.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load failed"})
		return
	}
	if role != "admin" && t.UserID != uid {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	assignments, _ := h.tests.ListAssignments(c.Request.Context(), t.ID)
	progress, _ := h.tests.Progress(c.Request.Context(), t.ID)

	// Tester email'leri için join
	emails := map[uuid.UUID]string{}
	if len(assignments) > 0 {
		const q = `SELECT id, email FROM testers WHERE id = ANY($1)`
		ids := make([]uuid.UUID, len(assignments))
		for i, a := range assignments {
			ids[i] = a.TesterID
		}
		rows, err := h.db.Query(c.Request.Context(), q, ids)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id uuid.UUID
				var email string
				if err := rows.Scan(&id, &email); err == nil {
					emails[id] = email
				}
			}
		}
	}

	asOut := make([]gin.H, 0, len(assignments))
	for _, a := range assignments {
		asOut = append(asOut, gin.H{
			"id":                 a.ID,
			"tester_email":       emails[a.TesterID],
			"status":             a.Status,
			"opt_in_at":          a.OptInAt,
			"install_at":         a.InstallAt,
			"last_engagement_at": a.LastEngagementAt,
			"error_message":      a.ErrorMessage,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           t.ID,
		"package_name": t.PackageName,
		"test_link":    t.TestLink,
		"status":       t.Status,
		"started_at":   t.StartedAt,
		"ends_at":      t.EndsAt,
		"created_at":   t.CreatedAt,
		"assignments":  asOut,
		"progress":     progress,
	})
}

func (h *TestHandler) Activity(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	rows, err := h.tests.ListActivity(c.Request.Context(), id, 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load failed"})
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		var meta map[string]any
		_ = json.Unmarshal(r.Metadata, &meta)
		out = append(out, gin.H{
			"id":                r.ID,
			"assignment_id":     r.TestAssignmentID,
			"action":            r.Action,
			"performed_at":      r.PerformedAt,
			"success":           r.Success,
			"error_message":     r.ErrorMessage,
			"screenshot_path":   r.ScreenshotPath,
			"metadata":          meta,
		})
	}
	c.JSON(http.StatusOK, out)
}

func (h *TestHandler) Reviews(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	reviews, err := h.tests.ListReviews(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load failed"})
		return
	}
	out := make([]gin.H, 0, len(reviews))
	for _, r := range reviews {
		out = append(out, gin.H{
			"id":          r.ID,
			"rating":      r.Rating,
			"review_text": r.ReviewText,
			"language":    r.Language,
			"posted_at":   r.PostedAt,
		})
	}
	c.JSON(http.StatusOK, out)
}
