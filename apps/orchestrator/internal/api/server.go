package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/testerscommunity/orchestrator/internal/emulator"
	"github.com/testerscommunity/orchestrator/internal/lifecycle"
)

type Server struct {
	Pool        *emulator.Pool
	Lifecycle   *lifecycle.Manager
	TaskHandler *TaskHandler
	Logger      *zap.Logger
	APIToken    string
	StartedAt   time.Time
}

func NewServer(pool *emulator.Pool, lifecycle *lifecycle.Manager, taskHandler *TaskHandler, logger *zap.Logger, token string) *Server {
	return &Server{
		Pool:        pool,
		Lifecycle:   lifecycle,
		TaskHandler: taskHandler,
		Logger:      logger,
		APIToken:    token,
		StartedAt:   time.Now(),
	}
}

func (s *Server) Register(r *gin.Engine) {
	r.GET("/health", s.Health)
	r.GET("/liveness", s.Liveness)

	r.GET("/emulators", s.ListEmulators)
	r.GET("/emulators/status", s.EmulatorStatus)
	r.GET("/emulators/counts", s.EmulatorCounts)
	r.GET("/emulators/:serial", s.GetEmulator)

	auth := r.Group("/", s.AuthMiddleware())
	{
		auth.POST("/emulators/start-all", s.StartAll)
		auth.POST("/emulators/stop-all", s.StopAll)
		auth.POST("/emulators/:serial/start", s.StartEmulator)
		auth.POST("/emulators/:serial/stop", s.StopEmulator)
		auth.POST("/emulators/:serial/restart", s.RestartEmulator)
		auth.POST("/emulators/:serial/wipe", s.WipeEmulator)
		auth.POST("/emulators/:serial/reset", s.ResetForTest)
	}
}

func (s *Server) Health(c *gin.Context) {
	counts := s.Pool.Counts()
	ready := counts[emulator.StatusReady]
	total := s.Pool.Count()
	booting := counts[emulator.StatusBooting]
	errored := counts[emulator.StatusError]

	status := "ok"
	if errored > 0 {
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     status,
		"service":    "orchestrator",
		"uptime":     time.Since(s.StartedAt).String(),
		"emulators":  gin.H{"total": total, "ready": ready, "booting": booting, "error": errored},
		"counts":     counts,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}

type EmulatorDTO struct {
	Serial      string `json:"serial"`
	Index       int    `json:"index"`
	Status      string `json:"status"`
	Tester      string `json:"tester_id,omitempty"`
	ContainerID string `json:"container_id,omitempty"`
	BootedAt    string `json:"booted_at,omitempty"`
	LastCheck   string `json:"last_check,omitempty"`
	BootCount   int    `json:"boot_count"`
	ErrorMsg    string `json:"error_msg,omitempty"`
}

func (s *Server) ListEmulators(c *gin.Context) {
	list := s.Pool.List()
	out := make([]EmulatorDTO, 0, len(list))
	for _, e := range list {
		out = append(out, s.toDTO(e))
	}
	c.JSON(http.StatusOK, gin.H{
		"emulators": out,
		"total":     len(out),
	})
}

func (s *Server) EmulatorStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": s.Pool.Status(),
		"counts": s.Pool.Counts(),
	})
}

func (s *Server) EmulatorCounts(c *gin.Context) {
	counts := s.Pool.Counts()
	c.JSON(http.StatusOK, gin.H{
		"total":  s.Pool.Count(),
		"counts": counts,
	})
}

func (s *Server) GetEmulator(c *gin.Context) {
	serial := c.Param("serial")
	e, err := s.Pool.Get(serial)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s.toDTO(e))
}

func (s *Server) StartAll(c *gin.Context) {
	s.Logger.Info("API: start all emulators")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		if err := s.Lifecycle.StartAllEmulators(ctx); err != nil {
			s.Logger.Error("start-all failed", zap.Error(err))
		}
	}()
	c.JSON(http.StatusAccepted, gin.H{
		"message": "all emulators starting",
		"count":   s.Pool.Count(),
	})
}

func (s *Server) StopAll(c *gin.Context) {
	s.Logger.Warn("API: stop all emulators")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := s.Lifecycle.StopAllEmulators(ctx); err != nil {
			s.Logger.Error("stop-all failed", zap.Error(err))
		}
	}()
	c.JSON(http.StatusAccepted, gin.H{"message": "all emulators stopping"})
}

func (s *Server) StartEmulator(c *gin.Context) {
	serial := c.Param("serial")
	s.Logger.Info("API: start emulator", zap.String("serial", serial))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := s.Lifecycle.StartEmulator(ctx, serial); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"message": "emulator starting",
		"serial":  serial,
	})
}

func (s *Server) StopEmulator(c *gin.Context) {
	serial := c.Param("serial")
	s.Logger.Info("API: stop emulator", zap.String("serial", serial))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := s.Lifecycle.StopEmulator(ctx, serial); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "emulator stopped",
		"serial":  serial,
	})
}

func (s *Server) RestartEmulator(c *gin.Context) {
	serial := c.Param("serial")
	s.Logger.Info("API: restart emulator", zap.String("serial", serial))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := s.Lifecycle.RestartEmulator(ctx, serial); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"message": "emulator restarting",
		"serial":  serial,
	})
}

func (s *Server) WipeEmulator(c *gin.Context) {
	serial := c.Param("serial")
	s.Logger.Warn("API: wipe emulator (factory reset)", zap.String("serial", serial))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	if err := s.Lifecycle.WipeEmulator(ctx, serial); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "emulator wiped",
		"serial":  serial,
	})
}

func (s *Server) ResetForTest(c *gin.Context) {
	serial := c.Param("serial")
	s.Logger.Info("API: reset for test", zap.String("serial", serial))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	if err := s.Lifecycle.ResetForTest(ctx, serial); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "emulator reset for test",
		"serial":  serial,
	})
}

func (s *Server) toDTO(e *emulator.Emulator) EmulatorDTO {
	dto := EmulatorDTO{
		Serial:      e.Serial,
		Index:       e.Index,
		Status:      string(e.Status),
		Tester:      e.TesterID,
		ContainerID: e.ContainerID,
		BootCount:   e.BootCount,
		ErrorMsg:    e.ErrorMsg,
	}
	if !e.BootedAt.IsZero() {
		dto.BootedAt = e.BootedAt.UTC().Format(time.RFC3339)
	}
	if !e.LastCheck.IsZero() {
		dto.LastCheck = e.LastCheck.UTC().Format(time.RFC3339)
	}
	return dto
}

func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-API-Token")
		if token == "" {
			token = c.Query("token")
		}
		if token != s.APIToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or missing API token",
			})
			return
		}
		c.Next()
	}
}

func parseIndex(s string) (int, error) {
	return strconv.Atoi(s)
}
