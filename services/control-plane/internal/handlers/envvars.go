package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/antisky/services/control-plane/internal/models"
	"github.com/antisky/services/control-plane/internal/store"
)

type EnvVarHandler struct {
	envVarStore  *store.EnvVarStore
	projectStore *store.ProjectStore
}

func NewEnvVarHandler(envVarStore *store.EnvVarStore, projectStore *store.ProjectStore) *EnvVarHandler {
	return &EnvVarHandler{
		envVarStore:  envVarStore,
		projectStore: projectStore,
	}
}

func (h *EnvVarHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid project ID", "INVALID_ID")
		return
	}

	envVars, err := h.envVarStore.List(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list environment variables", "INTERNAL_ERROR")
		return
	}

	if envVars == nil {
		envVars = []*models.EnvVar{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"env_vars": envVars})
}

func (h *EnvVarHandler) Set(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid project ID", "INVALID_ID")
		return
	}

	var req models.SetEnvVarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.Key == "" || req.Value == "" {
		writeError(w, http.StatusBadRequest, "Key and value are required", "MISSING_FIELDS")
		return
	}

	envVar, err := h.envVarStore.Set(r.Context(), projectID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to set environment variable", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusCreated, envVar)
}

func (h *EnvVarHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid project ID", "INVALID_ID")
		return
	}

	key := chi.URLParam(r, "key")
	if key == "" {
		writeError(w, http.StatusBadRequest, "Key is required", "MISSING_FIELDS")
		return
	}

	if err := h.envVarStore.Delete(r.Context(), projectID, key); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete environment variable", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Environment variable deleted"})
}
