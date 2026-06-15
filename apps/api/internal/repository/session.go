package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	UserAgent        *string
	IPAddress        *string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	LastUsedAt       time.Time
	CreatedAt        time.Time
}

type SessionRepository struct{ db *pgxpool.Pool }

func NewSessionRepository(db *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, s *Session) error {
	const q = `
		INSERT INTO sessions (user_id, refresh_token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, last_used_at, created_at
	`
	return r.db.QueryRow(ctx, q, s.UserID, s.RefreshTokenHash, s.UserAgent, s.IPAddress, s.ExpiresAt).
		Scan(&s.ID, &s.LastUsedAt, &s.CreatedAt)
}

func (r *SessionRepository) GetByTokenHash(ctx context.Context, hash string) (*Session, error) {
	const q = `
		SELECT id, user_id, refresh_token_hash, user_agent, ip_address, expires_at,
		       revoked_at, last_used_at, created_at
		FROM sessions WHERE refresh_token_hash = $1
	`
	s := &Session{}
	err := r.db.QueryRow(ctx, q, hash).Scan(
		&s.ID, &s.UserID, &s.RefreshTokenHash, &s.UserAgent, &s.IPAddress,
		&s.ExpiresAt, &s.RevokedAt, &s.LastUsedAt, &s.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *SessionRepository) Touch(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE sessions SET last_used_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}

func (r *SessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE sessions SET revoked_at = NOW() WHERE id = $1 AND revoked_at IS NULL`
	_, err := r.db.Exec(ctx, q, id)
	return err
}
