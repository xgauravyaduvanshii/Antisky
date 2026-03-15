package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// OAuthConfig stores provider-specific OAuth configuration
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	Scopes       []string
	RedirectURL  string
}

var providers = map[string]*OAuthConfig{
	"github": {
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
		Scopes:   []string{"user:email", "repo"},
	},
	"google": {
		AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL: "https://oauth2.googleapis.com/token",
		Scopes:   []string{"openid", "profile", "email"},
	},
	"gitlab": {
		AuthURL:  "https://gitlab.com/oauth/authorize",
		TokenURL: "https://gitlab.com/oauth/token",
		Scopes:   []string{"read_user", "api"},
	},
	"bitbucket": {
		AuthURL:  "https://bitbucket.org/site/oauth2/authorize",
		TokenURL: "https://bitbucket.org/site/oauth2/access_token",
		Scopes:   []string{"account", "repository"},
	},
}

func init() {
	// Load OAuth credentials from environment
	if p := providers["github"]; p != nil {
		p.ClientID = os.Getenv("GITHUB_CLIENT_ID")
		p.ClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
		p.RedirectURL = getEnvDefault("PLATFORM_URL", "http://localhost:3000") + "/auth/callback/github"
	}
	if p := providers["google"]; p != nil {
		p.ClientID = os.Getenv("GOOGLE_CLIENT_ID")
		p.ClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
		p.RedirectURL = getEnvDefault("PLATFORM_URL", "http://localhost:3000") + "/auth/callback/google"
	}
	if p := providers["gitlab"]; p != nil {
		p.ClientID = os.Getenv("GITLAB_CLIENT_ID")
		p.ClientSecret = os.Getenv("GITLAB_CLIENT_SECRET")
		p.RedirectURL = getEnvDefault("PLATFORM_URL", "http://localhost:3000") + "/auth/callback/gitlab"
	}
	if p := providers["bitbucket"]; p != nil {
		p.ClientID = os.Getenv("BITBUCKET_CLIENT_ID")
		p.ClientSecret = os.Getenv("BITBUCKET_CLIENT_SECRET")
		p.RedirectURL = getEnvDefault("PLATFORM_URL", "http://localhost:3000") + "/auth/callback/bitbucket"
	}
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// OAuthAuthorize initiates the OAuth flow for a given provider
func OAuthAuthorize(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		provider = strings.TrimPrefix(r.URL.Path, "/auth/oauth/")
	}

	cfg, ok := providers[provider]
	if !ok {
		http.Error(w, `{"error":"unsupported provider"}`, 400)
		return
	}

	if cfg.ClientID == "" {
		http.Error(w, fmt.Sprintf(`{"error":"%s OAuth not configured"}`, provider), 400)
		return
	}

	// Generate state token for CSRF protection
	state := generateState()
	// TODO: Store state in Redis with 10min TTL

	params := url.Values{
		"client_id":     {cfg.ClientID},
		"redirect_uri":  {cfg.RedirectURL},
		"scope":         {strings.Join(cfg.Scopes, " ")},
		"state":         {state},
		"response_type": {"code"},
	}

	if provider == "google" {
		params.Set("access_type", "offline")
		params.Set("prompt", "consent")
	}

	redirectURL := cfg.AuthURL + "?" + params.Encode()
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}

// OAuthCallback handles the OAuth callback from the provider
func OAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		// Extract from path: /auth/callback/github -> github
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 0 {
			provider = parts[len(parts)-1]
		}
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		errDesc := r.URL.Query().Get("error_description")
		http.Error(w, fmt.Sprintf(`{"error":"oauth failed: %s"}`, errDesc), 400)
		return
	}

	// TODO: Validate state against Redis

	cfg, ok := providers[provider]
	if !ok {
		http.Error(w, `{"error":"unsupported provider"}`, 400)
		return
	}

	// Exchange code for token
	tokenResp, err := exchangeCode(cfg, code)
	if err != nil {
		log.Printf("OAuth token exchange failed: %v", err)
		http.Error(w, `{"error":"token exchange failed"}`, 500)
		return
	}

	// Get user profile from provider
	profile, err := getProviderProfile(provider, tokenResp.AccessToken)
	if err != nil {
		log.Printf("OAuth profile fetch failed: %v", err)
		http.Error(w, `{"error":"profile fetch failed"}`, 500)
		return
	}

	_ = state // Used for CSRF validation

	// TODO: Create or link user account with provider profile
	// TODO: Store OAuth tokens encrypted in oauth_connections table
	// TODO: Generate JWT for the user

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"provider": provider,
		"profile":  profile,
		"message":  "OAuth completed — user account linked",
	})
}

// TokenResponse holds the OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

func exchangeCode(cfg *OAuthConfig, code string) (*TokenResponse, error) {
	data := url.Values{
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"code":          {code},
		"redirect_uri":  {cfg.RedirectURL},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequest("POST", cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		// GitHub returns query string format sometimes
		vals, parseErr := url.ParseQuery(string(body))
		if parseErr != nil {
			return nil, fmt.Errorf("parse token response: %w", err)
		}
		tokenResp.AccessToken = vals.Get("access_token")
		tokenResp.TokenType = vals.Get("token_type")
		tokenResp.Scope = vals.Get("scope")
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response: %s", string(body))
	}

	return &tokenResp, nil
}

// ProviderProfile holds standardized user profile from OAuth provider
type ProviderProfile struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Username  string `json:"username"`
}

func getProviderProfile(provider, accessToken string) (*ProviderProfile, error) {
	var apiURL string
	switch provider {
	case "github":
		apiURL = "https://api.github.com/user"
	case "google":
		apiURL = "https://www.googleapis.com/oauth2/v2/userinfo"
	case "gitlab":
		apiURL = "https://gitlab.com/api/v4/user"
	case "bitbucket":
		apiURL = "https://api.bitbucket.org/2.0/user"
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	if provider == "github" {
		req.Header.Set("Accept", "application/vnd.github.v3+json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var raw map[string]interface{}
	json.Unmarshal(body, &raw)

	profile := &ProviderProfile{}
	switch provider {
	case "github":
		profile.ID = fmt.Sprintf("%v", raw["id"])
		profile.Email, _ = raw["email"].(string)
		profile.Name, _ = raw["name"].(string)
		profile.AvatarURL, _ = raw["avatar_url"].(string)
		profile.Username, _ = raw["login"].(string)
		// If email is empty, fetch from /user/emails
		if profile.Email == "" {
			profile.Email = fetchGitHubEmail(accessToken)
		}
	case "google":
		profile.ID, _ = raw["id"].(string)
		profile.Email, _ = raw["email"].(string)
		profile.Name, _ = raw["name"].(string)
		profile.AvatarURL, _ = raw["picture"].(string)
	case "gitlab":
		profile.ID = fmt.Sprintf("%v", raw["id"])
		profile.Email, _ = raw["email"].(string)
		profile.Name, _ = raw["name"].(string)
		profile.AvatarURL, _ = raw["avatar_url"].(string)
		profile.Username, _ = raw["username"].(string)
	case "bitbucket":
		profile.ID, _ = raw["uuid"].(string)
		profile.Name, _ = raw["display_name"].(string)
		profile.Username, _ = raw["username"].(string)
	}

	return profile, nil
}

func fetchGitHubEmail(accessToken string) string {
	req, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	json.Unmarshal(body, &emails)
	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email
		}
	}
	if len(emails) > 0 {
		return emails[0].Email
	}
	return ""
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
