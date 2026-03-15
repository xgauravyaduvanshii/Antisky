package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Hostname        string    `json:"hostname"`
	IPAddress       string    `json:"ip_address"`
	Port            int       `json:"port"`
	Region          string    `json:"region"`
	Zone            *string   `json:"zone,omitempty"`
	ServerType      string    `json:"server_type"`
	Status          string    `json:"status"`
	ServerKey       string    `json:"server_key"`
	OSInfo          *string   `json:"os_info,omitempty"`
	DockerVersion   *string   `json:"docker_version,omitempty"`
	CPUCores        *int      `json:"cpu_cores,omitempty"`
	RAMMB           *int      `json:"ram_mb,omitempty"`
	DiskGB          *int      `json:"disk_gb,omitempty"`
	Labels          string    `json:"labels"`
	LastHeartbeatAt *time.Time `json:"last_heartbeat_at,omitempty"`
	RegisteredAt    time.Time `json:"registered_at"`
	CreatedAt       time.Time `json:"created_at"`
}

type ServerMetrics struct {
	ID               string    `json:"id"`
	ServerID         string    `json:"server_id"`
	CPUPercent       float64   `json:"cpu_percent"`
	RAMUsedMB        int       `json:"ram_used_mb"`
	RAMTotalMB       int       `json:"ram_total_mb"`
	DiskUsedGB       float64   `json:"disk_used_gb"`
	DiskTotalGB      float64   `json:"disk_total_gb"`
	NetworkRxBytes   int64     `json:"network_rx_bytes"`
	NetworkTxBytes   int64     `json:"network_tx_bytes"`
	ActiveContainers int       `json:"active_containers"`
	RecordedAt       time.Time `json:"recorded_at"`
}

