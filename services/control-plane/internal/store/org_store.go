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
	ErrOrgNotFound = errors.New("organization not found")
)

type OrgStore struct {
	pool *pgxpool.Pool
}

func NewOrgStore(pool *pgxpool.Pool) *OrgStore {
	return &OrgStore{pool: pool}
}

func (s *OrgStore) Create(ctx context.Context, req *models.CreateOrgRequest, ownerID uuid.UUID) (*models.Organization, error) {
	org := &models.Organization{
		ID:        uuid.New(),
		Name:      req.Name,
		Slug:      req.Slug,
		Plan:      "free",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO organizations (id, name, slug, plan, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6)`,
		org.ID, org.Name, org.Slug, org.Plan, org.CreatedAt, org.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Add creator as owner
	_, err = tx.Exec(ctx,
		`INSERT INTO org_members (id, org_id, user_id, role, joined_at)
		 VALUES ($1,$2,$3,$4,$5)`,
		uuid.New(), org.ID, ownerID, "owner", time.Now(),
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return org, nil
}

func (s *OrgStore) GetByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	org := &models.Organization{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, slug, avatar_url, plan, billing_email, created_at, updated_at
		 FROM organizations WHERE id = $1`, id,
	).Scan(&org.ID, &org.Name, &org.Slug, &org.AvatarURL, &org.Plan, &org.BillingEmail, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrgNotFound
		}
		return nil, err
	}
	return org, nil
}

func (s *OrgStore) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.Organization, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT o.id, o.name, o.slug, o.avatar_url, o.plan, o.created_at, o.updated_at
		 FROM organizations o
		 JOIN org_members om ON o.id = om.org_id
		 WHERE om.user_id = $1
		 ORDER BY o.created_at`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orgs []*models.Organization
	for rows.Next() {
		org := &models.Organization{}
		if err := rows.Scan(&org.ID, &org.Name, &org.Slug, &org.AvatarURL, &org.Plan, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}

func (s *OrgStore) Update(ctx context.Context, id uuid.UUID, name string) (*models.Organization, error) {
	_, err := s.pool.Exec(ctx, `UPDATE organizations SET name = $1 WHERE id = $2`, name, id)
	if err != nil {
		return nil, err
	}
	return s.GetByID(ctx, id)
}

func (s *OrgStore) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := s.pool.Exec(ctx, `DELETE FROM organizations WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrOrgNotFound
	}
	return nil
}

func (s *OrgStore) ListMembers(ctx context.Context, orgID uuid.UUID) ([]*models.OrgMember, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT om.id, om.org_id, om.user_id, om.role, om.invited_by, om.joined_at,
		        u.email, u.name
		 FROM org_members om
		 JOIN users u ON om.user_id = u.id
		 WHERE om.org_id = $1
		 ORDER BY om.joined_at`, orgID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.OrgMember
	for rows.Next() {
		m := &models.OrgMember{}
		if err := rows.Scan(
			&m.ID, &m.OrgID, &m.UserID, &m.Role, &m.InvitedBy, &m.JoinedAt, &m.Email, &m.Name,
		); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, nil
}

func (s *OrgStore) AddMember(ctx context.Context, orgID, userID, invitedBy uuid.UUID, role string) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO org_members (id, org_id, user_id, role, invited_by, joined_at)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 ON CONFLICT (org_id, user_id) DO UPDATE SET role = EXCLUDED.role`,
		uuid.New(), orgID, userID, role, invitedBy, time.Now(),
	)
	return err
}

func (s *OrgStore) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM org_members WHERE org_id = $1 AND user_id = $2`, orgID, userID)
	return err
}

func (s *OrgStore) GetMemberRole(ctx context.Context, orgID, userID uuid.UUID) (string, error) {
	var role string
	err := s.pool.QueryRow(ctx,
		`SELECT role FROM org_members WHERE org_id = $1 AND user_id = $2`, orgID, userID,
	).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", errors.New("not a member of this organization")
		}
		return "", err
	}
	return role, nil
}
