#!/bin/bash
# Health check script
AGENT_HEALTH=$(curl -sf http://localhost:8090/health 2>/dev/null)
TERMINAL_HEALTH=$(curl -sf http://localhost:8091/health 2>/dev/null)

if echo "$AGENT_HEALTH" | grep -q "healthy"; then
    echo "[✓] Agent healthy"
else
    echo "[✗] Agent unhealthy"
fi

if echo "$TERMINAL_HEALTH" | grep -q "healthy"; then
    echo "[✓] Terminal proxy healthy"
else
    echo "[✗] Terminal proxy unhealthy"
fi
