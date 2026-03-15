package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/antisky/services/auth/internal/models"
)

var (
	ErrAPIKeyNotFound = errors.New("api key not found")
)

type APIKeyStore struct {
	pool *pgxpool.Pool
}

func NewAPIKeyStore(pool *pgxpool.Pool) *APIKeyStore {
	return &APIKeyStore{pool: pool}
}

// Create generates a new API key
func (s *APIKeyStore) Create(ctx context.Context, userID uuid.UUID, req *models.CreateAPIKeyRequest) (*models.APIKey, string, error) {
	// Generate a random API key: ask_ prefix + 40 random hex chars
	rawKey, err := generateAPIKey()
	if err != nil {
		return nil, "", err
	}

	fullKey := "ask_" + rawKey
	prefix := fullKey[:12]

	hash, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	apiKey := &models.APIKey{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      req.Name,
		KeyPrefix: prefix,
		KeyHash:   string(hash),
		Scopes:    req.Scopes,
		CreatedAt: time.Now(),
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO api_keys (id, user_id, name, key_prefix, key_hash, scopes, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		apiKey.ID, apiKey.UserID, apiKey.Name, apiKey.KeyPrefix, apiKey.KeyHash, apiKey.Scopes, apiKey.CreatedAt,
	)
	if err != nil {
		return nil, "", err
	}

	return apiKey, fullKey, nil
}

// ValidateKey checks an API key and returns the associated key record
func (s *APIKeyStore) ValidateKey(ctx context.Context, key string) (*models.APIKey, error) {
	if len(key) < 12 {
		return nil, ErrAPIKeyNotFound
	}

	prefix := key[:12]

	// Find candidates by prefix
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, name, key_prefix, key_hash, scopes, last_used_at, expires_at, created_at
		 FROM api_keys WHERE key_prefix = $1`, prefix,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		apiKey := &models.APIKey{}
		if err := rows.Scan(
			&apiKey.ID, &apiKey.UserID, &apiKey.Name, &apiKey.KeyPrefix,
			&apiKey.KeyHash, &apiKey.Scopes, &apiKey.LastUsedAt, &apiKey.ExpiresAt, &apiKey.CreatedAt,
		); err != nil {
			return nil, err
		}

		// Check if expired
		if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
			continue
		}

		// Verify key matches hash
		if err := bcrypt.CompareHashAndPassword([]byte(apiKey.KeyHash), []byte(key)); err == nil {
			// Update last_used_at
			s.pool.Exec(ctx, `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`, apiKey.ID)
			return apiKey, nil
		}
	}

	return nil, ErrAPIKeyNotFound
}

// ListByUser returns all API keys for a user
func (s *APIKeyStore) ListByUser(ctx context.Context, userID uuid.UUID) ([]*models.APIKey, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at
		 FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*models.APIKey
	for rows.Next() {
		key := &models.APIKey{}
		if err := rows.Scan(
			&key.ID, &key.UserID, &key.Name, &key.KeyPrefix,
			&key.Scopes, &key.LastUsedAt, &key.ExpiresAt, &key.CreatedAt,
		); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

// Delete removes an API key
func (s *APIKeyStore) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	result, err := s.pool.Exec(ctx,
		`DELETE FROM api_keys WHERE id = $1 AND user_id = $2`, id, userID,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrAPIKeyNotFound
	}
	return nil
}

// GetByID retrieves an API key by ID
func (s *APIKeyStore) GetByID(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
	apiKey := &models.APIKey{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, name, key_prefix, scopes, last_used_at, expires_at, created_at
		 FROM api_keys WHERE id = $1`, id,
	).Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.Name, &apiKey.KeyPrefix,
		&apiKey.Scopes, &apiKey.LastUsedAt, &apiKey.ExpiresAt, &apiKey.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAPIKeyNotFound
		}
		return nil, err
	}
	return apiKey, nil
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
