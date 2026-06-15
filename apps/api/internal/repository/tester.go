package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Tester struct {
	ID              uuid.UUID
	Email           string
	PasswordEncrypted []byte
	RecoveryEmail   *string
	Phone           *string
	GoogleGroupID   *uuid.UUID
	DeviceProfileID *uuid.UUID
	Status          string
	Notes           *string
	LastUsedAt      *time.Time
	CreatedAt       time.Time
}

type TesterRepository struct{ db *pgxpool.Pool }

func NewTesterRepository(db *pgxpool.Pool) *TesterRepository {
	return &TesterRepository{db: db}
}

func (r *TesterRepository) ListAll(ctx context.Context) ([]*Tester, error) {
	const q = `
		SELECT id, email, password_encrypted, recovery_email, phone, google_group_id,
		       device_profile_id, status, notes, last_used_at, created_at
		FROM testers ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ts []*Tester
	for rows.Next() {
		t := &Tester{}
		err := rows.Scan(
			&t.ID, &t.Email, &t.PasswordEncrypted, &t.RecoveryEmail, &t.Phone, &t.GoogleGroupID,
			&t.DeviceProfileID, &t.Status, &t.Notes, &t.LastUsedAt, &t.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, rows.Err()
}

func (r *TesterRepository) GetByID(ctx context.Context, id uuid.UUID) (*Tester, error) {
	const q = `
		SELECT id, email, password_encrypted, recovery_email, phone, google_group_id,
		       device_profile_id, status, notes, last_used_at, created_at
		FROM testers WHERE id = $1
	`
	t := &Tester{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.Email, &t.PasswordEncrypted, &t.RecoveryEmail, &t.Phone, &t.GoogleGroupID,
		&t.DeviceProfileID, &t.Status, &t.Notes, &t.LastUsedAt, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

func (r *TesterRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	const q = `UPDATE testers SET status = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, q, id, status)
	return err
}

func (r *TesterRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE testers SET last_used_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}

type AdminTesterRow struct {
	Tester
	GroupEmail       *string
	DeviceModel      *string
	TasksCompleted30d int
}

func (r *TesterRepository) ListAdmin(ctx context.Context) ([]*AdminTesterRow, error) {
	const q = `
		SELECT t.id, t.email, t.password_encrypted, t.recovery_email, t.phone,
		       t.google_group_id, t.device_profile_id, t.status, t.notes, t.last_used_at, t.created_at,
		       gg.group_email, dp.model,
		       COALESCE((SELECT SUM(tasks_completed) FROM tester_daily_usage tdu WHERE tdu.tester_id = t.id AND tdu.usage_date > NOW() - INTERVAL '30 days'), 0)
		FROM testers t
		LEFT JOIN google_groups gg ON gg.id = t.google_group_id
		LEFT JOIN device_profiles dp ON dp.id = t.device_profile_id
		ORDER BY t.created_at ASC
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ts []*AdminTesterRow
	for rows.Next() {
		t := &AdminTesterRow{}
		err := rows.Scan(
			&t.ID, &t.Email, &t.PasswordEncrypted, &t.RecoveryEmail, &t.Phone,
			&t.GoogleGroupID, &t.DeviceProfileID, &t.Status, &t.Notes, &t.LastUsedAt, &t.CreatedAt,
			&t.GroupEmail, &t.DeviceModel, &t.TasksCompleted30d,
		)
		if err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, rows.Err()
}
