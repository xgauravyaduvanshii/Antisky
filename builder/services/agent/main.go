package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type AgentConfig struct {
	Port              string
	AdminAPIURL       string
	ServerKey         string
	ServerID          string
	RedisURL          string
	ClusterSecret     string
	HeartbeatInterval time.Duration
	MetricsInterval   time.Duration
}

type SystemMetrics struct {
	CPUPercent       float64   `json:"cpu_percent"`
	RAMUsedMB        int       `json:"ram_used_mb"`
	RAMTotalMB       int       `json:"ram_total_mb"`
	DiskUsedGB       float64   `json:"disk_used_gb"`
	DiskTotalGB      float64   `json:"disk_total_gb"`
	NetworkRxBytes   int64     `json:"network_rx_bytes"`
	NetworkTxBytes   int64     `json:"network_tx_bytes"`
	ActiveContainers int       `json:"active_containers"`
	LoadAverage      []float64 `json:"load_average"`
	Uptime           int64     `json:"uptime_seconds"`
}

type Heartbeat struct {
	ServerKey string        `json:"server_key"`
	ServerID  string        `json:"server_id"`
	Status    string        `json:"status"`
	Metrics   SystemMetrics `json:"metrics"`
	Timestamp string        `json:"timestamp"`
}

type CommandRequest struct {
	ID      string `json:"id"`
	Command string `json:"command"`
	Args    []string `json:"args"`
	Timeout int    `json:"timeout"`
}

type CommandResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

var config AgentConfig

func main() {
	log.Println("╔══════════════════════════════════════╗")
	log.Println("║     Antisky Server Agent v1.0        ║")
	log.Println("╚══════════════════════════════════════╝")

	config = AgentConfig{
		Port:              getEnv("AGENT_PORT", "8090"),
		AdminAPIURL:       getEnv("ADMIN_API_URL", "http://localhost:8080"),
		ServerKey:         getEnv("SERVER_KEY", "dev-server-key"),
		ServerID:          getEnv("SERVER_ID", "dev-server"),
		RedisURL:          getEnv("REDIS_URL", ""),
		ClusterSecret:     getEnv("CLUSTER_SECRET", ""),
		HeartbeatInterval: getDurationEnv("HEARTBEAT_INTERVAL", 10) * time.Second,
		MetricsInterval:   getDurationEnv("METRICS_INTERVAL", 10) * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start heartbeat loop
	go heartbeatLoop(ctx)

	// HTTP API
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/metrics", metricsHandler)
	mux.HandleFunc("/exec", authMiddleware(execHandler))
	mux.HandleFunc("/containers", authMiddleware(containersHandler))
	mux.HandleFunc("/containers/", authMiddleware(containerActionHandler))
	mux.HandleFunc("/logs", authMiddleware(logsHandler))
	mux.HandleFunc("/info", infoHandler)

	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: mux,
	}

	go func() {
		log.Printf("🚀 Agent listening on :%s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Agent failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down agent...")
	cancel()
	server.Shutdown(context.Background())
}

// --- Heartbeat ---

func heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sendHeartbeat()
		}
	}
}

func sendHeartbeat() {
	metrics := collectMetrics()
	hb := Heartbeat{
		ServerKey: config.ServerKey,
		ServerID:  config.ServerID,
		Status:    "online",
		Metrics:   metrics,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	body, _ := json.Marshal(hb)
	req, err := http.NewRequest("POST", config.AdminAPIURL+"/api/v1/servers/heartbeat", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Server-Key", config.ServerKey)
	req.Header.Set("X-Cluster-Secret", config.ClusterSecret)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Heartbeat failed: %v", err)
		return
	}
	resp.Body.Close()
}

// --- Metrics Collection ---

func collectMetrics() SystemMetrics {
	m := SystemMetrics{
		LoadAverage: getLoadAverage(),
		Uptime:      getUptime(),
	}

	// CPU
	m.CPUPercent = getCPUPercent()

	// RAM
	m.RAMUsedMB, m.RAMTotalMB = getRAMUsage()

	// Disk
	m.DiskUsedGB, m.DiskTotalGB = getDiskUsage()

	// Network
	m.NetworkRxBytes, m.NetworkTxBytes = getNetworkStats()

	// Docker containers
	m.ActiveContainers = getActiveContainers()

	return m
}

func getCPUPercent() float64 {
	data, err := os.ReadFile("/host/proc/stat")
	if err != nil {
		data, _ = os.ReadFile("/proc/stat")
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0
	}
	fields := strings.Fields(lines[0])
	if len(fields) < 5 {
		return 0
	}
	var total, idle float64
	for i := 1; i < len(fields); i++ {
		v, _ := strconv.ParseFloat(fields[i], 64)
		total += v
		if i == 4 {
			idle = v
		}
	}
	if total == 0 {
		return 0
	}
	return math.Round((1-idle/total)*10000) / 100
}

func getRAMUsage() (int, int) {
	data, err := os.ReadFile("/host/proc/meminfo")
	if err != nil {
		data, _ = os.ReadFile("/proc/meminfo")
	}
	var total, available int
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fmt.Sscanf(line, "MemTotal: %d kB", &total)
		}
		if strings.HasPrefix(line, "MemAvailable:") {
			fmt.Sscanf(line, "MemAvailable: %d kB", &available)
		}
	}
	totalMB := total / 1024
	usedMB := (total - available) / 1024
	return usedMB, totalMB
}

