#!/bin/bash
# ===================================
# Antisky Server Node — Dependency Installer
# ===================================
# Run on a fresh Ubuntu 22.04+ server (e.g. AWS EC2)
# Usage: sudo bash install.sh

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

log() { echo -e "${CYAN}[ANTISKY]${NC} $1"; }
ok()  { echo -e "${GREEN}[✓]${NC} $1"; }
err() { echo -e "${RED}[✗]${NC} $1"; exit 1; }

echo -e "${BOLD}${BLUE}"
echo "╔══════════════════════════════════════════╗"
echo "║     Antisky Server Node Installer        ║"
echo "║           v1.0.0                         ║"
echo "╚══════════════════════════════════════════╝"
echo -e "${NC}"

# Check root
if [ "$EUID" -ne 0 ]; then
    err "Please run as root: sudo bash install.sh"
fi

INSTALL_DIR="/opt/antisky"
mkdir -p "$INSTALL_DIR"

# --- System Updates ---
log "Updating system packages..."
apt-get update -qq
apt-get upgrade -y -qq
ok "System packages updated"

# --- Docker ---
log "Installing Docker..."
if ! command -v docker &>/dev/null; then
    curl -fsSL https://get.docker.com | sh
    systemctl enable docker
    systemctl start docker
    usermod -aG docker ubuntu 2>/dev/null || true
    ok "Docker installed"
else
    ok "Docker already installed ($(docker --version))"
fi

# --- Docker Compose ---
log "Installing Docker Compose..."
if ! command -v docker-compose &>/dev/null && ! docker compose version &>/dev/null; then
    apt-get install -y -qq docker-compose-plugin
    ok "Docker Compose installed"
else
    ok "Docker Compose already installed"
fi

# --- Node.js 20 ---
log "Installing Node.js 20..."
if ! command -v node &>/dev/null; then
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
    apt-get install -y -qq nodejs
    ok "Node.js installed ($(node --version))"
else
    ok "Node.js already installed ($(node --version))"
fi

# --- Go 1.22 ---
log "Installing Go 1.22..."
if ! command -v go &>/dev/null; then
    wget -q https://go.dev/dl/go1.22.10.linux-amd64.tar.gz -O /tmp/go.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf /tmp/go.tar.gz
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile.d/go.sh
    export PATH=$PATH:/usr/local/go/bin
    rm /tmp/go.tar.gz
    ok "Go installed ($(go version))"
else
    ok "Go already installed ($(go version))"
fi

# --- Monitoring Tools ---
log "Installing monitoring tools..."
apt-get install -y -qq htop iotop sysstat jq curl wget net-tools
ok "Monitoring tools installed"

# --- Firewall ---
log "Configuring firewall..."
if command -v ufw &>/dev/null; then
    ufw allow 22/tcp    # SSH
    ufw allow 80/tcp    # HTTP
    ufw allow 443/tcp   # HTTPS
    ufw allow 8090/tcp  # Agent API
    ufw allow 8091/tcp  # Terminal proxy
    ufw --force enable 2>/dev/null || true
    ok "Firewall configured"
fi

# --- Create Antisky directories ---
log "Creating Antisky directories..."
mkdir -p "$INSTALL_DIR"/{data,logs,config,security,tmp}
chown -R ubuntu:ubuntu "$INSTALL_DIR" 2>/dev/null || true
ok "Directories created at $INSTALL_DIR"

# --- Generate server identity ---
log "Generating server identity..."
if [ ! -f "$INSTALL_DIR/security/server.key" ]; then
    openssl rand -hex 32 > "$INSTALL_DIR/security/server.key"
    openssl rand -hex 16 > "$INSTALL_DIR/security/server.id"
    chmod 600 "$INSTALL_DIR/security/server.key"
    chmod 600 "$INSTALL_DIR/security/server.id"
    ok "Server identity generated"
else
    ok "Server identity already exists"
fi

echo ""
echo -e "${BOLD}${GREEN}Installation complete!${NC}"
echo -e "Run ${CYAN}./start-server.sh${NC} to start the Antisky server node."
echo ""
