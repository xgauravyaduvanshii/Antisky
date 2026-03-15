#!/bin/bash
# ===================================
# Antisky Server Node — Register with Admin Panel
# ===================================
# Usage: ./register-server.sh [ADMIN_API_URL]

set -euo pipefail

CYAN='\033[0;36m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

log() { echo -e "${CYAN}[ANTISKY]${NC} $1"; }
ok()  { echo -e "${GREEN}[✓]${NC} $1"; }
err() { echo -e "${RED}[✗]${NC} $1"; exit 1; }

INSTALL_DIR="/opt/antisky"
ADMIN_API="${1:-${ADMIN_API_URL:-}}"

if [ -z "$ADMIN_API" ]; then
    err "Usage: ./register-server.sh https://admin-api.antisky.app"
fi

# Collect server info
SERVER_KEY=$(cat "$INSTALL_DIR/security/server.key" 2>/dev/null || echo "unknown")
SERVER_ID=$(cat "$INSTALL_DIR/security/server.id" 2>/dev/null || echo "unknown")
HOSTNAME=$(hostname)
IP_ADDRESS=$(hostname -I | awk '{print $1}')
OS_INFO=$(lsb_release -d -s 2>/dev/null || cat /etc/os-release | grep PRETTY_NAME | cut -d= -f2 | tr -d '"')
DOCKER_VERSION=$(docker --version 2>/dev/null | awk '{print $3}' | tr -d ',')
CPU_CORES=$(nproc 2>/dev/null || echo 0)
RAM_MB=$(free -m 2>/dev/null | awk '/Mem:/{print $2}' || echo 0)
DISK_GB=$(df -BG / 2>/dev/null | awk 'NR==2{print $2}' | tr -d 'G' || echo 0)
REGION="${AWS_REGION:-$(curl -sf http://169.254.169.254/latest/meta-data/placement/region 2>/dev/null || echo 'local')}"

log "Registering server with admin panel..."
log "  Admin API:     $ADMIN_API"
log "  Server ID:     $SERVER_ID"
log "  Hostname:      $HOSTNAME"
log "  IP:            $IP_ADDRESS"
log "  CPU:           $CPU_CORES cores"
log "  RAM:           ${RAM_MB}MB"
log "  Disk:          ${DISK_GB}GB"
log "  Region:        $REGION"

# Register via API
RESPONSE=$(curl -sf -X POST "$ADMIN_API/api/v1/servers/register" \
    -H "Content-Type: application/json" \
    -H "X-Cluster-Secret: ${CLUSTER_SECRET:-antisky-cluster-secret-2026}" \
    -d '{
        "server_key": "'"$SERVER_KEY"'",
        "server_id": "'"$SERVER_ID"'",
        "hostname": "'"$HOSTNAME"'",
        "ip_address": "'"$IP_ADDRESS"'",
        "port": 8090,
        "region": "'"$REGION"'",
        "os_info": "'"$OS_INFO"'",
        "docker_version": "'"$DOCKER_VERSION"'",
        "cpu_cores": '"$CPU_CORES"',
        "ram_mb": '"$RAM_MB"',
        "disk_gb": '"$DISK_GB"'
    }' 2>&1) || err "Failed to register with admin panel. Is $ADMIN_API reachable?"

echo "$RESPONSE" | jq . 2>/dev/null || echo "$RESPONSE"

# Save server token from response
SERVER_TOKEN=$(echo "$RESPONSE" | jq -r '.auth_token // empty' 2>/dev/null)
if [ -n "$SERVER_TOKEN" ]; then
    echo "$SERVER_TOKEN" > "$INSTALL_DIR/security/server.token"
    chmod 600 "$INSTALL_DIR/security/server.token"
    ok "Server token saved"
fi

ok "Server registered successfully!"
