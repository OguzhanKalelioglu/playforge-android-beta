package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/middleware"
	"github.com/testerscommunity/api/internal/repository"
)

type AdminHandler struct {
	db      *pgxpool.Pool
	tests   *repository.TestRepository
	orders  *repository.OrderRepository
	testers *repository.TesterRepository
	logger  *zap.Logger
}

func NewAdminHandler(db *pgxpool.Pool, logger *zap.Logger) *AdminHandler {
	return &AdminHandler{
		db:      db,
		tests:   repository.NewTestRepository(db),
		orders:  repository.NewOrderRepository(db),
		testers: repository.NewTesterRepository(db),
		logger:  logger,
	}
}

func (h *AdminHandler) Register(r *gin.Engine) {
	g := r.Group("/api/v1/admin",
		middleware.AuthRequiredJWT(),
		middleware.AdminOnly(),
	)
	g.GET("/metrics", h.Metrics)
	g.GET("/orders", h.Orders)
	g.GET("/tests", h.Tests)
	g.GET("/testers", h.Testers)
	g.GET("/payments", h.Payments)
}

func (h *AdminHandler) Metrics(c *gin.Context) {
	ctx := c.Request.Context()
	var (
		activeTests    int
		pendingTests   int
		totalTesters   int
		activeTesters  int
		warmingTesters int
		revenue        float64
		failed24h      int
	)

	_ = h.db.QueryRow(ctx,
		`SELECT COUNT(*) FILTER (WHERE status='active'),
		        COUNT(*) FILTER (WHERE status='pending')
		 FROM tests`).Scan(&activeTests, &pendingTests)

	_ = h.db.QueryRow(ctx,
		`SELECT COUNT(*),
		        COUNT(*) FILTER (WHERE status='active'),
		        COUNT(*) FILTER (WHERE status='warming')
		 FROM testers`).Scan(&totalTesters, &activeTesters, &warmingTesters)

	_ = h.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0)
		 FROM payments
		 WHERE status='completed' AND paid_at > NOW() - INTERVAL '30 days'`).
		Scan(&revenue)

	_ = h.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM activity_logs
		 WHERE success = FALSE AND performed_at > NOW() - INTERVAL '24 hours'`).
		Scan(&failed24h)

	c.JSON(http.StatusOK, gin.H{
		"active_tests":      activeTests,
		"pending_tests":     pendingTests,
		"total_testers":     totalTesters,
		"active_testers":    activeTesters,
		"warming_testers":   warmingTesters,
		"emulators_ready":   25, // orchestrator'dan canlı çekilebilir; burada placeholder
		"emulators_total":   25,
		"revenue_month_try": revenue,
		"failed_tasks_24h":  failed24h,
	})
}

func (h *AdminHandler) Orders(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT o.id, u.email, o.created_at, o.status, o.total, o.currency,
		       pt.name, o.test_id
		FROM orders o
		JOIN users u ON u.id = o.user_id
		JOIN plan_tiers pt ON pt.id = o.plan_tier_id
		ORDER BY o.created_at DESC LIMIT 200
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load failed"})
		return
	}
	defer rows.Close()

	out := []gin.H{}
	for rows.Next() {
		var id, testID any
		var userEmail, status, currency, planName string
		var total float64
		var createdAt any
		if err := rows.Scan(&id, &userEmail, &createdAt, &status, &total, &currency, &planName, &testID); err != nil {
			continue
		}
		// Test package name (eğer varsa)
		var pkgName *string
		if testID != nil {
			_ = h.db.QueryRow(c.Request.Context(), `SELECT package_name FROM tests WHERE id = $1`, testID).Scan(&pkgName)
		}
		out = append(out, gin.H{
			"id":           id,
			"user_email":   userEmail,
			"package_name": pkgName,
			"plan_name":    planName,
			"status":       status,
			"total":        total,
			"currency":     currency,
			"created_at":   createdAt,
		})
	}
	c.JSON(http.StatusOK, out)
}

func (h *AdminHandler) Tests(c *gin.Context) {
	tests, err := h.tests.ListAll(c.Request.Context(), "", 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load failed"})
		return
	}
	out := []gin.H{}
	for _, t := range tests {
		progress, _ := h.tests.Progress(c.Request.Context(), t.ID)
		// User email
		var email string
		_ = h.db.QueryRow(c.Request.Context(), `SELECT email FROM users WHERE id = $1`, t.UserID).Scan(&email)
		out = append(out, gin.H{
			"id":           t.ID,
			"user_email":   email,
			"package_name": t.PackageName,
			"status":       t.Status,
			"progress":     progress,
			"started_at":   t.StartedAt,
			"ends_at":      t.EndsAt,
			"created_at":   t.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, out)
}

func (h *AdminHandler) Testers(c *gin.Context) {
	rows, err := h.testers.ListAdmin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load failed"})
		return
	}
	out := []gin.H{}
	for _, t := range rows {
		out = append(out, gin.H{
			"id":            t.ID,
			"email":         t.Email,
			"status":        t.Status,
			"group_email":   t.GroupEmail,
			"device_model":  t.DeviceModel,
			"last_used_at":  t.LastUsedAt,
			"tasks_completed_30d": t.TasksCompleted30d,
		})
	}
	c.JSON(http.StatusOK, out)
}

func (h *AdminHandler) Payments(c *gin.Context) {
	rows, err := h.db.Query(c.Request.Context(), `
		SELECT p.id, u.email, p.amount, p.currency, p.status,
		       p.iyzico_payment_id, p.paid_at, p.created_at
		FROM payments p
		JOIN users u ON u.id = p.user_id
		ORDER BY p.created_at DESC LIMIT 200
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load failed"})
		return
	}
	defer rows.Close()

	out := []gin.H{}
	for rows.Next() {
		var id, amount any
		var email, status, currency string
		var iyzicoID, paidAt any
		var createdAt any
		if err := rows.Scan(&id, &email, &amount, &currency, &status, &iyzicoID, &paidAt, &createdAt); err != nil {
			continue
		}
		out = append(out, gin.H{
			"id":                id,
			"user_email":        email,
			"amount":            amount,
			"currency":          currency,
			"status":            status,
			"iyzico_payment_id": iyzicoID,
			"paid_at":           paidAt,
			"created_at":        createdAt,
		})
	}
	c.JSON(http.StatusOK, out)
}
