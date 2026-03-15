package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"github.com/antisky/services/server-manager/internal/store"
)

type ServerHandler struct {
	store         *store.ServerStore
	rdb           *redis.Client
	clusterSecret string
}

func NewServerHandler(store *store.ServerStore, rdb *redis.Client, clusterSecret string) *ServerHandler {
	return &ServerHandler{store: store, rdb: rdb, clusterSecret: clusterSecret}
}

// --- Server Registration ---

type RegisterRequest struct {
	ServerKey     string  `json:"server_key"`
	ServerID      string  `json:"server_id"`
	Hostname      string  `json:"hostname"`
	IPAddress     string  `json:"ip_address"`
	Port          int     `json:"port"`
	Region        string  `json:"region"`
	OSInfo        *string `json:"os_info"`
	DockerVersion *string `json:"docker_version"`
	CPUCores      *int    `json:"cpu_cores"`
	RAMMB         *int    `json:"ram_mb"`
	DiskGB        *int    `json:"disk_gb"`
}

func (h *ServerHandler) RegisterServer(w http.ResponseWriter, r *http.Request) {
	// Verify cluster secret
	secret := r.Header.Get("X-Cluster-Secret")
	if secret != h.clusterSecret {
		jsonError(w, "invalid cluster secret", 401)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", 400)
		return
	}

	if req.ServerKey == "" || req.Hostname == "" || req.IPAddress == "" {
		jsonError(w, "server_key, hostname, and ip_address are required", 400)
		return
	}

	server, authToken, err := h.store.RegisterServer(r.Context(),
		req.Hostname, req.IPAddress, req.Port, req.Region, req.ServerKey,
		req.OSInfo, req.DockerVersion, req.CPUCores, req.RAMMB, req.DiskGB,
	)
	if err != nil {
		jsonError(w, "registration failed: "+err.Error(), 500)
		return
	}

	jsonResponse(w, 201, map[string]interface{}{
		"server":     server,
		"auth_token": authToken,
		"message":    "Server registered successfully",
	})
}

// --- Heartbeat ---

type HeartbeatRequest struct {
	ServerKey string              `json:"server_key"`
	ServerID  string              `json:"server_id"`
	Status    string              `json:"status"`
	Metrics   *store.ServerMetrics `json:"metrics"`
	Timestamp string              `json:"timestamp"`
}

func (h *ServerHandler) ReceiveHeartbeat(w http.ResponseWriter, r *http.Request) {
	var req HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid body", 400)
		return
	}

	err := h.store.UpdateHeartbeat(r.Context(), req.ServerKey, req.Metrics)
	if err != nil {
		jsonError(w, "heartbeat failed: "+err.Error(), 500)
		return
	}

	// Cache latest metrics in Redis for fast real-time dashboard access
	if h.rdb != nil && req.Metrics != nil {
		metricsJSON, _ := json.Marshal(req.Metrics)
		h.rdb.Set(r.Context(), "server:metrics:"+req.ServerKey, metricsJSON, 0)
	}

	jsonResponse(w, 200, map[string]string{"status": "ok"})
}

// --- Server Management ---

func (h *ServerHandler) ListServers(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	servers, err := h.store.ListServers(r.Context(), status)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResponse(w, 200, map[string]interface{}{
		"servers": servers,
		"count":   len(servers),
	})
}

func (h *ServerHandler) GetServer(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")
	server, err := h.store.GetServer(r.Context(), serverID)
	if err != nil {
		jsonError(w, "server not found", 404)
		return
	}
	jsonResponse(w, 200, server)
}

func (h *ServerHandler) DecommissionServer(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")
	if err := h.store.DecommissionServer(r.Context(), serverID); err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "decommissioned"})
}

func (h *ServerHandler) GetServerMetrics(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if l, err := strconv.Atoi(limitStr); err == nil {
		limit = l
	}

	metrics, err := h.store.GetServerMetrics(r.Context(), serverID, limit)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResponse(w, 200, map[string]interface{}{
		"metrics": metrics,
		"count":   len(metrics),
	})
}

// --- Remote Command ---

type CommandBody struct {
	Command string            `json:"command"`
	Args    map[string]interface{} `json:"args"`
}

