#!/bin/bash
# Key generation for server identity
set -euo pipefail
INSTALL_DIR="/opt/antisky"
mkdir -p "$INSTALL_DIR/security"
openssl rand -hex 32 > "$INSTALL_DIR/security/server.key"
openssl rand -hex 16 > "$INSTALL_DIR/security/server.id"
chmod 600 "$INSTALL_DIR/security/server.key" "$INSTALL_DIR/security/server.id"
echo "[✓] Server keys generated at $INSTALL_DIR/security/"
