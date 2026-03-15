package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/antisky/services/control-plane/internal/models"
)

// writeJSON encodes v as JSON and writes to response
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes an error response
func writeError(w http.ResponseWriter, status int, message, code string) {
	writeJSON(w, status, &models.ErrorResponse{
		Error: message,
		Code:  code,
	})
}
