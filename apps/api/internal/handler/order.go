package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/middleware"
	"github.com/testerscommunity/api/internal/model"
	"github.com/testerscommunity/api/internal/service"
)

type OrderHandler struct {
	orders     *service.OrderService
	logger     *zap.Logger
	asynq      *asynq.Client
}

func NewOrderHandler(orders *service.OrderService, asynq *asynq.Client, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{orders: orders, asynq: asynq, logger: logger}
}

func (h *OrderHandler) Register(r *gin.Engine) {
	api := r.Group("/api/v1")
	api.GET("/plans", h.ListPlans)

	auth := api.Group("", middleware.AuthRequiredJWT())
	auth.POST("/orders", h.Create)
	auth.GET("/orders", h.List)
	auth.GET("/orders/:id", h.Detail)

	// Iyzico callback — public, token-based
	r.POST("/api/v1/payments/iyzico/callback", h.IyzicoCallback)
}

func (h *OrderHandler) ListPlans(c *gin.Context) {
	plans, err := h.orders.GetPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "plans load failed"})
		return
	}
	out := make([]gin.H, 0, len(plans))
	for _, p := range plans {
		var features []string
		_ = json.Unmarshal(p.Features, &features)
		out = append(out, gin.H{
			"id":             p.ID,
			"slug":           p.Slug,
			"name":           p.Name,
			"description":    p.Description,
			"tester_count":   p.TesterCount,
			"duration_days":  p.DurationDays,
			"price_try":      p.PriceTRY,
			"price_usd":      p.PriceUSD,
			"features":       features,
			"sort_order":     p.SortOrder,
		})
	}
	c.JSON(http.StatusOK, out)
}

type createOrderReq struct {
	PlanSlug       string `json:"plan_slug"`
	PackageName    string `json:"package_name"`
	TestLink       string `json:"test_link"`
	BillingEmail   string `json:"billing_email"`
	BillingName    string `json:"billing_name"`
	BillingPhone   string `json:"billing_phone"`
	BillingAddress string `json:"billing_address"`
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if req.PlanSlug == "" || req.PackageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "plan_slug and package_name required"})
		return
	}

	uid, _ := uuid.Parse(c.GetString(middleware.CtxUserID))

	result, err := h.orders.Create(c.Request.Context(), service.CreateOrderInput{
		UserID:         uid,
		PlanSlug:       req.PlanSlug,
		PackageName:    req.PackageName,
		TestLink:       req.TestLink,
		BillingEmail:   req.BillingEmail,
		BillingName:    req.BillingName,
		BillingPhone:   req.BillingPhone,
		BillingAddress: req.BillingAddress,
	})
	if err != nil {
		h.logger.Error("order create failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          result.Order.ID,
		"plan_slug":   result.Plan.Slug,
		"plan_name":   result.Plan.Name,
		"status":      result.Order.Status,
		"total":       result.Order.Total,
		"currency":    result.Order.Currency,
		"created_at":  result.Order.CreatedAt,
		"expires_at":  result.Order.ExpiresAt,
		"payment_url": result.PaymentURL,
	})
}

func (h *OrderHandler) List(c *gin.Context) {
	uid, _ := uuid.Parse(c.GetString(middleware.CtxUserID))
	orders, err := h.orders.ListByUser(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list failed"})
		return
	}
	out := make([]gin.H, 0, len(orders))
	for _, o := range orders {
		out = append(out, orderToDTO(o))
	}
	c.JSON(http.StatusOK, out)
}

func (h *OrderHandler) Detail(c *gin.Context) {
	uid, _ := uuid.Parse(c.GetString(middleware.CtxUserID))
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	order, plan, err := h.orders.GetOrder(c.Request.Context(), id, uid)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load failed"})
		return
	}
	dto := orderToDTO(order)
	if plan != nil {
		dto["plan_slug"] = plan.Slug
		dto["plan_name"] = plan.Name
	}
	c.JSON(http.StatusOK, dto)
}

func orderToDTO(o any) gin.H {
	// Tip dönüşümü (Go'da reflection olmadan struct field erişimi zor)
	// Order struct'ı burada import etmek yerine json round-trip yapıyoruz
	b, _ := json.Marshal(o)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return m
}

// Iyzico callback — kullanıcı 3D Secure sonrası dönüşünde
func (h *OrderHandler) IyzicoCallback(c *gin.Context) {
	orderIDStr := c.Query("order")
	if orderIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order required"})
		return
	}
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	// Iyzico token'ı ile retrieve çağrısı
	// (Gerçek flow: token checkout form'dan döner, biz de retrieve ile sonucu alırız)
	// Burada basitleştirilmiş: token query param'dan
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}

	// Burada normalde iyzico client ile RetrieveCheckoutForm çağrılır
	// ve başarılıysa order paid olarak işaretlenir
	_ = token

	// Demo/dev: token "dev-" prefixli ise direkt success say
	iyzicoID := ""
	if len(token) > 4 && token[:4] == "dev-" {
		iyzicoID = "dev-pay-" + strconv.FormatInt(int64(orderID[0]), 16)
		if err := h.orders.MarkPaid(c.Request.Context(), orderID, iyzicoID); err != nil {
			h.logger.Error("mark paid failed", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "mark paid failed"})
			return
		}
		// Web success sayfasına yönlendir
		c.Redirect(http.StatusFound, "/dashboard/orders/"+orderIDStr+"/success")
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// Asynq job enqueue helper (kullanılmıyor, scheduler tarafından tetiklenir)
// Burada service'e bağımlılığı azaltmak için
func enqueueTestStart(asynqClient *asynq.Client, testID, pkg string) error {
	if asynqClient == nil {
		return nil
	}
	payload, _ := json.Marshal(model.TestStartPayload{
		Payload: model.Payload{TestID: testID, PackageName: pkg},
	})
	task := asynq.NewTask(string(model.TaskTypeTestStart), payload,
		asynq.TaskID(model.JobID(testID, model.TaskTypeTestStart, 0)),
	)
	_, err := asynqClient.Enqueue(task, asynq.ProcessIn(2 * 1_000_000_000))
	return err
}
