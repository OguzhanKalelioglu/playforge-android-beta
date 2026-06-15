package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/webhook"
	"go.uber.org/zap"
)

// StripeClient, Stripe Checkout Session + webhook için minimum client
// Best practices:
//   - payment_method_types ASLA set edilmez (dynamic payment methods)
//   - Restricted API key (rk_) production'da tercih edilir
//   - Webhook signature verification her zaman yapılır
type StripeClient struct {
	secretKey      string
	webhookSecret  string
	publishableKey string
	apiVersion     string
	logger         *zap.Logger
}

const stripeAPIVersion = "2026-05-27.dahlia"

func NewStripeClient(logger *zap.Logger) *StripeClient {
	c := &StripeClient{
		secretKey:      os.Getenv("STRIPE_SECRET_KEY"),
		webhookSecret:  os.Getenv("STRIPE_WEBHOOK_SECRET"),
		publishableKey: os.Getenv("STRIPE_PUBLISHABLE_KEY"),
		apiVersion:     stripeAPIVersion,
		logger:         logger,
	}
	// Stripe SDK v82 hala global stripe.Key kullanıyor.
	// Production'da init sırasında set edilir; test/dev'de boş bırakılır.
	if c.secretKey != "" {
		stripe.Key = c.secretKey
	}
	return c
}

func (c *StripeClient) IsConfigured() bool {
	return c.secretKey != ""
}

func (c *StripeClient) WebhookSecret() string { return c.webhookSecret }

func (c *StripeClient) PublishableKey() string { return c.publishableKey }

// CreateCheckoutSessionInput, sipariş için Stripe Checkout Session oluşturur
type CreateCheckoutSessionInput struct {
	OrderID        string  // metadata.order_id
	UserID         string  // metadata.user_id
	PackageName    string  // line item name
	PlanName       string  // description
	AmountTRY      float64 // kuruş cinsinden değil, TL olarak (price_data currency=try)
	SuccessURL     string
	CancelURL      string
	BillingEmail   string
	CustomerName   string
	Lang           string // "tr" default
}

type CreateCheckoutSessionResult struct {
	SessionID  string
	URL        string
	ExpiresAt  time.Time
	CustomerID string
	// Dev stub alanları (Stripe olmadan test için)
	DevStub bool
}

// CreateCheckoutSession, hosted Checkout Session oluşturur
// payment_method_types omitted → dynamic payment methods
func (c *StripeClient) CreateCheckoutSession(ctx context.Context, in CreateCheckoutSessionInput) (*CreateCheckoutSessionResult, error) {
	if !c.IsConfigured() {
		// Dev stub: gerçek API çağrısı yapma, sahte session döndür
		// UI tarafı success URL'e direkt yönlenir (test amaçlı)
		return &CreateCheckoutSessionResult{
			SessionID: fmt.Sprintf("cs_dev_%d", time.Now().UnixNano()),
			URL:       in.SuccessURL + "?dev_stub=1",
			ExpiresAt: time.Now().Add(30 * time.Minute),
			DevStub:   true,
		}, nil
	}

	params := &stripe.CheckoutSessionParams{
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(in.SuccessURL),
		CancelURL:  stripe.String(in.CancelURL),
		// payment_method_types YOK — dynamic payment methods
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String("try"),
					UnitAmount: stripe.Int64(int64(in.AmountTRY * 100)), // TL → kuruş
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:        stripe.String(in.PlanName + " — " + in.PackageName),
						Description: stripe.String("14 günlük 25 hesap kapalı beta testi"),
					},
				},
			},
		},
		Metadata: map[string]string{
			"order_id":     in.OrderID,
			"user_id":      in.UserID,
			"package_name": in.PackageName,
		},
		Locale: stripe.String(langOrDefault(in.Lang)),
		// 30 dakika expire
		ExpiresAt: stripe.Int64(time.Now().Add(30 * time.Minute).Unix()),
	}

	if in.BillingEmail != "" {
		params.CustomerEmail = stripe.String(in.BillingEmail)
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe checkout session: %w", err)
	}

	return &CreateCheckoutSessionResult{
		SessionID:  sess.ID,
		URL:        sess.URL,
		ExpiresAt:  time.Unix(sess.ExpiresAt, 0),
		CustomerID: "", // sess.Customer.ID yok (guest checkout)
		DevStub:    false,
	}, nil
}

// VerifyWebhook, Stripe'dan gelen webhook body'sini imza ile doğrular
// VerifyWebhook, Stripe'dan gelen webhook body'sini imza ile doğrular
// Webhook secret mutlaka set edilmeli (production'da whsec_...)
func (c *StripeClient) VerifyWebhook(payload []byte, signature string) (stripe.Event, error) {
	if c.webhookSecret == "" {
		return stripe.Event{}, errors.New("STRIPE_WEBHOOK_SECRET not configured")
	}
	event, err := webhook.ConstructEvent(payload, signature, c.webhookSecret)
	if err != nil {
		return stripe.Event{}, fmt.Errorf("verify webhook: %w", err)
	}
	return event, nil
}

func langOrDefault(lang string) string {
	switch lang {
	case "tr", "en", "de", "fr":
		return lang
	}
	return "tr"
}
