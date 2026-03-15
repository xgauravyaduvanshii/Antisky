#!/bin/bash
# ===================================
# Antisky Server Node — Bootstrap & Start
# ===================================
# Usage: ./start-server.sh [ADMIN_API_URL]

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

log() { echo -e "${CYAN}[ANTISKY]${NC} $1"; }
ok()  { echo -e "${GREEN}[✓]${NC} $1"; }
warn(){ echo -e "${YELLOW}[!]${NC} $1"; }
err() { echo -e "${RED}[✗]${NC} $1"; exit 1; }

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INSTALL_DIR="/opt/antisky"
ENV_FILE="$SCRIPT_DIR/.env"
ADMIN_API="${1:-}"

echo -e "${BOLD}${BLUE}"
echo "╔══════════════════════════════════════════╗"
echo "║     Antisky Server Node — Starting       ║"
echo "╚══════════════════════════════════════════╝"
echo -e "${NC}"

# --- Step 1: Check dependencies ---
log "Checking dependencies..."
command -v docker &>/dev/null || err "Docker not found. Run install.sh first."
command -v curl &>/dev/null   || err "curl not found. Run install.sh first."
ok "Dependencies verified"

# --- Step 2: Load or create env ---
if [ ! -f "$ENV_FILE" ]; then
    log "Creating environment from template..."
    cp "$SCRIPT_DIR/.env.template" "$ENV_FILE"
fi
source "$ENV_FILE"

# --- Step 3: Server identity ---
SERVER_KEY=$(cat "$INSTALL_DIR/security/server.key" 2>/dev/null || openssl rand -hex 32)
SERVER_ID=$(cat "$INSTALL_DIR/security/server.id" 2>/dev/null || openssl rand -hex 16)
HOSTNAME=$(hostname)
IP_ADDRESS=$(hostname -I | awk '{print $1}')

log "Server ID:  $SERVER_ID"
log "Hostname:   $HOSTNAME"
log "IP Address: $IP_ADDRESS"

# --- Step 4: Update env with server info ---
sed -i "s|SERVER_KEY=.*|SERVER_KEY=$SERVER_KEY|g" "$ENV_FILE" 2>/dev/null || true
sed -i "s|SERVER_ID=.*|SERVER_ID=$SERVER_ID|g" "$ENV_FILE" 2>/dev/null || true
sed -i "s|SERVER_HOSTNAME=.*|SERVER_HOSTNAME=$HOSTNAME|g" "$ENV_FILE" 2>/dev/null || true
sed -i "s|SERVER_IP=.*|SERVER_IP=$IP_ADDRESS|g" "$ENV_FILE" 2>/dev/null || true

# --- Step 5: Start services via Docker Compose ---
log "Starting Antisky server services..."
cd "$SCRIPT_DIR"
docker compose -f docker-compose.yml up -d --build

ok "Services started"

# --- Step 6: Wait for agent to be ready ---
log "Waiting for agent health check..."
for i in $(seq 1 30); do
    if curl -sf http://localhost:8090/health >/dev/null 2>&1; then
        ok "Agent is healthy"
        break
    fi
    sleep 2
    if [ "$i" -eq 30 ]; then
        warn "Agent health check timeout — services may still be starting"
    fi
done

# --- Step 7: Register with admin panel ---
if [ -n "$ADMIN_API" ]; then
    log "Registering with admin panel at $ADMIN_API..."
    bash "$SCRIPT_DIR/register-server.sh" "$ADMIN_API"
elif [ -n "${ADMIN_API_URL:-}" ]; then
    log "Registering with admin panel at $ADMIN_API_URL..."
    bash "$SCRIPT_DIR/register-server.sh" "$ADMIN_API_URL"
else
    warn "No ADMIN_API_URL set. Server running in standalone mode."
    warn "To register: ./register-server.sh https://admin.antisky.app"
fi

# --- Done ---
echo ""
echo -e "${BOLD}${GREEN}═══════════════════════════════════════════${NC}"
echo -e "${BOLD}${GREEN}  Antisky Server Node is running!          ${NC}"
echo -e "${BOLD}${GREEN}═══════════════════════════════════════════${NC}"
echo ""
echo -e "  Agent API:    ${CYAN}http://$IP_ADDRESS:8090${NC}"
echo -e "  Terminal:     ${CYAN}http://$IP_ADDRESS:8091${NC}"
echo -e "  Server ID:    ${CYAN}$SERVER_ID${NC}"
echo ""
echo -e "  Logs:         ${CYAN}docker compose logs -f${NC}"
echo -e "  Stop:         ${CYAN}./stop-server.sh${NC}"
echo ""
