package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antisky/services/control-plane/internal/models"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrProjectSlugExists = errors.New("project slug already exists in this org")
)

type ProjectStore struct {
	pool *pgxpool.Pool
}

func NewProjectStore(pool *pgxpool.Pool) *ProjectStore {
	return &ProjectStore{pool: pool}
}

func (s *ProjectStore) Create(ctx context.Context, req *models.CreateProjectRequest, userID uuid.UUID) (*models.Project, error) {
	slug := generateSlug(req.Name)
	if req.RepoBranch == "" {
		req.RepoBranch = "main"
	}
	if req.RootDir == "" {
		req.RootDir = "/"
	}
	if req.Runtime == "" {
		req.Runtime = "nodejs"
	}

	project := &models.Project{
		ID:           uuid.New(),
		OrgID:        req.OrgID,
		Name:         req.Name,
		Slug:         slug,
		RepoURL:      req.RepoURL,
		RepoProvider: req.RepoProvider,
		Framework:    req.Framework,
		Runtime:      req.Runtime,
		BuildCommand: req.BuildCommand,
		StartCommand: req.StartCommand,
		OutputDir:    req.OutputDir,
		RootDir:      req.RootDir,
		RepoBranch:   req.RepoBranch,
		AutoDeploy:   true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := s.pool.Exec(ctx,
		`INSERT INTO projects (id, org_id, name, slug, repo_url, repo_provider, framework, runtime,
		     build_command, start_command, output_dir, root_dir, repo_branch, auto_deploy, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		project.ID, project.OrgID, project.Name, project.Slug, project.RepoURL, project.RepoProvider,
		project.Framework, project.Runtime, project.BuildCommand, project.StartCommand,
		project.OutputDir, project.RootDir, project.RepoBranch, project.AutoDeploy,
		project.CreatedAt, project.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrProjectSlugExists
		}
		return nil, err
	}

	return project, nil
}

func (s *ProjectStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	p := &models.Project{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, org_id, name, slug, description, repo_url, repo_provider, repo_id,
		        repo_branch, framework, runtime, build_command, start_command, output_dir,
		        root_dir, install_command, node_version, auto_deploy, config, created_at, updated_at
		 FROM projects WHERE id = $1`, id,
	).Scan(
		&p.ID, &p.OrgID, &p.Name, &p.Slug, &p.Description, &p.RepoURL, &p.RepoProvider, &p.RepoID,
		&p.RepoBranch, &p.Framework, &p.Runtime, &p.BuildCommand, &p.StartCommand, &p.OutputDir,
		&p.RootDir, &p.InstallCommand, &p.NodeVersion, &p.AutoDeploy, &p.Config, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *ProjectStore) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*models.Project, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, org_id, name, slug, description, repo_url, repo_provider,
		        framework, runtime, auto_deploy, created_at, updated_at
		 FROM projects WHERE org_id = $1 ORDER BY created_at DESC`, orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		p := &models.Project{}
		if err := rows.Scan(
			&p.ID, &p.OrgID, &p.Name, &p.Slug, &p.Description, &p.RepoURL, &p.RepoProvider,
			&p.Framework, &p.Runtime, &p.AutoDeploy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *ProjectStore) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Project, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT p.id, p.org_id, p.name, p.slug, p.description, p.repo_url, p.repo_provider,
		        p.framework, p.runtime, p.auto_deploy, p.created_at, p.updated_at
		 FROM projects p
		 JOIN org_members om ON p.org_id = om.org_id
		 WHERE om.user_id = $1
		 ORDER BY p.updated_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		p := &models.Project{}
		if err := rows.Scan(
			&p.ID, &p.OrgID, &p.Name, &p.Slug, &p.Description, &p.RepoURL, &p.RepoProvider,
			&p.Framework, &p.Runtime, &p.AutoDeploy, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, nil
}

func (s *ProjectStore) Update(ctx context.Context, id uuid.UUID, req *models.UpdateProjectRequest) (*models.Project, error) {
	// Build dynamic update
	sets := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Name != nil {
		sets = append(sets, "name = $"+itoa(argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Description != nil {
		sets = append(sets, "description = $"+itoa(argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.BuildCommand != nil {
		sets = append(sets, "build_command = $"+itoa(argIdx))
		args = append(args, *req.BuildCommand)
		argIdx++
	}
	if req.StartCommand != nil {
		sets = append(sets, "start_command = $"+itoa(argIdx))
		args = append(args, *req.StartCommand)
		argIdx++
	}
	if req.OutputDir != nil {
		sets = append(sets, "output_dir = $"+itoa(argIdx))
		args = append(args, *req.OutputDir)
		argIdx++
	}
	if req.Framework != nil {
		sets = append(sets, "framework = $"+itoa(argIdx))
		args = append(args, *req.Framework)
		argIdx++
	}
	if req.Runtime != nil {
		sets = append(sets, "runtime = $"+itoa(argIdx))
		args = append(args, *req.Runtime)
		argIdx++
	}
	if req.AutoDeploy != nil {
		sets = append(sets, "auto_deploy = $"+itoa(argIdx))
		args = append(args, *req.AutoDeploy)
		argIdx++
	}
	if req.RepoBranch != nil {
		sets = append(sets, "repo_branch = $"+itoa(argIdx))
		args = append(args, *req.RepoBranch)
		argIdx++
	}

	if len(sets) == 0 {
		return s.GetByID(ctx, id)
	}

	args = append(args, id)
	query := "UPDATE projects SET " + strings.Join(sets, ", ") + " WHERE id = $" + itoa(argIdx)
	_, err := s.pool.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return s.GetByID(ctx, id)
}

func (s *ProjectStore) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func (s *ProjectStore) GetByRepoID(ctx context.Context, provider, repoID string) (*models.Project, error) {
	p := &models.Project{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, org_id, name, slug, repo_url, repo_provider, repo_id, repo_branch,
		        framework, runtime, build_command, start_command, output_dir, root_dir,
		        auto_deploy, created_at, updated_at
		 FROM projects WHERE repo_provider = $1 AND repo_id = $2`, provider, repoID,
	).Scan(
		&p.ID, &p.OrgID, &p.Name, &p.Slug, &p.RepoURL, &p.RepoProvider, &p.RepoID,
		&p.RepoBranch, &p.Framework, &p.Runtime, &p.BuildCommand, &p.StartCommand,
		&p.OutputDir, &p.RootDir, &p.AutoDeploy, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProjectNotFound
		}
		return nil, err
	}
	return p, nil
}

