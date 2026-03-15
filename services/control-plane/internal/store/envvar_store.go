package store

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antisky/services/control-plane/internal/models"
)

type EnvVarStore struct {
	pool          *pgxpool.Pool
	encryptionKey []byte
}

func NewEnvVarStore(pool *pgxpool.Pool) *EnvVarStore {
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		key = "dev-encryption-key-32bytes!!" // 32 bytes
	}
	// Ensure exactly 32 bytes for AES-256
	keyBytes := []byte(key)
	if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}
	for len(keyBytes) < 32 {
		keyBytes = append(keyBytes, '0')
	}

	return &EnvVarStore{
		pool:          pool,
		encryptionKey: keyBytes,
	}
}

func (s *EnvVarStore) Set(ctx context.Context, projectID uuid.UUID, req *models.SetEnvVarRequest) (*models.EnvVar, error) {
	// Encrypt the value
	encrypted, err := s.encrypt(req.Value)
	if err != nil {
		return nil, err
	}

	target := req.Target
	if len(target) == 0 {
		target = []string{"production", "preview", "development"}
	}

	envVar := &models.EnvVar{
		ID:             uuid.New(),
		ProjectID:      projectID,
		Key:            req.Key,
		EncryptedValue: encrypted,
		Target:         target,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO project_env_vars (id, project_id, key, encrypted_value, target, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)
		 ON CONFLICT (project_id, key) DO UPDATE SET
		   encrypted_value = EXCLUDED.encrypted_value,
		   target = EXCLUDED.target`,
		envVar.ID, envVar.ProjectID, envVar.Key, envVar.EncryptedValue, envVar.Target, envVar.CreatedAt, envVar.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return envVar, nil
}

func (s *EnvVarStore) List(ctx context.Context, projectID uuid.UUID) ([]*models.EnvVar, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, project_id, key, encrypted_value, target, created_at, updated_at
		 FROM project_env_vars WHERE project_id = $1 ORDER BY key`, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var envVars []*models.EnvVar
	for rows.Next() {
		ev := &models.EnvVar{}
		if err := rows.Scan(&ev.ID, &ev.ProjectID, &ev.Key, &ev.EncryptedValue, &ev.Target, &ev.CreatedAt, &ev.UpdatedAt); err != nil {
			return nil, err
		}

		// Decrypt value for response
		decrypted, err := s.decrypt(ev.EncryptedValue)
		if err != nil {
			ev.Value = "***DECRYPTION_ERROR***"
		} else {
			ev.Value = decrypted
		}

		envVars = append(envVars, ev)
	}
	return envVars, nil
}

func (s *EnvVarStore) Delete(ctx context.Context, projectID uuid.UUID, key string) error {
	result, err := s.pool.Exec(ctx,
		`DELETE FROM project_env_vars WHERE project_id = $1 AND key = $2`, projectID, key,
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("environment variable not found")
	}
	return nil
}

func (s *EnvVarStore) GetDecryptedForProject(ctx context.Context, projectID uuid.UUID, target string) (map[string]string, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT key, encrypted_value FROM project_env_vars
		 WHERE project_id = $1 AND $2 = ANY(target)`, projectID, target,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	envMap := make(map[string]string)
	for rows.Next() {
		var key, encrypted string
		if err := rows.Scan(&key, &encrypted); err != nil {
			return nil, err
		}
		decrypted, err := s.decrypt(encrypted)
		if err != nil {
			continue
		}
		envMap[key] = decrypted
	}
	return envMap, nil
}

// AES-256-GCM encryption
func (s *EnvVarStore) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *EnvVarStore) decrypt(encoded string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
