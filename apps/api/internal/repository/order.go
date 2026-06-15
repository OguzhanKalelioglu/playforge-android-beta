package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PlanTier struct {
	ID            uuid.UUID
	Slug          string
	Name          string
	Description   *string
	TesterCount   int
	DurationDays  int
	PriceTRY      float64
	PriceUSD      *float64
	Features      []byte
	SortOrder     int
	IsActive      bool
}

type PlanRepository struct{ db *pgxpool.Pool }

func NewPlanRepository(db *pgxpool.Pool) *PlanRepository { return &PlanRepository{db: db} }

func (r *PlanRepository) List(ctx context.Context) ([]*PlanTier, error) {
	const q = `
		SELECT id, slug, name, description, tester_count, duration_days,
		       price_try, price_usd, features, sort_order, is_active
		FROM plan_tiers WHERE is_active = TRUE ORDER BY sort_order ASC
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ps []*PlanTier
	for rows.Next() {
		p := &PlanTier{}
		err := rows.Scan(
			&p.ID, &p.Slug, &p.Name, &p.Description, &p.TesterCount, &p.DurationDays,
			&p.PriceTRY, &p.PriceUSD, &p.Features, &p.SortOrder, &p.IsActive,
		)
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}
	return ps, rows.Err()
}

func (r *PlanRepository) GetBySlug(ctx context.Context, slug string) (*PlanTier, error) {
	const q = `
		SELECT id, slug, name, description, tester_count, duration_days,
		       price_try, price_usd, features, sort_order, is_active
		FROM plan_tiers WHERE slug = $1 AND is_active = TRUE
	`
	p := &PlanTier{}
	err := r.db.QueryRow(ctx, q, slug).Scan(
		&p.ID, &p.Slug, &p.Name, &p.Description, &p.TesterCount, &p.DurationDays,
		&p.PriceTRY, &p.PriceUSD, &p.Features, &p.SortOrder, &p.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *PlanRepository) GetByID(ctx context.Context, id uuid.UUID) (*PlanTier, error) {
	const q = `
		SELECT id, slug, name, description, tester_count, duration_days,
		       price_try, price_usd, features, sort_order, is_active
		FROM plan_tiers WHERE id = $1
	`
	p := &PlanTier{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.Slug, &p.Name, &p.Description, &p.TesterCount, &p.DurationDays,
		&p.PriceTRY, &p.PriceUSD, &p.Features, &p.SortOrder, &p.IsActive,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

type Order struct {
	ID                    uuid.UUID
	UserID                uuid.UUID
	PlanTierID            uuid.UUID
	Status                string
	Subtotal              float64
	TaxTotal              float64
	Total                 float64
	Currency              string
	TestID                *uuid.UUID
	IyzicoCheckoutToken   *string
	IyzicoTokenExpiresAt  *time.Time
	BillingEmail          *string
	BillingName           *string
	BillingPhone          *string
	BillingAddress        *string
	Metadata              []byte
	ExpiresAt             time.Time
	PaidAt                *time.Time
	CancelledAt           *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type OrderRepository struct{ db *pgxpool.Pool }

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository { return &OrderRepository{db: db} }

func (r *OrderRepository) Create(ctx context.Context, o *Order) error {
	const q = `
		INSERT INTO orders (user_id, plan_tier_id, status, subtotal, tax_total, total,
		                    currency, billing_email, billing_name, billing_phone, billing_address,
		                    metadata, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, q,
		o.UserID, o.PlanTierID, o.Status, o.Subtotal, o.TaxTotal, o.Total, o.Currency,
		o.BillingEmail, o.BillingName, o.BillingPhone, o.BillingAddress, o.Metadata, o.ExpiresAt,
	).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
}

