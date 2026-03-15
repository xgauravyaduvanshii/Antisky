package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/antisky/services/control-plane/internal/middleware"
	"github.com/antisky/services/control-plane/internal/models"
	"github.com/antisky/services/control-plane/internal/store"
)

type OrgHandler struct {
	orgStore *store.OrgStore
}

func NewOrgHandler(orgStore *store.OrgStore) *OrgHandler {
	return &OrgHandler{orgStore: orgStore}
}

func (h *OrgHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	uid, _ := uuid.Parse(userID)

	orgs, err := h.orgStore.ListByUser(r.Context(), uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list organizations", "INTERNAL_ERROR")
		return
	}
	if orgs == nil {
		orgs = []*models.Organization{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"organizations": orgs})
}

func (h *OrgHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	uid, _ := uuid.Parse(userID)

	var req models.CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}
	if req.Name == "" || req.Slug == "" {
		writeError(w, http.StatusBadRequest, "Name and slug are required", "MISSING_FIELDS")
		return
	}

	org, err := h.orgStore.Create(r.Context(), &req, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create organization", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusCreated, org)
}

func (h *OrgHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid org ID", "INVALID_ID")
		return
	}

	org, err := h.orgStore.GetByID(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
		return
	}

	writeJSON(w, http.StatusOK, org)
}

func (h *OrgHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid org ID", "INVALID_ID")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	org, err := h.orgStore.Update(r.Context(), orgID, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update organization", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, org)
}

func (h *OrgHandler) Delete(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid org ID", "INVALID_ID")
		return
	}

	if err := h.orgStore.Delete(r.Context(), orgID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete organization", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Organization deleted"})
}

func (h *OrgHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid org ID", "INVALID_ID")
		return
	}

	members, err := h.orgStore.ListMembers(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list members", "INTERNAL_ERROR")
		return
	}
	if members == nil {
		members = []*models.OrgMember{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"members": members})
}

func (h *OrgHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid org ID", "INVALID_ID")
		return
	}

	inviterID := middleware.GetUserID(r.Context())
	inviterUID, _ := uuid.Parse(inviterID)

	var req models.InviteMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_BODY")
		return
	}

	// TODO: Look up user by email, create invite if not found
	// For now, assume user ID is passed
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"message":    "Member invited",
		"org_id":     orgID,
		"invited_by": inviterUID,
		"role":       req.Role,
	})
}

func (h *OrgHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(chi.URLParam(r, "orgId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid org ID", "INVALID_ID")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid user ID", "INVALID_ID")
		return
	}

	if err := h.orgStore.RemoveMember(r.Context(), orgID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to remove member", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Member removed"})
}
