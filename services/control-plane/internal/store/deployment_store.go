package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antisky/services/control-plane/internal/models"
)

var (
	ErrDeploymentNotFound = errors.New("deployment not found")
)

type DeploymentStore struct {
	pool *pgxpool.Pool
}

func NewDeploymentStore(pool *pgxpool.Pool) *DeploymentStore {
	return &DeploymentStore{pool: pool}
}

func (s *DeploymentStore) Create(ctx context.Context, projectID uuid.UUID, triggeredBy *uuid.UUID, ref, commitSHA, commitMessage, commitAuthor, deployType, source string) (*models.Deployment, error) {
	meta := map[string]interface{}{"source": source}
	d := &models.Deployment{
		ID:            uuid.New(),
		ProjectID:     projectID,
		TriggeredBy:   triggeredBy,
		Ref:           &ref,
		CommitSHA:     &commitSHA,
		CommitMessage: &commitMessage,
		CommitAuthor:  &commitAuthor,
		Status:        "queued",
		Type:          deployType,
		Meta:          meta,
		CreatedAt:     time.Now(),
	}

	_, err := s.pool.Exec(ctx,
		`INSERT INTO deployments (id, project_id, triggered_by, ref, commit_sha, commit_message,
		     commit_author, status, type, meta, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		d.ID, d.ProjectID, d.TriggeredBy, d.Ref, d.CommitSHA, d.CommitMessage,
		d.CommitAuthor, d.Status, d.Type, d.Meta, d.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (s *DeploymentStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Deployment, error) {
	d := &models.Deployment{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, project_id, triggered_by, ref, commit_sha, commit_message, commit_author,
		        status, type, url, preview_url, build_log_url, build_duration_ms, meta,
		        error_message, started_at, completed_at, created_at
		 FROM deployments WHERE id = $1`, id,
	).Scan(
		&d.ID, &d.ProjectID, &d.TriggeredBy, &d.Ref, &d.CommitSHA, &d.CommitMessage, &d.CommitAuthor,
		&d.Status, &d.Type, &d.URL, &d.PreviewURL, &d.BuildLogURL, &d.BuildDurationMs, &d.Meta,
		&d.ErrorMessage, &d.StartedAt, &d.CompletedAt, &d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrDeploymentNotFound
		}
		return nil, err
	}
	return d, nil
}

func (s *DeploymentStore) ListByProject(ctx context.Context, projectID uuid.UUID, limit int) ([]*models.Deployment, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id, project_id, triggered_by, ref, commit_sha, commit_message, commit_author,
		        status, type, url, preview_url, build_duration_ms, error_message,
		        started_at, completed_at, created_at
		 FROM deployments WHERE project_id = $1
		 ORDER BY created_at DESC LIMIT $2`, projectID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deployments []*models.Deployment
	for rows.Next() {
		d := &models.Deployment{}
		if err := rows.Scan(
			&d.ID, &d.ProjectID, &d.TriggeredBy, &d.Ref, &d.CommitSHA, &d.CommitMessage, &d.CommitAuthor,
			&d.Status, &d.Type, &d.URL, &d.PreviewURL, &d.BuildDurationMs, &d.ErrorMessage,
			&d.StartedAt, &d.CompletedAt, &d.CreatedAt,
		); err != nil {
			return nil, err
		}
		deployments = append(deployments, d)
	}
	return deployments, nil
}

func (s *DeploymentStore) UpdateStatus(ctx context.Context, id uuid.UUID, status string, errorMsg *string) error {
	now := time.Now()
	switch status {
	case "building":
		_, err := s.pool.Exec(ctx,
			`UPDATE deployments SET status = $1, started_at = $2 WHERE id = $3`,
			status, now, id)
		return err
	case "ready", "failed", "cancelled":
		_, err := s.pool.Exec(ctx,
			`UPDATE deployments SET status = $1, completed_at = $2, error_message = $3,
			 build_duration_ms = EXTRACT(EPOCH FROM ($2 - started_at)) * 1000
			 WHERE id = $4`,
			status, now, errorMsg, id)
		return err
	default:
		_, err := s.pool.Exec(ctx,
			`UPDATE deployments SET status = $1 WHERE id = $2`, status, id)
		return err
	}
}

func (s *DeploymentStore) SetURL(ctx context.Context, id uuid.UUID, url, previewURL string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE deployments SET url = $1, preview_url = $2 WHERE id = $3`,
		url, previewURL, id)
	return err
}

func (s *DeploymentStore) GetLatestByProject(ctx context.Context, projectID uuid.UUID) (*models.Deployment, error) {
	d := &models.Deployment{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, project_id, ref, commit_sha, status, type, url, created_at
		 FROM deployments WHERE project_id = $1 AND status = 'ready'
		 ORDER BY created_at DESC LIMIT 1`, projectID,
	).Scan(&d.ID, &d.ProjectID, &d.Ref, &d.CommitSHA, &d.Status, &d.Type, &d.URL, &d.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return d, nil
}