func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	const q = `
		SELECT id, user_id, plan_tier_id, status, subtotal, tax_total, total, currency,
		       test_id, iyzico_checkout_form_token, iyzico_token_expires_at,
		       billing_email, billing_name, billing_phone, billing_address,
		       metadata, expires_at, paid_at, cancelled_at, created_at, updated_at
		FROM orders WHERE id = $1
	`
	o := &Order{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&o.ID, &o.UserID, &o.PlanTierID, &o.Status, &o.Subtotal, &o.TaxTotal, &o.Total, &o.Currency,
		&o.TestID, &o.IyzicoCheckoutToken, &o.IyzicoTokenExpiresAt,
		&o.BillingEmail, &o.BillingName, &o.BillingPhone, &o.BillingAddress,
		&o.Metadata, &o.ExpiresAt, &o.PaidAt, &o.CancelledAt, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return o, nil
}

func (r *OrderRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*Order, error) {
	const q = `
		SELECT id, user_id, plan_tier_id, status, subtotal, tax_total, total, currency,
		       test_id, iyzico_checkout_form_token, iyzico_token_expires_at,
		       billing_email, billing_name, billing_phone, billing_address,
		       metadata, expires_at, paid_at, cancelled_at, created_at, updated_at
		FROM orders WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var os []*Order
	for rows.Next() {
		o := &Order{}
		err := rows.Scan(
			&o.ID, &o.UserID, &o.PlanTierID, &o.Status, &o.Subtotal, &o.TaxTotal, &o.Total, &o.Currency,
			&o.TestID, &o.IyzicoCheckoutToken, &o.IyzicoTokenExpiresAt,
			&o.BillingEmail, &o.BillingName, &o.BillingPhone, &o.BillingAddress,
			&o.Metadata, &o.ExpiresAt, &o.PaidAt, &o.CancelledAt, &o.CreatedAt, &o.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		os = append(os, o)
	}
	return os, rows.Err()
}

func (r *OrderRepository) ListAll(ctx context.Context, status string, limit int) ([]*Order, error) {
	q := `
		SELECT o.id, o.user_id, o.plan_tier_id, o.status, o.subtotal, o.tax_total, o.total, o.currency,
		       o.test_id, o.iyzico_checkout_form_token, o.iyzico_token_expires_at,
		       o.billing_email, o.billing_name, o.billing_phone, o.billing_address,
		       o.metadata, o.expires_at, o.paid_at, o.cancelled_at, o.created_at, o.updated_at,
		       u.email as user_email, pt.name as plan_name
		FROM orders o
		JOIN users u ON u.id = o.user_id
		JOIN plan_tiers pt ON pt.id = o.plan_tier_id
	`
	args := []any{}
	if status != "" {
		q += " WHERE o.status = $1"
		args = append(args, status)
		q += " ORDER BY o.created_at DESC LIMIT $2"
		args = append(args, limit)
	} else {
		q += " ORDER BY o.created_at DESC LIMIT $1"
		args = append(args, limit)
	}
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type row struct {
		Order
		UserEmail string
		PlanName  string
	}
	var rs []row
	for rows.Next() {
		var x row
		err := rows.Scan(
			&x.ID, &x.UserID, &x.PlanTierID, &x.Status, &x.Subtotal, &x.TaxTotal, &x.Total, &x.Currency,
			&x.TestID, &x.IyzicoCheckoutToken, &x.IyzicoTokenExpiresAt,
			&x.BillingEmail, &x.BillingName, &x.BillingPhone, &x.BillingAddress,
			&x.Metadata, &x.ExpiresAt, &x.PaidAt, &x.CancelledAt, &x.CreatedAt, &x.UpdatedAt,
			&x.UserEmail, &x.PlanName,
		)
		if err != nil {
			return nil, err
		}
		rs = append(rs, x)
	}
	return nil, rows.Err()
}

func (r *OrderRepository) MarkPaid(ctx context.Context, id uuid.UUID, testID uuid.UUID) error {
	const q = `
		UPDATE orders SET status = 'paid', paid_at = NOW(), test_id = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, q, id, testID)
	return err
}

func (r *OrderRepository) SetCheckoutToken(ctx context.Context, id uuid.UUID, token string, expires time.Time) error {
	const q = `
		UPDATE orders SET status = 'awaiting_payment',
		       iyzico_checkout_form_token = $2, iyzico_token_expires_at = $3,
		       updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, q, id, token, expires)
	return err
}
