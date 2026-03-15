package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/antisky/services/auth/internal/models"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailExists      = errors.New("email already exists")
	ErrInvalidPassword  = errors.New("invalid password")
)

type UserStore struct {
	pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) *UserStore {
	return &UserStore{pool: pool}
}

// Create registers a new user with hashed password
func (s *UserStore) Create(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hash),
		Provider:     "email",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO users (id, email, name, password_hash, provider, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.ID, user.Email, user.Name, user.PasswordHash, user.Provider, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		// Check for unique constraint violation
		if isDuplicateKeyError(err) {
			return nil, ErrEmailExists
		}
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (s *UserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, name, avatar_url, password_hash, mfa_enabled, email_verified,
		        provider, provider_id, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(
		&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.PasswordHash,
		&user.MFAEnabled, &user.EmailVerified, &user.Provider, &user.ProviderID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetByID retrieves a user by ID
func (s *UserStore) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, email, name, avatar_url, password_hash, mfa_enabled, email_verified,
		        provider, provider_id, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(
		&user.ID, &user.Email, &user.Name, &user.AvatarURL, &user.PasswordHash,
		&user.MFAEnabled, &user.EmailVerified, &user.Provider, &user.ProviderID,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// Update updates a user's profile
func (s *UserStore) Update(ctx context.Context, id uuid.UUID, req *models.UpdateProfileRequest) (*models.User, error) {
	if req.Name != nil {
		_, err := s.pool.Exec(ctx, `UPDATE users SET name = $1 WHERE id = $2`, *req.Name, id)
		if err != nil {
			return nil, err
		}
	}
	if req.AvatarURL != nil {
		_, err := s.pool.Exec(ctx, `UPDATE users SET avatar_url = $1 WHERE id = $2`, *req.AvatarURL, id)
		if err != nil {
			return nil, err
		}
	}
	return s.GetByID(ctx, id)
}

// VerifyPassword checks if the password matches the stored hash
func (s *UserStore) VerifyPassword(user *models.User, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidPassword
	}
	return nil
}

// CreateOAuth creates or updates a user from an OAuth provider
func (s *UserStore) CreateOAuth(ctx context.Context, email, name, provider, providerID, avatarURL string) (*models.User, error) {
	user := &models.User{
		ID:            uuid.New(),
		Email:         email,
		Name:          name,
		Provider:      provider,
		ProviderID:    &providerID,
		AvatarURL:     &avatarURL,
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err := s.pool.Exec(ctx,
		`INSERT INTO users (id, email, name, avatar_url, provider, provider_id, email_verified, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 ON CONFLICT (email) DO UPDATE SET
		   provider_id = EXCLUDED.provider_id,
		   avatar_url = EXCLUDED.avatar_url,
		   updated_at = NOW()`,
		user.ID, user.Email, user.Name, user.AvatarURL, user.Provider, user.ProviderID,
		user.EmailVerified, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Retrieve the actual user (might have been updated rather than inserted)
	return s.GetByEmail(ctx, email)
}

func isDuplicateKeyError(err error) bool {
	return err != nil && (errors.Is(err, pgx.ErrNoRows) == false) &&
		(err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" ||
			len(err.Error()) > 0 && contains(err.Error(), "23505"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
