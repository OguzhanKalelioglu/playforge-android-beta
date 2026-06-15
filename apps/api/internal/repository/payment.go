package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Payment struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	TestID        *uuid.UUID
	Amount        float64
	Currency      string
	Status        string
	IyzicoToken   *string
	IyzicoPaymentID *string
	PaidAt        *time.Time
	RefundedAt    *time.Time
	RefundAmount  *float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PaymentRepository struct{ db *pgxpool.Pool }

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, p *Payment) error {
	const q = `
		INSERT INTO payments (user_id, test_id, amount, currency, status, iyzico_token, iyzico_payment_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, q,
		p.UserID, p.TestID, p.Amount, p.Currency, p.Status, p.IyzicoToken, p.IyzicoPaymentID,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PaymentRepository) GetByID(ctx context.Context, id uuid.UUID) (*Payment, error) {
	const q = `
		SELECT id, user_id, test_id, amount, currency, status, iyzico_token, iyzico_payment_id,
		       paid_at, refunded_at, refund_amount, created_at, updated_at
		FROM payments WHERE id = $1
	`
	p := &Payment{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.UserID, &p.TestID, &p.Amount, &p.Currency, &p.Status,
		&p.IyzicoToken, &p.IyzicoPaymentID, &p.PaidAt, &p.RefundedAt, &p.RefundAmount,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *PaymentRepository) MarkCompleted(ctx context.Context, id uuid.UUID, iyzicoID string) error {
	const q = `
		UPDATE payments SET status = 'completed', iyzico_payment_id = $2,
		       paid_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, q, id, iyzicoID)
	return err
}