func getDiskUsage() (float64, float64) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err != nil {
		return 0, 0
	}
	totalGB := float64(stat.Blocks*uint64(stat.Bsize)) / (1024 * 1024 * 1024)
	freeGB := float64(stat.Bfree*uint64(stat.Bsize)) / (1024 * 1024 * 1024)
	return math.Round((totalGB-freeGB)*100) / 100, math.Round(totalGB*100) / 100
}

func getNetworkStats() (int64, int64) {
	data, err := os.ReadFile("/host/proc/net/dev")
	if err != nil {
		data, _ = os.ReadFile("/proc/net/dev")
	}
	var rx, tx int64
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, "eth0") || strings.Contains(line, "ens") {
			fields := strings.Fields(line)
			if len(fields) >= 10 {
				r, _ := strconv.ParseInt(fields[1], 10, 64)
				t, _ := strconv.ParseInt(fields[9], 10, 64)
				rx += r
				tx += t
			}
		}
	}
	return rx, tx
}

func getLoadAverage() []float64 {
	data, err := os.ReadFile("/host/proc/loadavg")
	if err != nil {
		data, _ = os.ReadFile("/proc/loadavg")
	}
	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return []float64{0, 0, 0}
	}
	l1, _ := strconv.ParseFloat(fields[0], 64)
	l5, _ := strconv.ParseFloat(fields[1], 64)
	l15, _ := strconv.ParseFloat(fields[2], 64)
	return []float64{l1, l5, l15}
}

func getUptime() int64 {
	data, err := os.ReadFile("/host/proc/uptime")
	if err != nil {
		data, _ = os.ReadFile("/proc/uptime")
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return 0
	}
	v, _ := strconv.ParseFloat(fields[0], 64)
	return int64(v)
}

func getActiveContainers() int {
	out, err := exec.Command("docker", "ps", "-q").Output()
	if err != nil {
		return 0
	}
	lines := strings.TrimSpace(string(out))
	if lines == "" {
		return 0
	}
	return len(strings.Split(lines, "\n"))
}

// --- HTTP Handlers ---

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"service":   "antisky-agent",
		"server_id": config.ServerID,
		"uptime":    getUptime(),
		"go_version": runtime.Version(),
	})
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(collectMetrics())
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"server_id":  config.ServerID,
		"hostname":   hostname,
		"os":         runtime.GOOS,
		"arch":       runtime.GOARCH,
		"cpus":       runtime.NumCPU(),
		"go_version": runtime.Version(),
	})
}

func execHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, 400)
		return
	}

	timeout := time.Duration(req.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Printf("Executing command: %s %v", req.Command, req.Args)

	cmd := exec.CommandContext(ctx, req.Command, req.Args...)
	out, err := cmd.CombinedOutput()

	resp := CommandResponse{
		ID:     req.ID,
		Status: "completed",
		Output: string(out),
	}

	if err != nil {
		resp.Status = "failed"
		resp.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func containersHandler(w http.ResponseWriter, r *http.Request) {
	out, err := exec.Command("docker", "ps", "--format", "{{json .}}").Output()
	if err != nil {
		http.Error(w, `{"error":"docker not available"}`, 500)
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	containers := make([]json.RawMessage, 0)
	for _, line := range lines {
		if line != "" {
			containers = append(containers, json.RawMessage(line))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"containers": containers,
		"count":      len(containers),
	})
}

func containerActionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req struct {
		ContainerID string `json:"container_id"`
		Action      string `json:"action"` // start, stop, restart, remove
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid body"}`, 400)
		return
	}

	allowedActions := map[string]bool{"start": true, "stop": true, "restart": true, "remove": true}
	if !allowedActions[req.Action] {
		http.Error(w, `{"error":"invalid action"}`, 400)
		return
	}

	args := []string{req.Action, req.ContainerID}
	if req.Action == "remove" {
		args = []string{"rm", "-f", req.ContainerID}
	}

	out, err := exec.Command("docker", args...).CombinedOutput()
	resp := map[string]string{"status": "ok", "output": string(out)}
	if err != nil {
		resp["status"] = "error"
		resp["error"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	containerID := r.URL.Query().Get("container")
	lines := r.URL.Query().Get("lines")
	if lines == "" {
		lines = "100"
	}

	if containerID == "" {
		http.Error(w, `{"error":"container param required"}`, 400)
		return
	}

	out, err := exec.Command("docker", "logs", "--tail", lines, containerID).CombinedOutput()
	if err != nil {
		http.Error(w, `{"error":"failed to get logs"}`, 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"container": containerID,
		"logs":      string(out),
	})
}

// --- Auth Middleware ---

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Server-Key")
		secret := r.Header.Get("X-Cluster-Secret")

		if key == "" && secret == "" {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, `{"error":"unauthorized"}`, 401)
				return
			}
		}

		// In production, validate against stored server keys
		next(w, r)
	}
}

// --- Helpers ---

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDurationEnv(key string, fallback int) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := strconv.Atoi(v)
		if err == nil {
			return time.Duration(d)
		}
	}
	return time.Duration(fallback)
}