type ServerCommand struct {
	ID          string    `json:"id"`
	ServerID    string    `json:"server_id"`
	IssuedBy    string    `json:"issued_by"`
	Command     string    `json:"command"`
	Status      string    `json:"status"`
	Output      *string   `json:"output,omitempty"`
	Error       *string   `json:"error,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type ServerStore struct {
	pool *pgxpool.Pool
}

func NewServerStore(pool *pgxpool.Pool) *ServerStore {
	return &ServerStore{pool: pool}
}

func (s *ServerStore) RegisterServer(ctx context.Context, hostname, ipAddress string, port int, region string, serverKey string, osInfo, dockerVersion *string, cpuCores, ramMB, diskGB *int) (*Server, string, error) {
	id := uuid.New().String()

	// Generate auth token
	authToken := uuid.New().String() + "-" + uuid.New().String()
	tokenHash, err := bcrypt.GenerateFromPassword([]byte(authToken), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("hash token: %w", err)
	}

	name := hostname
	if name == "" {
		name = "server-" + id[:8]
	}

	server := &Server{}
	err = s.pool.QueryRow(ctx,
		`INSERT INTO servers (id, name, hostname, ip_address, port, region, server_key, auth_token_hash, os_info, docker_version, cpu_cores, ram_mb, disk_gb, status, last_heartbeat_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, 'online', NOW())
		 ON CONFLICT (server_key) DO UPDATE SET
			hostname = EXCLUDED.hostname, ip_address = EXCLUDED.ip_address, port = EXCLUDED.port,
			os_info = EXCLUDED.os_info, docker_version = EXCLUDED.docker_version,
			cpu_cores = EXCLUDED.cpu_cores, ram_mb = EXCLUDED.ram_mb, disk_gb = EXCLUDED.disk_gb,
			status = 'online', last_heartbeat_at = NOW(), updated_at = NOW()
		 RETURNING id, name, hostname, ip_address, port, region, server_type, status, server_key, registered_at, created_at`,
		id, name, hostname, ipAddress, port, region, serverKey, string(tokenHash), osInfo, dockerVersion, cpuCores, ramMB, diskGB,
	).Scan(&server.ID, &server.Name, &server.Hostname, &server.IPAddress, &server.Port, &server.Region, &server.ServerType, &server.Status, &server.ServerKey, &server.RegisteredAt, &server.CreatedAt)
	if err != nil {
		return nil, "", fmt.Errorf("register server: %w", err)
	}

	return server, authToken, nil
}

func (s *ServerStore) UpdateHeartbeat(ctx context.Context, serverKey string, metrics *ServerMetrics) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update server heartbeat
	_, err = tx.Exec(ctx,
		`UPDATE servers SET status = 'online', last_heartbeat_at = NOW(), updated_at = NOW() WHERE server_key = $1`,
		serverKey,
	)
	if err != nil {
		return err
	}

	// Get server ID
	var serverID string
	err = tx.QueryRow(ctx, `SELECT id FROM servers WHERE server_key = $1`, serverKey).Scan(&serverID)
	if err != nil {
		return err
	}

	// Store metrics
	if metrics != nil {
		_, err = tx.Exec(ctx,
			`INSERT INTO server_metrics (server_id, cpu_percent, ram_used_mb, ram_total_mb, disk_used_gb, disk_total_gb, network_rx_bytes, network_tx_bytes, active_containers)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			serverID, metrics.CPUPercent, metrics.RAMUsedMB, metrics.RAMTotalMB, metrics.DiskUsedGB, metrics.DiskTotalGB, metrics.NetworkRxBytes, metrics.NetworkTxBytes, metrics.ActiveContainers,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *ServerStore) ListServers(ctx context.Context, statusFilter string) ([]Server, error) {
	query := `SELECT id, name, hostname, ip_address, port, region, server_type, status, server_key, os_info, docker_version, cpu_cores, ram_mb, disk_gb, last_heartbeat_at, registered_at, created_at
		       FROM servers`
	args := []interface{}{}

	if statusFilter != "" && statusFilter != "all" {
		query += ` WHERE status = $1`
		args = append(args, statusFilter)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []Server
	for rows.Next() {
		var sv Server
		err := rows.Scan(&sv.ID, &sv.Name, &sv.Hostname, &sv.IPAddress, &sv.Port, &sv.Region, &sv.ServerType, &sv.Status, &sv.ServerKey, &sv.OSInfo, &sv.DockerVersion, &sv.CPUCores, &sv.RAMMB, &sv.DiskGB, &sv.LastHeartbeatAt, &sv.RegisteredAt, &sv.CreatedAt)
		if err != nil {
			return nil, err
		}
		servers = append(servers, sv)
	}
	return servers, nil
}

func (s *ServerStore) GetServer(ctx context.Context, serverID string) (*Server, error) {
	sv := &Server{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, hostname, ip_address, port, region, server_type, status, server_key, os_info, docker_version, cpu_cores, ram_mb, disk_gb, last_heartbeat_at, registered_at, created_at
		 FROM servers WHERE id = $1`, serverID,
	).Scan(&sv.ID, &sv.Name, &sv.Hostname, &sv.IPAddress, &sv.Port, &sv.Region, &sv.ServerType, &sv.Status, &sv.ServerKey, &sv.OSInfo, &sv.DockerVersion, &sv.CPUCores, &sv.RAMMB, &sv.DiskGB, &sv.LastHeartbeatAt, &sv.RegisteredAt, &sv.CreatedAt)
	if err != nil {
		return nil, err
	}
	return sv, nil
}

func (s *ServerStore) DecommissionServer(ctx context.Context, serverID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE servers SET status = 'decommissioned', updated_at = NOW() WHERE id = $1`, serverID)
	return err
}

func (s *ServerStore) GetServerMetrics(ctx context.Context, serverID string, limit int) ([]ServerMetrics, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, server_id, cpu_percent, ram_used_mb, ram_total_mb, disk_used_gb, disk_total_gb, network_rx_bytes, network_tx_bytes, active_containers, recorded_at
		 FROM server_metrics WHERE server_id = $1 ORDER BY recorded_at DESC LIMIT $2`, serverID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []ServerMetrics
	for rows.Next() {
		var m ServerMetrics
		if err := rows.Scan(&m.ID, &m.ServerID, &m.CPUPercent, &m.RAMUsedMB, &m.RAMTotalMB, &m.DiskUsedGB, &m.DiskTotalGB, &m.NetworkRxBytes, &m.NetworkTxBytes, &m.ActiveContainers, &m.RecordedAt); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, nil
}

func (s *ServerStore) CreateCommand(ctx context.Context, serverID, issuedBy, command string, args map[string]interface{}) (*ServerCommand, error) {
	id := uuid.New().String()
	cmd := &ServerCommand{}
	argsJSON := "{}"
	if args != nil {
		// Simple marshal
		parts := []string{}
		for k, v := range args {
			parts = append(parts, fmt.Sprintf(`"%s":"%v"`, k, v))
		}
		argsJSON = "{" + strings.Join(parts, ",") + "}"
	}

	err := s.pool.QueryRow(ctx,
		`INSERT INTO server_commands (id, server_id, issued_by, command, args) VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, server_id, issued_by, command, status, created_at`,
		id, serverID, issuedBy, command, argsJSON,
	).Scan(&cmd.ID, &cmd.ServerID, &cmd.IssuedBy, &cmd.Command, &cmd.Status, &cmd.CreatedAt)
	return cmd, err
}

func (s *ServerStore) GetServerCommands(ctx context.Context, serverID string, limit int) ([]ServerCommand, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, server_id, issued_by, command, status, output, error, created_at, completed_at
		 FROM server_commands WHERE server_id = $1 ORDER BY created_at DESC LIMIT $2`, serverID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cmds []ServerCommand
	for rows.Next() {
		var c ServerCommand
		if err := rows.Scan(&c.ID, &c.ServerID, &c.IssuedBy, &c.Command, &c.Status, &c.Output, &c.Error, &c.CreatedAt, &c.CompletedAt); err != nil {
			return nil, err
		}
		cmds = append(cmds, c)
	}
	return cmds, nil
}

func (s *ServerStore) MarkStaleServers(ctx context.Context, timeout time.Duration) (int, error) {
	result, err := s.pool.Exec(ctx,
		`UPDATE servers SET status = 'offline', updated_at = NOW()
		 WHERE status = 'online' AND last_heartbeat_at < $1`,
		time.Now().Add(-timeout),
	)
	if err != nil {
		return 0, err
	}
	return int(result.RowsAffected()), nil
}

// --- User management ---

type UserInfo struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	IsBanned    bool      `json:"is_banned"`
	LoginCount  int       `json:"login_count"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *ServerStore) ListUsers(ctx context.Context, search string, limit, offset int) ([]UserInfo, int, error) {
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT id, email, name, COALESCE(role,'user'), COALESCE(is_banned,false), COALESCE(login_count,0), last_login_at, created_at FROM users`
	countQuery := `SELECT COUNT(*) FROM users`
	args := []interface{}{}
	n := 1

	if search != "" {
		where := fmt.Sprintf(` WHERE email ILIKE $%d OR name ILIKE $%d`, n, n)
		query += where
		countQuery += where
		args = append(args, "%"+search+"%")
		n++
	}

	var total int
	s.pool.QueryRow(ctx, countQuery, args...).Scan(&total)

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, n, n+1)
	args = append(args, limit, offset)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []UserInfo
	for rows.Next() {
		var u UserInfo
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.IsBanned, &u.LoginCount, &u.LastLoginAt, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (s *ServerStore) BanUser(ctx context.Context, userID, reason string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET is_banned = true, banned_reason = $2, updated_at = NOW() WHERE id = $1`,
		userID, reason,
	)
	return err
}

func (s *ServerStore) UnbanUser(ctx context.Context, userID string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET is_banned = false, banned_reason = NULL, updated_at = NOW() WHERE id = $1`,
		userID,
	)
	return err
}

