package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/testerscommunity/api/internal/model"
	"github.com/testerscommunity/api/internal/repository"
)

var (
	ErrOrderNotFound   = errors.New("order not found")
	ErrOrderExpired    = errors.New("order expired")
	ErrOrderNotPending = errors.New("order not in pending state")
	ErrPlanNotFound    = errors.New("plan not found")
)

// OrderService, sipariş + ödeme + test tetikleme akışını yönetir
// Ödeme: Stripe Checkout Sessions
type OrderService struct {
	orders      *repository.OrderRepository
	plans       *repository.PlanRepository
	tests       *repository.TestRepository
	payments    *repository.PaymentRepository
	stripe      *StripeClient
	asynqClient *asynq.Client
	logger      *zap.Logger
}

func NewOrderService(
	orders *repository.OrderRepository,
	plans *repository.PlanRepository,
	tests *repository.TestRepository,
	payments *repository.PaymentRepository,
	stripe *StripeClient,
	asynqClient *asynq.Client,
	logger *zap.Logger,
) *OrderService {
	return &OrderService{
		orders:      orders,
		plans:       plans,
		tests:       tests,
		payments:    payments,
		stripe:      stripe,
		asynqClient: asynqClient,
		logger:      logger,
	}
}

type CreateOrderInput struct {
	UserID         uuid.UUID
	PlanSlug       string
	PackageName    string
	TestLink       string
	BillingEmail   string
	BillingName    string
	BillingPhone   string
	BillingAddress string
}

type CreateOrderResult struct {
	Order      *repository.Order
	Plan       *repository.PlanTier
	PaymentURL string
	SessionID  string
	ExpiresAt  time.Time
	DevStub    bool
}

func (s *OrderService) Create(ctx context.Context, in CreateOrderInput) (*CreateOrderResult, error) {
	plan, err := s.plans.GetBySlug(ctx, in.PlanSlug)
	if err != nil {
		return nil, ErrPlanNotFound
	}

	order := &repository.Order{
		UserID:     in.UserID,
		PlanTierID: plan.ID,
		Status:     "pending",
		Subtotal:   plan.PriceTRY,
		TaxTotal:   0,
		Total:      plan.PriceTRY,
		Currency:   "TRY",
		BillingEmail:   &in.BillingEmail,
		BillingName:    &in.BillingName,
		BillingPhone:   &in.BillingPhone,
		BillingAddress: &in.BillingAddress,
		Metadata:       []byte(`{}`),
		ExpiresAt:      time.Now().Add(30 * time.Minute),
	}
	if err := s.orders.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	// Stripe Checkout Session
	webBase := os.Getenv("PUBLIC_WEB_URL")
	if webBase == "" {
		webBase = "http://localhost:3000"
	}
	successURL := fmt.Sprintf("%s/dashboard/orders/%s/success?session_id={CHECKOUT_SESSION_ID}", webBase, order.ID.String())
	cancelURL := fmt.Sprintf("%s/dashboard/orders/%s/pay?cancelled=1", webBase, order.ID.String())

	checkout, err := s.stripe.CreateCheckoutSession(ctx, CreateCheckoutSessionInput{
		OrderID:      order.ID.String(),
		UserID:       in.UserID.String(),
		PackageName:  in.PackageName,
		PlanName:     plan.Name,
		AmountTRY:    plan.PriceTRY,
		SuccessURL:   successURL,
		CancelURL:    cancelURL,
		BillingEmail: in.BillingEmail,
		CustomerName: in.BillingName,
		Lang:         "tr",
	})
	if err != nil {
		s.logger.Error("stripe checkout session failed",
			zap.Error(err),
			zap.String("order_id", order.ID.String()))
		return nil, fmt.Errorf("stripe checkout: %w", err)
	}

	if err := s.orders.SetStripeSession(ctx, order.ID, checkout.SessionID, checkout.ExpiresAt); err != nil {
		return nil, fmt.Errorf("set stripe session: %w", err)
	}

	return &CreateOrderResult{
		Order:      order,
		Plan:       plan,
		PaymentURL: checkout.URL,
		SessionID:  checkout.SessionID,
		ExpiresAt:  checkout.ExpiresAt,
		DevStub:    checkout.DevStub,
	}, nil
}

