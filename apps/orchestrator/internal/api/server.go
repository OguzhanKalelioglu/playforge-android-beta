package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/testerscommunity/orchestrator/internal/emulator"
)

type Server struct {
	Pool      *emulator.Pool
	Logger    *zap.Logger
	APIToken  string
	StartedAt time.Time
}

func NewServer(pool *emulator.Pool, logger *zap.Logger, token string) *Server {
	return &Server{
		Pool:      pool,
		Logger:    logger,
		APIToken:  token,
		StartedAt: time.Now(),
	}
}

func (s *Server) Register(r *gin.Engine) {
	r.GET("/health", s.Health)
	r.GET("/liveness", s.Liveness)
	r.GET("/emulators", s.ListEmulators)
	r.GET("/emulators/status", s.EmulatorStatus)

	auth := r.Group("/", s.authMiddleware())
	{
		auth.POST("/tests/:id/start", s.StartTest)
		auth.POST("/tests/:id/stop", s.StopTest)
		auth.GET("/tests/:id/status", s.TestStatus)
	}
}

func (s *Server) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":     "ok",
		"service":    "orchestrator",
		"uptime":     time.Since(s.StartedAt).String(),
		"emulators":  len(s.Pool.List()),
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}

type EmulatorDTO struct {
	Serial  string `json:"serial"`
	Index   int    `json:"index"`
	Status  string `json:"status"`
	Tester  string `json:"tester_id,omitempty"`
}

func (s *Server) ListEmulators(c *gin.Context) {
	list := s.Pool.List()
	out := make([]EmulatorDTO, 0, len(list))
	for _, e := range list {
		out = append(out, EmulatorDTO{
			Serial: e.Serial,
			Index:  e.Index,
			Status: string(e.Status),
			Tester: e.TesterID,
		})
	}
	c.JSON(http.StatusOK, gin.H{"emulators": out})
}

func (s *Server) EmulatorStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": s.Pool.Status(),
		"counts": s.countByStatus(),
	})
}

func (s *Server) countByStatus() map[string]int {
	counts := map[string]int{
		"idle": 0, "busy": 0, "booting": 0, "offline": 0,
	}
	for _, e := range s.Pool.List() {
		counts[string(e.Status)]++
	}
	return counts
}

func (s *Server) StartTest(c *gin.Context) {
	testID := c.Param("id")
	s.Logger.Info("test start requested", zap.String("test_id", testID))
	c.JSON(http.StatusAccepted, gin.H{
		"message": "test start queued",
		"test_id": testID,
	})
}

func (s *Server) StopTest(c *gin.Context) {
	testID := c.Param("id")
	s.Logger.Info("test stop requested", zap.String("test_id", testID))
	c.JSON(http.StatusOK, gin.H{
		"message": "test stop queued",
		"test_id": testID,
	})
}

func (s *Server) TestStatus(c *gin.Context) {
	testID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"test_id":  testID,
		"status":   "not_implemented",
		"progress": 0,
	})
}

func (s *Server) authMiddleware() gin.HandlerFunc {
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