func (h *ServerHandler) SendCommand(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")

	var body CommandBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonError(w, "invalid body", 400)
		return
	}

	// TODO: Extract admin user ID from JWT
	adminID := "00000000-0000-0000-0000-000000000000"

	cmd, err := h.store.CreateCommand(r.Context(), serverID, adminID, body.Command, body.Args)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	// Forward command to server agent
	server, err := h.store.GetServer(r.Context(), serverID)
	if err == nil && server.Status == "online" {
		go forwardCommandToAgent(server, cmd)
	}

	jsonResponse(w, 201, cmd)
}

func (h *ServerHandler) GetServerCommands(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")
	cmds, err := h.store.GetServerCommands(r.Context(), serverID, 50)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResponse(w, 200, map[string]interface{}{"commands": cmds, "count": len(cmds)})
}

func forwardCommandToAgent(server *store.Server, cmd *store.ServerCommand) {
	url := fmt.Sprintf("http://%s:%d/exec", server.IPAddress, server.Port)
	body, _ := json.Marshal(map[string]interface{}{
		"id":      cmd.ID,
		"command": cmd.Command,
		"timeout": 30,
	})
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)
}

// --- Admin Infrastructure Actions ---

func (h *ServerHandler) FlushCache(w http.ResponseWriter, r *http.Request) {
	if h.rdb == nil {
		jsonError(w, "redis is not connected", 503)
		return
	}
	if err := h.rdb.FlushAll(r.Context()).Err(); err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResponse(w, 200, map[string]string{"message": "Cache flushed successfully"})
}

func (h *ServerHandler) DrainServer(w http.ResponseWriter, r *http.Request) {
	serverID := chi.URLParam(r, "serverID")
	if serverID == "" {
		jsonError(w, "missing server ID", 400)
		return
	}
	
	err := h.store.DecommissionServer(r.Context(), serverID)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResponse(w, 200, map[string]string{"message": "Server draining initiated"})
}

// --- User Management ---

func (h *ServerHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit := 50
	offset := 0
	if l, err := strconv.Atoi(limitStr); err == nil {
		limit = l
	}
	if o, err := strconv.Atoi(offsetStr); err == nil {
		offset = o
	}

	users, total, err := h.store.ListUsers(r.Context(), search, limit, offset)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResponse(w, 200, map[string]interface{}{
		"users": users,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}

func (h *ServerHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Delegate to existing user store — simplified here
	userID := chi.URLParam(r, "userID")
	jsonResponse(w, 200, map[string]string{"user_id": userID, "message": "user details"})
}

func (h *ServerHandler) BanUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	var body struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&body)
	if err := h.store.BanUser(r.Context(), userID, body.Reason); err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "banned"})
}

func (h *ServerHandler) UnbanUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if err := h.store.UnbanUser(r.Context(), userID); err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "unbanned"})
}

func (h *ServerHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if err := h.store.DeleteUser(r.Context(), userID); err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "deleted"})
}

func (h *ServerHandler) ImpersonateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	var body struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	// TODO: Extract admin user ID from JWT
	adminID := "00000000-0000-0000-0000-000000000000"

	logID, err := h.store.CreateImpersonationLog(r.Context(), adminID, userID, body.Reason, r.RemoteAddr)
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}

	// Generate impersonation JWT
	// TODO: Generate actual JWT for target user with admin flag
	jsonResponse(w, 200, map[string]interface{}{
		"impersonation_id": logID,
		"target_user_id":   userID,
		"message":          "Impersonation session created",
	})
}

func (h *ServerHandler) SetUserRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	var body struct {
		Role string `json:"role"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	validRoles := map[string]bool{"user": true, "admin": true, "super_admin": true}
	if !validRoles[body.Role] {
		jsonError(w, "invalid role", 400)
		return
	}

	if err := h.store.SetUserRole(r.Context(), userID, body.Role); err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "role updated"})
}

// --- Platform Stats ---

func (h *ServerHandler) GetPlatformStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.store.GetPlatformStats(r.Context())
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResponse(w, 200, stats)
}

func (h *ServerHandler) GetClusterConfig(w http.ResponseWriter, r *http.Request) {
	config, err := h.store.GetClusterConfig(r.Context())
	if err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResponse(w, 200, config)
}

func (h *ServerHandler) UpdateClusterConfig(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	var body struct {
		Value interface{} `json:"value"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	adminID := "00000000-0000-0000-0000-000000000000" // TODO: from JWT
	if err := h.store.UpdateClusterConfig(r.Context(), key, body.Value, adminID); err != nil {
		jsonError(w, err.Error(), 500)
		return
	}
	jsonResponse(w, 200, map[string]string{"status": "updated"})
}

// --- Helpers ---

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
