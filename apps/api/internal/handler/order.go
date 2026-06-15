package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/stripe/stripe-go/v82"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/middleware"
	"github.com/testerscommunity/api/internal/model"
	"github.com/testerscommunity/api/internal/service"
)

type OrderHandler struct {
	orders *service.OrderService
	stripe *service.StripeClient
	logger *zap.Logger
	asynq  *asynq.Client
}

func NewOrderHandler(orders *service.OrderService, stripe *service.StripeClient, asynq *asynq.Client, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{orders: orders, stripe: stripe, asynq: asynq, logger: logger}
}

func (h *OrderHandler) Register(r *gin.Engine) {
	api := r.Group("/api/v1")
	api.GET("/plans", h.ListPlans)

	// Stripe webhook — public, signature-based (no JWT, no CSRF)
	r.POST("/api/v1/payments/stripe/webhook", h.StripeWebhook)

	auth := api.Group("", middleware.AuthRequiredJWT())
	auth.POST("/orders", h.Create)
	auth.GET("/orders", h.List)
	auth.GET("/orders/:id", h.Detail)
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

// StripeWebhook — Stripe'tan gelen event'leri işler
//  - checkout.session.completed: order paid işaretlenir, test tetiklenir
//  - charge.refunded: payment refunded
//  - Her event 200 dönmek zorundadır; aksi halde Stripe tekrar dener
func (h *OrderHandler) StripeWebhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body"})
		return
	}
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing signature"})
		return
	}

	event, err := h.stripe.VerifyWebhook(payload, signature)
	if err != nil {
		h.logger.Warn("webhook signature verify failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			h.logger.Error("decode session", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": "decode"})
			return
		}
		orderIDStr := sess.Metadata["order_id"]
		orderID, err := uuid.Parse(orderIDStr)
		if err != nil {
			h.logger.Warn("webhook order_id parse failed", zap.String("raw", orderIDStr))
			c.JSON(http.StatusOK, gin.H{"ok": true})
			return
		}
		paymentIntent := ""
		if sess.PaymentIntent != nil {
			paymentIntent = sess.PaymentIntent.ID
		}
		if _, err := h.orders.MarkPaid(c.Request.Context(), orderID, sess.ID, paymentIntent); err != nil {
			h.logger.Error("mark paid via webhook failed",
				zap.Error(err),
				zap.String("order_id", orderIDStr),
				zap.String("stripe_session", sess.ID))
			// 500 dön → Stripe tekrar gönderir (idempotent olmalı)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "mark paid failed"})
			return
		}
		h.logger.Info("order paid via webhook",
			zap.String("order_id", orderIDStr),
			zap.String("stripe_session", sess.ID),
			zap.String("payment_intent", paymentIntent))

	case "charge.refunded":
		// İade akışı — şimdilik log
		h.logger.Info("stripe charge.refunded received", zap.String("event_id", event.ID))

	default:
		h.logger.Debug("unhandled stripe event", zap.String("type", string(event.Type)))
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
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