// --- Domain Store Methods ---

func (s *ProjectStore) ListDomains(ctx context.Context, projectID uuid.UUID) ([]*models.ProjectDomain, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, project_id, domain, is_primary, verified, ssl_status, created_at, updated_at
		 FROM project_domains WHERE project_id = $1 ORDER BY created_at`, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []*models.ProjectDomain
	for rows.Next() {
		d := &models.ProjectDomain{}
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Domain, &d.IsPrimary, &d.Verified,
			&d.SSLStatus, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, nil
}

func (s *ProjectStore) AddDomain(ctx context.Context, projectID uuid.UUID, domain string) (*models.ProjectDomain, error) {
	token := uuid.New().String()[:16]
	d := &models.ProjectDomain{
		ID:                uuid.New(),
		ProjectID:         projectID,
		Domain:            domain,
		VerificationToken: &token,
		SSLStatus:         "pending",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	_, err := s.pool.Exec(ctx,
		`INSERT INTO project_domains (id, project_id, domain, verification_token, ssl_status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		d.ID, d.ProjectID, d.Domain, d.VerificationToken, d.SSLStatus, d.CreatedAt, d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (s *ProjectStore) RemoveDomain(ctx context.Context, domainID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM project_domains WHERE id = $1`, domainID)
	return err
}

// Helpers

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove non-alphanumeric (except hyphens)
	result := make([]byte, 0, len(slug))
	for _, c := range []byte(slug) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result = append(result, c)
		}
	}
	return string(result)
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}

func itoa(i int) string {
	const digits = "0123456789"
	if i < 10 {
		return string(digits[i])
	}
	return itoa(i/10) + string(digits[i%10])
}
