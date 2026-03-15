package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/antisky/services/control-plane/internal/middleware"
	"github.com/antisky/services/control-plane/internal/models"
	"github.com/antisky/services/control-plane/internal/store"
)

type ProjectHandler struct {
	projectStore *store.ProjectStore
	orgStore     *store.OrgStore
}

func NewProjectHandler(projectStore *store.ProjectStore, orgStore *store.OrgStore) *ProjectHandler {
	return &ProjectHandler{
		projectStore: projectStore,
		orgStore:     orgStore,
	}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	uid, err := uuid.Parse(userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid user", "UNAUTHORIZED")
		return
	}

	projects, err := h.projectStore.ListByUser(r.Context(), uid)
	if err != nil {
		log.Printf("Error listing projects: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to list projects", "INTERNAL_ERROR")
		return
	}

	if projects == nil {
		projects = []*models.Project{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"projects": projects,
		"count":    len(projects),
	})
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	uid, _ := uuid.Parse(userID)

	var req models.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "Project name is required", "MISSING_FIELDS")
		return
	}

	// If no org specified, create a personal org or use first
	if req.OrgID == uuid.Nil {
		orgs, err := h.orgStore.ListByUser(r.Context(), uid)
		if err != nil || len(orgs) == 0 {
			// Create personal org
			org, err := h.orgStore.Create(r.Context(), &models.CreateOrgRequest{
				Name: "Personal",
				Slug: "personal-" + userID[:8],
			}, uid)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to create organization", "INTERNAL_ERROR")
				return
			}
			req.OrgID = org.ID
		} else {
			req.OrgID = orgs[0].ID
		}
	}

	project, err := h.projectStore.Create(r.Context(), &req, uid)
	if err != nil {
		if errors.Is(err, store.ErrProjectSlugExists) {
			writeError(w, http.StatusConflict, "A project with this name already exists", "SLUG_EXISTS")
			return
		}
		log.Printf("Error creating project: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create project", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid project ID", "INVALID_ID")
		return
	}

	project, err := h.projectStore.GetByID(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			writeError(w, http.StatusNotFound, "Project not found", "NOT_FOUND")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch project", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid project ID", "INVALID_ID")
		return
	}

	var req models.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	project, err := h.projectStore.Update(r.Context(), projectID, &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update project", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid project ID", "INVALID_ID")
		return
	}

	if err := h.projectStore.Delete(r.Context(), projectID); err != nil {
		if errors.Is(err, store.ErrProjectNotFound) {
			writeError(w, http.StatusNotFound, "Project not found", "NOT_FOUND")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to delete project", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Project deleted"})
}

// --- Domain handlers ---

func (h *ProjectHandler) ListDomains(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid project ID", "INVALID_ID")
		return
	}

	domains, err := h.projectStore.ListDomains(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list domains", "INTERNAL_ERROR")
		return
	}
	if domains == nil {
		domains = []*models.ProjectDomain{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"domains": domains})
}

func (h *ProjectHandler) AddDomain(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid project ID", "INVALID_ID")
		return
	}

	var req models.AddDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	if req.Domain == "" {
		writeError(w, http.StatusBadRequest, "Domain is required", "MISSING_FIELDS")
		return
	}

	domain, err := h.projectStore.AddDomain(r.Context(), projectID, req.Domain)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to add domain", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusCreated, domain)
}

func (h *ProjectHandler) RemoveDomain(w http.ResponseWriter, r *http.Request) {
	domainID, err := uuid.Parse(chi.URLParam(r, "domainId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid domain ID", "INVALID_ID")
		return
	}

	if err := h.projectStore.RemoveDomain(r.Context(), domainID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to remove domain", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Domain removed"})
}
