package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/antisky/services/auth/internal/middleware"
	"github.com/antisky/services/auth/internal/models"
	"github.com/antisky/services/auth/internal/store"
)

type AuthHandler struct {
	userStore    *store.UserStore
	sessionStore *store.SessionStore
	apiKeyStore  *store.APIKeyStore
	jwtManager   *middleware.JWTManager
}

func NewAuthHandler(
	userStore *store.UserStore,
	sessionStore *store.SessionStore,
	apiKeyStore *store.APIKeyStore,
	jwtManager *middleware.JWTManager,
) *AuthHandler {
	return &AuthHandler{
		userStore:    userStore,
		sessionStore: sessionStore,
		apiKeyStore:  apiKeyStore,
		jwtManager:   jwtManager,
	}
}

// Register creates a new user account
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	// Validate input
	if req.Email == "" || req.Name == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Email, name, and password are required", "MISSING_FIELDS")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "Password must be at least 8 characters", "WEAK_PASSWORD")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Create user
	user, err := h.userStore.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, store.ErrEmailExists) {
			writeError(w, http.StatusConflict, "An account with this email already exists", "EMAIL_EXISTS")
			return
		}
		log.Printf("Error creating user: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create account", "INTERNAL_ERROR")
		return
	}

	// Create session
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	ua := r.UserAgent()
	session, refreshToken, err := h.sessionStore.Create(r.Context(), user.ID, ip, ua, h.jwtManager.GetRefreshExpiry())
	if err != nil {
		log.Printf("Error creating session: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create session", "INTERNAL_ERROR")
		return
	}
	_ = session // session stored in DB

	// Generate access token
	accessToken, expiresIn, err := h.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to generate token", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusCreated, &models.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	})
}

// Login authenticates a user with email and password
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Email and password are required", "MISSING_FIELDS")
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Find user
	user, err := h.userStore.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			writeError(w, http.StatusUnauthorized, "Invalid email or password", "INVALID_CREDENTIALS")
			return
		}
		log.Printf("Error finding user: %v", err)
		writeError(w, http.StatusInternalServerError, "Authentication failed", "INTERNAL_ERROR")
		return
	}

	// Verify password
	if err := h.userStore.VerifyPassword(user, req.Password); err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid email or password", "INVALID_CREDENTIALS")
		return
	}

	// Create session
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	ua := r.UserAgent()
	_, refreshToken, err := h.sessionStore.Create(r.Context(), user.ID, ip, ua, h.jwtManager.GetRefreshExpiry())
	if err != nil {
		log.Printf("Error creating session: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create session", "INTERNAL_ERROR")
		return
	}

	// Generate access token
	accessToken, expiresIn, err := h.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to generate token", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, &models.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	})
}

// RefreshToken issues a new access token using a refresh token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "Refresh token is required", "MISSING_TOKEN")
		return
	}

	// Validate refresh token
	session, err := h.sessionStore.ValidateRefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, store.ErrSessionNotFound) || errors.Is(err, store.ErrSessionExpired) {
			writeError(w, http.StatusUnauthorized, "Invalid or expired refresh token", "INVALID_REFRESH_TOKEN")
			return
		}
		log.Printf("Error validating refresh token: %v", err)
		writeError(w, http.StatusInternalServerError, "Token refresh failed", "INTERNAL_ERROR")
		return
	}

	// Get user
	user, err := h.userStore.GetByID(r.Context(), session.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to refresh token", "INTERNAL_ERROR")
		return
	}

	// Generate new access token
	accessToken, expiresIn, err := h.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": accessToken,
		"expires_in":   expiresIn,
	})
}

// Logout invalidates the current session
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated", "UNAUTHORIZED")
		return
	}

	// Delete all sessions for user (full logout)
	if err := h.sessionStore.DeleteAllForUser(r.Context(), userID); err != nil {
		log.Printf("Error deleting sessions: %v", err)
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// GetCurrentUser returns the authenticated user's profile
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated", "UNAUTHORIZED")
		return
	}

	user, err := h.userStore.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "User not found", "USER_NOT_FOUND")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch user", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// UpdateProfile updates the authenticated user's profile
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated", "UNAUTHORIZED")
		return
	}

	var req models.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	user, err := h.userStore.Update(r.Context(), userID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update profile", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// Impersonate allows an admin to generate an access token for a target user
func (h *AuthHandler) Impersonate(w http.ResponseWriter, r *http.Request) {
	// 1. Get current admin user from token
	adminID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated", "UNAUTHORIZED")
		return
	}

	adminUser, err := h.userStore.GetByID(r.Context(), adminID)
	if err != nil || adminUser.Role != "admin" {
		writeError(w, http.StatusForbidden, "Admin access required", "FORBIDDEN")
		return
	}

	// 2. Parse target user ID
	var req struct {
		TargetUserID string `json:"target_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	targetID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid target_user_id format", "INVALID_ID")
		return
	}

	// 3. Fetch target user
	targetUser, err := h.userStore.GetByID(r.Context(), targetID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Target user not found", "USER_NOT_FOUND")
		return
	}

	// 4. Generate access token FOR target user
	accessToken, expiresIn, err := h.jwtManager.GenerateAccessToken(targetUser.ID, targetUser.Email, targetUser.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token", "INTERNAL_ERROR")
		return
	}

	log.Printf("Admin %s impersonified user %s", adminUser.Email, targetUser.Email)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": accessToken,
		"expires_in":   expiresIn,
		"user":         targetUser,
	})
}

// GitHubCallback handles GitHub OAuth callback
func (h *AuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement GitHub OAuth flow
	// 1. Exchange code for access token
	// 2. Fetch user info from GitHub API
	// 3. Create or update user via CreateOAuth
	// 4. Generate JWT tokens
	// 5. Redirect to dashboard with tokens
	writeError(w, http.StatusNotImplemented, "GitHub OAuth not yet implemented", "NOT_IMPLEMENTED")
}

// --- API Key Handler ---

type APIKeyHandler struct {
	apiKeyStore *store.APIKeyStore
	jwtManager  *middleware.JWTManager
}

func NewAPIKeyHandler(apiKeyStore *store.APIKeyStore, jwtManager *middleware.JWTManager) *APIKeyHandler {
	return &APIKeyHandler{
		apiKeyStore: apiKeyStore,
		jwtManager:  jwtManager,
	}
}

func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated", "UNAUTHORIZED")
		return
	}

	keys, err := h.apiKeyStore.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list API keys", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"api_keys": keys})
}

func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated", "UNAUTHORIZED")
		return
	}

	var req models.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Name is required", "MISSING_FIELDS")
		return
	}

	apiKey, fullKey, err := h.apiKeyStore.Create(r.Context(), userID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create API key", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusCreated, &models.CreateAPIKeyResponse{
		APIKey: apiKey,
		Key:    fullKey,
	})
}

func (h *APIKeyHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Not authenticated", "UNAUTHORIZED")
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		// Try chi URL param
		idStr = chi_param(r, "id")
	}

	keyID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid key ID", "INVALID_ID")
		return
	}

	if err := h.apiKeyStore.Delete(r.Context(), keyID, userID); err != nil {
		if errors.Is(err, store.ErrAPIKeyNotFound) {
			writeError(w, http.StatusNotFound, "API key not found", "NOT_FOUND")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to revoke API key", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "API key revoked"})
}

// --- Helpers ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message, code string) {
	writeJSON(w, status, &models.ErrorResponse{
		Error: message,
		Code:  code,
	})
}

func chi_param(r *http.Request, name string) string {
	return chi.URLParam(r, name)
}
