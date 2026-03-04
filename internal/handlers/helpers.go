package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// write utility
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeAPIError(w http.ResponseWriter, httpCode int, code, message string, details any) {
	resp := map[string]interface{}{
		"error_code": code,
		"message":    message,
	}
	if details != nil {
		resp["details"] = details
	}
	writeJSON(w, httpCode, resp)
}

func toOpenAPIUUID(u uuid.UUID) *openapi_types.UUID {
	x := openapi_types.UUID(u)
	return &x
}
