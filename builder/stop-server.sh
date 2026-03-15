#!/bin/bash
# Stop Antisky server node
set -euo pipefail
echo "[ANTISKY] Stopping server node..."
cd "$(dirname "$0")"
docker compose down
echo "[✓] Server node stopped"
