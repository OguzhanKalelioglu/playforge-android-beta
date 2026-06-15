package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Test struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	PackageName   string
	TestLink      *string
	Notes         *string
	StarPreference string
	Status        string
	GoogleGroupID *uuid.UUID
	StartedAt     *time.Time
	EndsAt        *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type TestRepository struct {
	db *pgxpool.Pool
}

func NewTestRepository(db *pgxpool.Pool) *TestRepository {
	return &TestRepository{db: db}
}

func (r *TestRepository) Create(ctx context.Context, t *Test) error {
	const q = `
		INSERT INTO tests (user_id, package_name, test_link, notes, star_preference, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, q, t.UserID, t.PackageName, t.TestLink, t.Notes, t.StarPreference, t.Status).
		Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *TestRepository) GetByID(ctx context.Context, id uuid.UUID) (*Test, error) {
	const q = `
		SELECT id, user_id, package_name, test_link, notes, star_preference, status,
		       google_group_id, started_at, ends_at, created_at, updated_at
		FROM tests WHERE id = $1
	`
	t := &Test{}
	err := r.db.QueryRow(ctx, q, id).Scan(
		&t.ID, &t.UserID, &t.PackageName, &t.TestLink, &t.Notes, &t.StarPreference, &t.Status,
		&t.GoogleGroupID, &t.StartedAt, &t.EndsAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

func (r *TestRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*Test, error) {
	const q = `
		SELECT id, user_id, package_name, test_link, notes, star_preference, status,
		       google_group_id, started_at, ends_at, created_at, updated_at
		FROM tests WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []*Test
	for rows.Next() {
		t := &Test{}
		err := rows.Scan(
			&t.ID, &t.UserID, &t.PackageName, &t.TestLink, &t.Notes, &t.StarPreference, &t.Status,
			&t.GoogleGroupID, &t.StartedAt, &t.EndsAt, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tests = append(tests, t)
	}
	return tests, rows.Err()
}

func (r *TestRepository) ListAll(ctx context.Context, status string, limit int) ([]*Test, error) {
	q := `
		SELECT id, user_id, package_name, test_link, notes, star_preference, status,
		       google_group_id, started_at, ends_at, created_at, updated_at
		FROM tests
	`
	args := []any{}
	if status != "" {
		q += " WHERE status = $1"
		args = append(args, status)
		q += " ORDER BY created_at DESC LIMIT $2"
		args = append(args, limit)
	} else {
		q += " ORDER BY created_at DESC LIMIT $1"
		args = append(args, limit)
	}
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tests []*Test
	for rows.Next() {
		t := &Test{}
		err := rows.Scan(
			&t.ID, &t.UserID, &t.PackageName, &t.TestLink, &t.Notes, &t.StarPreference, &t.Status,
			&t.GoogleGroupID, &t.StartedAt, &t.EndsAt, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tests = append(tests, t)
	}
	return tests, rows.Err()
}

type Assignment struct {
	ID               uuid.UUID
	TestID           uuid.UUID
	TesterID         uuid.UUID
	Status           string
	OptInAt          *time.Time
	InstallAt        *time.Time
	LastEngagementAt *time.Time
	ReviewID         *uuid.UUID
	ErrorMessage     *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (r *TestRepository) ListAssignments(ctx context.Context, testID uuid.UUID) ([]*Assignment, error) {
	const q = `
		SELECT id, test_id, tester_id, status, opt_in_at, install_at,
		       last_engagement_at, review_id, error_message, created_at, updated_at
		FROM test_assignments WHERE test_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, q, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var as []*Assignment
	for rows.Next() {
		a := &Assignment{}
		err := rows.Scan(
			&a.ID, &a.TestID, &a.TesterID, &a.Status, &a.OptInAt, &a.InstallAt,
			&a.LastEngagementAt, &a.ReviewID, &a.ErrorMessage, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	return as, rows.Err()
}

type ActivityRow struct {
	ID                  int64
	TestAssignmentID    uuid.UUID
	Action              string
	PerformedAt         time.Time
	Success             bool
	ErrorMessage        *string
	ScreenshotPath      *string
	Metadata            []byte
}

func (r *TestRepository) ListActivity(ctx context.Context, testID uuid.UUID, limit int) ([]*ActivityRow, error) {
	const q = `
		SELECT al.id, al.test_assignment_id, al.action, al.performed_at, al.success,
		       al.error_message, al.screenshot_path, al.metadata
		FROM activity_logs al
		JOIN test_assignments ta ON ta.id = al.test_assignment_id
		WHERE ta.test_id = $1
		ORDER BY al.performed_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, q, testID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var as []*ActivityRow
	for rows.Next() {
		a := &ActivityRow{}
		err := rows.Scan(
			&a.ID, &a.TestAssignmentID, &a.Action, &a.PerformedAt, &a.Success,
			&a.ErrorMessage, &a.ScreenshotPath, &a.Metadata,
		)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	return as, rows.Err()
}

type Review struct {
	ID                 uuid.UUID
	TestAssignmentID   uuid.UUID
	Rating             int
	ReviewText         string
	Language           string
	PostedAt           time.Time
}

func (r *TestRepository) ListReviews(ctx context.Context, testID uuid.UUID) ([]*Review, error) {
	const q = `
		SELECT r.id, r.test_assignment_id, r.rating, r.review_text, r.language, r.posted_at
		FROM reviews r
		JOIN test_assignments ta ON ta.id = r.test_assignment_id
		WHERE ta.test_id = $1
		ORDER BY r.posted_at DESC
	`
	rows, err := r.db.Query(ctx, q, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rs []*Review
	for rows.Next() {
		rv := &Review{}
		err := rows.Scan(&rv.ID, &rv.TestAssignmentID, &rv.Rating, &rv.ReviewText, &rv.Language, &rv.PostedAt)
		if err != nil {
			return nil, err
		}
		rs = append(rs, rv)
	}
	return rs, rows.Err()
}

type ProgressSummary struct {
	Total    int
	OptIn    int
	Installed int
	Engaged  int
	Reviewed int
	Failed   int
}

func (r *TestRepository) Progress(ctx context.Context, testID uuid.UUID) (*ProgressSummary, error) {
	const q = `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status IN ('in_progress','completed','skipped') OR opt_in_at IS NOT NULL),
			COUNT(*) FILTER (WHERE install_at IS NOT NULL),
			COUNT(*) FILTER (WHERE last_engagement_at IS NOT NULL),
			COUNT(*) FILTER (WHERE status = 'completed' AND review_id IS NOT NULL),
			COUNT(*) FILTER (WHERE status = 'failed')
		FROM test_assignments WHERE test_id = $1
	`
	p := &ProgressSummary{}
	err := r.db.QueryRow(ctx, q, testID).Scan(
		&p.Total, &p.OptIn, &p.Installed, &p.Engaged, &p.Reviewed, &p.Failed,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}
