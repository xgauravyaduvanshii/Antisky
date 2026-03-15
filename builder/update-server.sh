#!/bin/bash
# Antisky Server Node — Auto Update
set -euo pipefail
CYAN='\033[0;36m'
GREEN='\033[0;32m'
NC='\033[0m'
log() { echo -e "${CYAN}[ANTISKY]${NC} $1"; }
ok()  { echo -e "${GREEN}[✓]${NC} $1"; }

INSTALL_DIR="/opt/antisky"
ADMIN_API="${ADMIN_API_URL:-http://localhost:8080}"
SERVER_TOKEN=$(cat "$INSTALL_DIR/security/server.token" 2>/dev/null || echo "")

log "Checking for updates..."
UPDATE_INFO=$(curl -sf "$ADMIN_API/api/v1/servers/updates" \
    -H "Authorization: Bearer $SERVER_TOKEN" 2>/dev/null || echo '{"update_available": false}')

UPDATE_AVAILABLE=$(echo "$UPDATE_INFO" | jq -r '.update_available' 2>/dev/null)
if [ "$UPDATE_AVAILABLE" = "true" ]; then
    DOWNLOAD_URL=$(echo "$UPDATE_INFO" | jq -r '.download_url')
    log "Update available! Downloading..."
    curl -sfL "$DOWNLOAD_URL" -o /tmp/antisky-update.tar.gz
    cd "$(dirname "$0")"
    docker compose down
    tar -xzf /tmp/antisky-update.tar.gz -C "$(dirname "$0")" --strip-components=1
    docker compose up -d --build
    rm /tmp/antisky-update.tar.gz
    ok "Update applied and services restarted"
else
    ok "No updates available"
fi
