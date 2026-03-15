package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/antisky/services/control-plane/internal/store"
)

type WebhookHandler struct {
	deployStore  *store.DeploymentStore
	projectStore *store.ProjectStore
}

func NewWebhookHandler(deployStore *store.DeploymentStore, projectStore *store.ProjectStore) *WebhookHandler {
	return &WebhookHandler{
		deployStore:  deployStore,
		projectStore: projectStore,
	}
}

// GitHubWebhook handles incoming GitHub webhook events
func (h *WebhookHandler) GitHubWebhook(w http.ResponseWriter, r *http.Request) {
	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read body", "INVALID_BODY")
		return
	}

	// Verify webhook signature
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if secret != "" {
		signature := r.Header.Get("X-Hub-Signature-256")
		if !verifyGitHubSignature(body, signature, secret) {
			writeError(w, http.StatusUnauthorized, "Invalid signature", "INVALID_SIGNATURE")
			return
		}
	}

	// Parse event type
	event := r.Header.Get("X-GitHub-Event")
	log.Printf("GitHub webhook received: event=%s", event)

	switch event {
	case "push":
		h.handleGitHubPush(w, r, body)
	case "pull_request":
		h.handleGitHubPR(w, r, body)
	case "ping":
		writeJSON(w, http.StatusOK, map[string]string{"message": "pong"})
	default:
		writeJSON(w, http.StatusOK, map[string]string{"message": "event ignored"})
	}
}

func (h *WebhookHandler) handleGitHubPush(w http.ResponseWriter, r *http.Request, body []byte) {
	var payload struct {
		Ref        string `json:"ref"`
		After      string `json:"after"`
		Repository struct {
			ID       int    `json:"id"`
			FullName string `json:"full_name"`
			CloneURL string `json:"clone_url"`
		} `json:"repository"`
		HeadCommit struct {
			ID      string `json:"id"`
			Message string `json:"message"`
			Author  struct {
				Name string `json:"name"`
			} `json:"author"`
		} `json:"head_commit"`
		Pusher struct {
			Name string `json:"name"`
		} `json:"pusher"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payload", "INVALID_PAYLOAD")
		return
	}

	// Extract branch name from ref (refs/heads/main -> main)
	ref := payload.Ref
	if strings.HasPrefix(ref, "refs/heads/") {
		ref = strings.TrimPrefix(ref, "refs/heads/")
	}

	// Find project by repo
	repoID := itoa(payload.Repository.ID)
	project, err := h.projectStore.GetByRepoID(r.Context(), "github", repoID)
	if err != nil {
		log.Printf("No project found for GitHub repo %s (ID: %s): %v", payload.Repository.FullName, repoID, err)
		writeJSON(w, http.StatusOK, map[string]string{"message": "No project linked to this repo"})
		return
	}

	// Check if auto-deploy is enabled
	if !project.AutoDeploy {
		writeJSON(w, http.StatusOK, map[string]string{"message": "Auto-deploy disabled for this project"})
		return
	}

	// Determine deploy type
	deployType := "preview"
	if ref == project.RepoBranch {
		deployType = "production"
	}

	// Create deployment
	deploy, err := h.deployStore.Create(
		r.Context(),
		project.ID,
		nil, // triggered by webhook, no user
		ref,
		payload.HeadCommit.ID,
		payload.HeadCommit.Message,
		payload.Pusher.Name,
		deployType,
		"webhook",
	)
	if err != nil {
		log.Printf("Failed to create deployment from webhook: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create deployment", "INTERNAL_ERROR")
		return
	}

	log.Printf("Deployment %s created from GitHub push (project: %s, ref: %s)", deploy.ID, project.Name, ref)
	writeJSON(w, http.StatusCreated, deploy)
}

func (h *WebhookHandler) handleGitHubPR(w http.ResponseWriter, r *http.Request, body []byte) {
	var payload struct {
		Action      string `json:"action"`
		PullRequest struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
			Head   struct {
				Ref string `json:"ref"`
				SHA string `json:"sha"`
			} `json:"head"`
		} `json:"pull_request"`
		Repository struct {
			ID       int    `json:"id"`
			FullName string `json:"full_name"`
		} `json:"repository"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid payload", "INVALID_PAYLOAD")
		return
	}

	// Only create preview on opened/synchronize
	if payload.Action != "opened" && payload.Action != "synchronize" {
		writeJSON(w, http.StatusOK, map[string]string{"message": "PR event ignored"})
		return
	}

	// Find project
	repoID := itoa(payload.Repository.ID)
	project, err := h.projectStore.GetByRepoID(r.Context(), "github", repoID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"message": "No project linked"})
		return
	}

	// Create preview deployment
	deploy, err := h.deployStore.Create(
		r.Context(),
		project.ID,
		nil,
		payload.PullRequest.Head.Ref,
		payload.PullRequest.Head.SHA,
		payload.PullRequest.Title,
		"",
		"preview",
		"webhook",
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create preview", "INTERNAL_ERROR")
		return
	}

	log.Printf("Preview deployment %s created for PR #%d (project: %s)", deploy.ID, payload.PullRequest.Number, project.Name)
	writeJSON(w, http.StatusCreated, deploy)
}

// GitLabWebhook handles incoming GitLab webhook events
func (h *WebhookHandler) GitLabWebhook(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement GitLab webhook handling
	writeJSON(w, http.StatusOK, map[string]string{"message": "GitLab webhooks coming soon"})
}

// verifyGitHubSignature verifies the HMAC-SHA256 signature from GitHub
func verifyGitHubSignature(payload []byte, signature, secret string) bool {
	if signature == "" {
		return false
	}

	sig := strings.TrimPrefix(signature, "sha256=")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(sig), []byte(expectedMAC))
}

// Helper to convert int to string (avoiding strconv import)
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}