func (s *ServerStore) SetUserRole(ctx context.Context, userID, role string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE users SET role = $2, updated_at = NOW() WHERE id = $1`,
		userID, role,
	)
	return err
}

func (s *ServerStore) CreateImpersonationLog(ctx context.Context, adminID, targetUserID, reason, ipAddress string) (string, error) {
	id := uuid.New().String()
	_, err := s.pool.Exec(ctx,
		`INSERT INTO admin_impersonation_logs (id, admin_id, target_user_id, reason, ip_address) VALUES ($1, $2, $3, $4, $5)`,
		id, adminID, targetUserID, reason, ipAddress,
	)
	return id, err
}

// --- Platform Stats ---

type PlatformStats struct {
	TotalUsers       int `json:"total_users"`
	TotalServers     int `json:"total_servers"`
	OnlineServers    int `json:"online_servers"`
	TotalProjects    int `json:"total_projects"`
	TotalDeployments int `json:"total_deployments"`
	ActiveBuilds     int `json:"active_builds"`
}

func (s *ServerStore) GetPlatformStats(ctx context.Context) (*PlatformStats, error) {
	stats := &PlatformStats{}
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM servers WHERE status != 'decommissioned'`).Scan(&stats.TotalServers)
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM servers WHERE status = 'online'`).Scan(&stats.OnlineServers)
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM projects`).Scan(&stats.TotalProjects)
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM deployments`).Scan(&stats.TotalDeployments)
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM deployments WHERE status = 'building'`).Scan(&stats.ActiveBuilds)
	return stats, nil
}

func (s *ServerStore) GetClusterConfig(ctx context.Context) (map[string]interface{}, error) {
	rows, err := s.pool.Query(ctx, `SELECT key, value, description FROM cluster_config ORDER BY key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	config := make(map[string]interface{})
	for rows.Next() {
		var key, desc string
		var value interface{}
		rows.Scan(&key, &value, &desc)
		config[key] = map[string]interface{}{"value": value, "description": desc}
	}
	return config, nil
}

func (s *ServerStore) UpdateClusterConfig(ctx context.Context, key string, value interface{}, updatedBy string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE cluster_config SET value = $2, updated_by = $3, updated_at = NOW() WHERE key = $1`,
		key, value, updatedBy,
	)
	return err
}
