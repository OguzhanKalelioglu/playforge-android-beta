package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/testerscommunity/orchestrator/internal/task"
	"github.com/testerscommunity/orchestrator/internal/taskrunner"
)

// TaskHandler, VPS API'den gelen task çağrılarını handle eder
// POST /v1/tasks/:type/start → task'ı çalıştırır, async result döner
type TaskHandler struct {
	Runner *taskrunner.Runner
	Logger *zap.Logger
}

func NewTaskHandler(runner *taskrunner.Runner, logger *zap.Logger) *TaskHandler {
	return &TaskHandler{Runner: runner, Logger: logger}
}

// Register, task endpoint'lerini gin router'a ekler
func (h *TaskHandler) Register(r *gin.Engine, authMW gin.HandlerFunc) {
	v1 := r.Group("/v1/tasks", authMW)
	{
		v1.POST("/opt_in/start", h.handleOptIn)
		v1.POST("/download/start", h.handleDownload)
		v1.POST("/engage/start", h.handleEngage)
		v1.POST("/review/start", h.handleReview)
		v1.POST("/login_google/start", h.handleLoginGoogle)
	}
}

// Ortak payload: test_id, assignment_id, package_name
type baseTaskPayload struct {
	TestID       string `json:"test_id" binding:"required"`
	AssignmentID string `json:"assignment_id" binding:"required"`
	PackageName  string `json:"package_name"`
	// Task-specific extra fields (opsiyonel)
	Day     *int   `json:"day,omitempty"`
	Stars   *int   `json:"stars,omitempty"`
	Comment string `json:"comment,omitempty"`
	Email   string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	TesterID string `json:"tester_id,omitempty"`
	TestLink string `json:"test_link,omitempty"`
}

func (h *TaskHandler) parse(c *gin.Context) (baseTaskPayload, error) {
	var p baseTaskPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		return p, err
	}
	return p, nil
}

func (h *TaskHandler) submit(c *gin.Context, t task.Task, p baseTaskPayload) {
	h.Logger.Info("task submit",
		zap.String("type", t.Type()),
		zap.String("test_id", p.TestID),
		zap.String("assignment_id", p.AssignmentID))

	resCh, err := h.Runner.Submit(c.Request.Context(), t, taskrunner.SubmitParams{
		TestID:       p.TestID,
		AssignmentID: p.AssignmentID,
		PackageName:  p.PackageName,
	})
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	// Async: hemen 202 döndür, result channel üzerinden client poll eder
	c.JSON(http.StatusAccepted, gin.H{
		"status":        "accepted",
		"task_type":     t.Type(),
		"test_id":       p.TestID,
		"assignment_id": p.AssignmentID,
		"message":       "task queued, will be processed asynchronously",
	})

	// Background'da sonucu logla
	go func() {
		res := <-resCh
		if res.Err != nil {
			h.Logger.Error("task failed",
				zap.String("type", t.Type()),
				zap.String("assignment_id", p.AssignmentID),
				zap.Error(res.Err))
		} else {
			h.Logger.Info("task succeeded",
				zap.String("type", t.Type()),
				zap.String("assignment_id", p.AssignmentID),
				zap.Duration("duration", res.Duration))
		}
	}()
}

func (h *TaskHandler) handleOptIn(c *gin.Context) {
	p, err := h.parse(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.submit(c, &task.OptInTask{}, p)
}

func (h *TaskHandler) handleDownload(c *gin.Context) {
	p, err := h.parse(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.submit(c, &task.DownloadTask{}, p)
}

func (h *TaskHandler) handleEngage(c *gin.Context) {
	p, err := h.parse(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	day := 1
	if p.Day != nil {
		day = *p.Day
	}
	h.submit(c, &task.EngageTask{Day: day}, p)
}

func (h *TaskHandler) handleReview(c *gin.Context) {
	p, err := h.parse(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	stars := 5
	if p.Stars != nil {
		stars = *p.Stars
	}
	h.submit(c, &task.ReviewTask{
		Stars:   stars,
		Comment: p.Comment,
	}, p)
}

func (h *TaskHandler) handleLoginGoogle(c *gin.Context) {
	p, err := h.parse(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if p.Email == "" || p.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password required for login_google"})
		return
	}
	h.submit(c, &task.LoginGoogleTask{
		Email:    p.Email,
		Password: p.Password,
	}, p)
}

// healthStatus, runtime response status kontrolü için
func jsonResponseOK(c *gin.Context, data interface{}) {
	if strings.Contains(c.GetHeader("Accept"), "json") {
		c.JSON(http.StatusOK, data)
		return
	}
	jsonBytes, _ := json.Marshal(data)
	c.Data(http.StatusOK, "application/json", jsonBytes)
}
