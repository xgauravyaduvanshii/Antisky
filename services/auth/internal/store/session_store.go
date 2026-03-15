package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/antisky/services/auth/internal/models"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExpired  = errors.New("session expired")
)

type SessionStore struct {
	pool *pgxpool.Pool
	rdb  *redis.Client
}

func NewSessionStore(pool *pgxpool.Pool, rdb *redis.Client) *SessionStore {
	return &SessionStore{pool: pool, rdb: rdb}
}

// Create creates a new session and returns the refresh token
func (s *SessionStore) Create(ctx context.Context, userID uuid.UUID, ip, userAgent string, expiry time.Duration) (*models.Session, string, error) {
	// Generate refresh token
	refreshToken, err := generateSecureToken(32)
	if err != nil {
		return nil, "", err
	}

	tokenHash := hashToken(refreshToken)
	session := &models.Session{
		ID:               uuid.New(),
		UserID:           userID,
		RefreshTokenHash: tokenHash,
		IPAddress:        &ip,
		UserAgent:        &userAgent,
		ExpiresAt:        time.Now().Add(expiry),
		CreatedAt:        time.Now(),
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO sessions (id, user_id, refresh_token_hash, ip_address, user_agent, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		session.ID, session.UserID, session.RefreshTokenHash,
		session.IPAddress, session.UserAgent, session.ExpiresAt, session.CreatedAt,
	)
	if err != nil {
		return nil, "", err
	}

	// Cache session in Redis for fast lookups
	cacheKey := fmt.Sprintf("session:%s", session.ID)
	s.rdb.Set(ctx, cacheKey, session.UserID.String(), expiry)

	return session, refreshToken, nil
}

// ValidateRefreshToken validates a refresh token and returns the session
func (s *SessionStore) ValidateRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	tokenHash := hashToken(refreshToken)

	session := &models.Session{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, refresh_token_hash, ip_address, user_agent, expires_at, created_at
		 FROM sessions WHERE refresh_token_hash = $1`, tokenHash,
	).Scan(
		&session.ID, &session.UserID, &session.RefreshTokenHash,
		&session.IPAddress, &session.UserAgent, &session.ExpiresAt, &session.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		s.Delete(ctx, session.ID)
		return nil, ErrSessionExpired
	}

	return session, nil
}

// Delete removes a session
func (s *SessionStore) Delete(ctx context.Context, sessionID uuid.UUID) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, sessionID)
	if err != nil {
		return err
	}

	// Remove from Redis cache
	cacheKey := fmt.Sprintf("session:%s", sessionID)
	s.rdb.Del(ctx, cacheKey)

	return nil
}

// DeleteAllForUser removes all sessions for a user
func (s *SessionStore) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	// Get all session IDs first for Redis cleanup
	rows, err := s.pool.Query(ctx, `SELECT id FROM sessions WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var sessionIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return err
		}
		sessionIDs = append(sessionIDs, id)
	}

	// Delete from Postgres
	_, err = s.pool.Exec(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	if err != nil {
		return err
	}

	// Clean Redis cache
	for _, id := range sessionIDs {
		cacheKey := fmt.Sprintf("session:%s", id)
		s.rdb.Del(ctx, cacheKey)
	}

	return nil
}

// CleanExpired removes all expired sessions
func (s *SessionStore) CleanExpired(ctx context.Context) (int64, error) {
	result, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

func generateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
