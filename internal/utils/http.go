package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func WriteJSONError(w http.ResponseWriter, status int, message string, log *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := map[string]string{"error": message}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("Failed to encode response error", slog.Any("error", err))
	}
}

func WriteJSONResponse(w http.ResponseWriter, status int, body interface{}, log *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(body)
	if err != nil {
		log.Error("Failed to encode response body", slog.Any("error", err))
	}
}
