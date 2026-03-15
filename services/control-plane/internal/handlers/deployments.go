package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/antisky/services/control-plane/internal/middleware"
	"github.com/antisky/services/control-plane/internal/models"
	"github.com/antisky/services/control-plane/internal/store"
)

type DeploymentHandler struct {
	deployStore  *store.DeploymentStore
	projectStore *store.ProjectStore
	rdb          *redis.Client
}

func NewDeploymentHandler(deployStore *store.DeploymentStore, projectStore *store.ProjectStore, rdb *redis.Client) *DeploymentHandler {
	return &DeploymentHandler{
		deployStore:  deployStore,
		projectStore: projectStore,
		rdb:          rdb,
	}
}

// TriggerDeploy starts a new deployment
func (h *DeploymentHandler) TriggerDeploy(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid project ID", "INVALID_ID")
		return
	}

	userID := middleware.GetUserID(r.Context())
	uid, _ := uuid.Parse(userID)

	var req models.TriggerDeployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	// Verify project exists
	project, err := h.projectStore.GetByID(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			writeError(w, http.StatusNotFound, "Project not found", "NOT_FOUND")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch project", "INTERNAL_ERROR")
		return
	}

	if req.Ref == "" {
		req.Ref = project.RepoBranch
	}
	if req.Source == "" {
		req.Source = "dashboard"
	}

	deployType := "production"
	if req.Ref != project.RepoBranch {
		deployType = "preview"
	}

	// Create deployment record
	deploy, err := h.deployStore.Create(r.Context(), projectID, &uid, req.Ref, "", "", "", deployType, req.Source)
	if err != nil {
		log.Printf("Error creating deployment: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create deployment", "INTERNAL_ERROR")
		return
	}

	// Publish build event to Redis for the build orchestrator to pick up
	buildEvent := map[string]interface{}{
		"deployment_id": deploy.ID.String(),
		"project_id":    project.ID.String(),
		"repo_url":      project.RepoURL,
		"ref":           req.Ref,
		"runtime":       project.Runtime,
		"framework":     project.Framework,
		"build_command": project.BuildCommand,
		"start_command": project.StartCommand,
		"output_dir":    project.OutputDir,
		"root_dir":      project.RootDir,
	}
	eventJSON, _ := json.Marshal(buildEvent)
	h.rdb.Publish(r.Context(), "builds:queue", string(eventJSON))

	// Generate deployment URLs
	previewURL := project.Slug + "--" + deploy.ID.String()[:8] + ".antisky.app"
	h.deployStore.SetURL(r.Context(), deploy.ID, "", previewURL)
	deploy.PreviewURL = &previewURL

	log.Printf("Deployment %s triggered for project %s (ref: %s)", deploy.ID, project.Name, req.Ref)

	writeJSON(w, http.StatusCreated, deploy)
}

func (h *DeploymentHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid project ID", "INVALID_ID")
		return
	}

	deployments, err := h.deployStore.ListByProject(r.Context(), projectID, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list deployments", "INTERNAL_ERROR")
		return
	}

	if deployments == nil {
		deployments = []*models.Deployment{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"deployments": deployments,
		"count":       len(deployments),
	})
}

func (h *DeploymentHandler) Get(w http.ResponseWriter, r *http.Request) {
	deployID, err := uuid.Parse(chi.URLParam(r, "deployId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid deployment ID", "INVALID_ID")
		return
	}

	deploy, err := h.deployStore.GetByID(r.Context(), deployID)
	if err != nil {
		if errors.Is(err, store.ErrDeploymentNotFound) {
			writeError(w, http.StatusNotFound, "Deployment not found", "NOT_FOUND")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch deployment", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, deploy)
}

func (h *DeploymentHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	deployID, err := uuid.Parse(chi.URLParam(r, "deployId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid deployment ID", "INVALID_ID")
		return
	}

	if err := h.deployStore.UpdateStatus(r.Context(), deployID, "cancelled", nil); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to cancel deployment", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Deployment cancelled"})
}

func (h *DeploymentHandler) Rollback(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid project ID", "INVALID_ID")
		return
	}

	deployID, err := uuid.Parse(chi.URLParam(r, "deployId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid deployment ID", "INVALID_ID")
		return
	}

	// Get the target deployment to rollback to
	targetDeploy, err := h.deployStore.GetByID(r.Context(), deployID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Target deployment not found", "NOT_FOUND")
		return
	}

	userID := middleware.GetUserID(r.Context())
	uid, _ := uuid.Parse(userID)

	// Create a new rollback deployment
	ref := ""
	if targetDeploy.Ref != nil {
		ref = *targetDeploy.Ref
	}
	commitSHA := ""
	if targetDeploy.CommitSHA != nil {
		commitSHA = *targetDeploy.CommitSHA
	}

	rollbackDeploy, err := h.deployStore.Create(r.Context(), projectID, &uid, ref, commitSHA, "Rollback to "+deployID.String()[:8], "", "rollback", "dashboard")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create rollback", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusCreated, rollbackDeploy)
}

func (h *DeploymentHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	deployID, err := uuid.Parse(chi.URLParam(r, "deployId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid deployment ID", "INVALID_ID")
		return
	}

	// For now, return logs from Redis pub/sub channel
	// In production, this would be a WebSocket endpoint streaming from CloudWatch
	logKey := "deploy:logs:" + deployID.String()
	logs, err := h.rdb.LRange(r.Context(), logKey, 0, -1).Result()
	if err != nil {
		logs = []string{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"deployment_id": deployID,
		"logs":          logs,
	})
}