// MarkPaid, Stripe webhook checkout.session.completed event'iyle çağrılır
// - Order paid işaretlenir
// - Test oluşturulur (plan bazlı)
// - Payment kaydı
// - Asynq test_start job hemen tetiklenir
func (s *OrderService) MarkPaid(ctx context.Context, orderID uuid.UUID, stripeSessionID, stripePaymentIntentID string) (*repository.Test, error) {
	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, ErrOrderNotFound
	}
	if order.Status == "paid" {
		// Idempotent: zaten paid, test'i getir ve dön
		if order.TestID != nil {
			t, terr := s.tests.GetByID(ctx, *order.TestID)
			if terr == nil {
				return t, nil
			}
		}
		return nil, nil
	}
	if order.Status != "awaiting_payment" && order.Status != "pending" {
		return nil, ErrOrderNotPending
	}
	if time.Now().After(order.ExpiresAt) {
		return nil, ErrOrderExpired
	}

	// Test oluştur (paket adı metadata'dan gelmeli — sipariş sırasında kaydedilmedi)
	// Test paket adı order'ın metadata'sında veya ek bir alanda olmalı
	// Şimdilik billing_address içinde "pkg:" prefix ile encoded
	pkgName := extractPackageName(order.Metadata, order.BillingAddress)
	if pkgName == "" {
		pkgName = "unknown.package"
	}

	testLink := extractTestLink(order.Metadata)

	plan, _ := s.plans.GetByID(ctx, order.PlanTierID)

	t := &repository.Test{
		UserID:         order.UserID,
		PackageName:    pkgName,
		TestLink:       &testLink,
		Status:         "active",
		StarPreference: "mixed",
	}
	if err := s.tests.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("create test: %w", err)
	}

	// Order'ı paid işaretle + test_id bağla
	if err := s.orders.MarkPaid(ctx, orderID, t.ID); err != nil {
		return nil, fmt.Errorf("mark paid: %w", err)
	}

	// Payment kaydı
	p := &repository.Payment{
		UserID:             order.UserID,
		TestID:             &t.ID,
		Amount:             order.Total,
		Currency:           order.Currency,
		Status:             "completed",
		StripeSessionID:    &stripeSessionID,
		StripePaymentID:    &stripePaymentIntentID,
		StripeChargeID:     nil,
	}
	if err := s.payments.Create(ctx, p); err != nil {
		s.logger.Error("payment create failed", zap.Error(err))
	}
	if stripePaymentIntentID != "" {
		_ = s.payments.MarkCompleted(ctx, p.ID, stripePaymentIntentID)
	}

	// Asynq test_start job
	if s.asynqClient != nil {
		payload, _ := json.Marshal(model.TestStartPayload{
			Payload: model.Payload{
				TestID:      t.ID.String(),
				PackageName: pkgName,
			},
		})
		task := asynq.NewTask(string(model.TaskTypeTestStart), payload,
			asynq.TaskID(model.JobID(t.ID.String(), model.TaskTypeTestStart, 0)),
		)
		if _, err := s.asynqClient.Enqueue(task, asynq.ProcessIn(2*time.Minute)); err != nil {
			s.logger.Error("asynq enqueue failed", zap.Error(err))
		}
	}

	s.logger.Info("order paid",
		zap.String("order_id", orderID.String()),
		zap.String("test_id", t.ID.String()),
		zap.String("stripe_session", stripeSessionID),
		zap.Float64("amount", order.Total),
		zap.String("plan", strDeref(plan, func(p *repository.PlanTier) string { return p.Slug }, "")))

	return t, nil
}

func strDeref[T any](p *T, fn func(*T) string, def string) string {
	if p == nil {
		return def
	}
	return fn(p)
}

// extractPackageName, metadata JSON içinden package_name alır
func extractPackageName(metadata []byte, fallback *string) string {
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err == nil {
			if v, ok := m["package_name"].(string); ok && v != "" {
				return v
			}
		}
	}
	if fallback != nil && *fallback != "" {
		// billing_address "pkg:com.x.y|name:test" formatında olabilir
		return *fallback
	}
	return ""
}

func extractTestLink(metadata []byte) string {
	if len(metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(metadata, &m); err == nil {
			if v, ok := m["test_link"].(string); ok {
				return v
			}
		}
	}
	return ""
}

// GetOrder, sahiplik kontrolü ile
func (s *OrderService) GetOrder(ctx context.Context, orderID, userID uuid.UUID) (*repository.Order, *repository.PlanTier, error) {
	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, nil, ErrOrderNotFound
	}
	if order.UserID != userID {
		return nil, nil, ErrOrderNotFound
	}
	plan, err := s.plans.GetByID(ctx, order.PlanTierID)
	if err != nil {
		return order, nil, nil
	}
	return order, plan, nil
}

func (s *OrderService) ListByUser(ctx context.Context, userID uuid.UUID) ([]*repository.Order, error) {
	return s.orders.ListByUser(ctx, userID)
}

func (s *OrderService) GetPlans(ctx context.Context) ([]*repository.PlanTier, error) {
	return s.plans.List(ctx)
}
